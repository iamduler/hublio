package main

import (
	"context"
	"fmt"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	eventsapp "hublio/internal/events/application"
	eventsdomain "hublio/internal/events/domain"
	eventsinfra "hublio/internal/events/infrastructure"
	identityinfra "hublio/internal/identity/infrastructure"
	"hublio/internal/integration/connectors"
	"hublio/internal/integration/connectors/fake"
	"hublio/internal/integration/connectors/misa"
	"hublio/internal/integration/connectors/nhanh"
	integrationinfra "hublio/internal/integration/infrastructure"
	orchestrationapp "hublio/internal/orchestration/application"
	orchestrationdomain "hublio/internal/orchestration/domain"
	orchestrationinfra "hublio/internal/orchestration/infrastructure"
	"hublio/internal/platform/config"
	"hublio/internal/platform/env"
	"hublio/internal/platform/logging"
	"hublio/internal/platform/metrics"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/queue"
	transformationapp "hublio/internal/transformation/application"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	rootDir := env.MustGetWorkingDir()
	logFilePath := filepath.Join(rootDir, "logs", "worker.log")

	logging.InitLogger(logging.LoggerConfig{
		LogLevel:   "info",
		Filename:   logFilePath,
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     30,
		Compress:   true,
		LocalTime:  true,
		IsDev:      env.GetEnv("DEVELOPMENT_MODE", "development"),
	})

	if err := godotenv.Load(filepath.Join(rootDir, ".env")); err != nil {
		logging.Log.Warn().Err(err).Msg("failed to load environment variables")
	} else {
		logging.Log.Info().Msg("environment variables loaded for worker")
	}

	cfg := config.NewConfig()
	redisClient := config.NewRedisClient()

	db, err := persistence.NewDatabase(cfg)
	if err != nil {
		logging.Log.Fatal().Err(err).Msg("failed to connect to database")
	}

	queueLogger := logging.NewLoggerWithPath("queue.log", "info")
	workQueue := queue.NewRedisQueue(redisClient, queueLogger)

	orchestrationSvc, err := newOrchestrationServices(db.Pool, workQueue)
	if err != nil {
		logging.Log.Fatal().Err(err).Msg("failed to initialize orchestration services")
	}

	metricsCounters := metrics.New()
	eventsSvc := newEventsServices(db.Pool, metricsCounters)
	orchestrationSvc.Events = eventsinfra.NewOrchestrationEventBridge(eventsSvc)
	orchestrationSvc.Audit = eventsinfra.NewOrchestrationAuditBridge(eventsSvc)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	logging.Log.Info().Msg("worker started; consuming platform work queue")

	handler := func(ctx context.Context, job queue.Job) error {
		switch job.Type {
		case queue.TypeHealth:
			logging.Log.Info().
				Str("job_id", job.ID).
				Str("job_type", job.Type).
				Msg("platform.health job processed")
			return nil
		case queue.TypeExecution:
			return handleExecutionJob(ctx, db.Pool, orchestrationSvc, job)
		default:
			logging.Log.Warn().
				Str("job_id", job.ID).
				Str("job_type", job.Type).
				Msg("unknown job type ignored")
			return nil
		}
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- workQueue.Consume(ctx, handler)
	}()

	select {
	case <-ctx.Done():
		logging.Log.Info().Msg("shutting down worker")
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			logging.Log.Error().Err(err).Msg("worker consume stopped")
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db.Close()
	_ = redisClient.Close()
	_ = shutdownCtx

	logging.Log.Info().Msg("worker shutdown successfully")
}

// newOrchestrationServices is the worker's own composition root for the Orchestration
// Application layer (mirrors internal/platform/server's wiring; the worker binary has no
// HTTP surface so it does not depend on the server package).
func newOrchestrationServices(pool *pgxpool.Pool, workQueue queue.Queue) (*orchestrationapp.Services, error) {
	key := env.GetEnv("CREDENTIAL_ENCRYPTION_KEY", "")
	if key == "" {
		key = "dev-only-insecure-32-byte-key!!!"
	}
	encryptor, err := integrationinfra.NewAESSecretEncryptor([]byte(key))
	if err != nil {
		return nil, err
	}

	runtimeRegistry := connectors.NewRegistry(fake.New(), misa.New(), nhanh.New())

	connectionGateway := orchestrationinfra.NewConnectionGateway(
		integrationinfra.NewConnectionRepository(pool),
		integrationinfra.NewConnectorRepository(pool),
		integrationinfra.NewCredentialRepository(pool),
		identityinfra.NewWorkspaceRepository(pool),
		identityinfra.NewOrganizationRepository(pool),
		encryptor,
	)
	connectorGateway := orchestrationinfra.NewConnectorGateway(runtimeRegistry)
	transformer := orchestrationinfra.NewTransformerAdapter(transformationapp.NewServices())

	return &orchestrationapp.Services{
		Intents:     orchestrationinfra.NewIntentRepository(pool),
		Executions:  orchestrationinfra.NewExecutionRepository(pool),
		Idempotency: orchestrationinfra.NewIdempotencyRepository(pool),
		Connections: connectionGateway,
		Connectors:  connectorGateway,
		Transformer: transformer,
		Jobs:        orchestrationinfra.NewQueueJobEnqueuer(workQueue),
	}, nil
}

// newEventsServices mirrors internal/platform/server's Events BC wiring (see that package's
// doc comment): the worker binary has no HTTP surface, so it composes its own Events
// Services rather than depending on internal/platform/server.
func newEventsServices(pool *pgxpool.Pool, counters *metrics.Counters) *eventsapp.Services {
	eventRepo := eventsinfra.NewEventRepository(pool)
	svc := &eventsapp.Services{
		Events:  eventRepo,
		Reader:  eventRepo,
		Audit:   eventsinfra.NewAuditRepository(pool),
		Metrics: counters,
		OnSubscriberError: func(event *eventsdomain.PlatformEvent, err error) {
			logging.Log.Warn().Err(err).
				Str("event_name", event.EventName()).
				Str("event_id", event.ID().String()).
				Msg("events: in-process subscriber failed")
		},
	}
	svc.Subscribe(orchestrationdomain.EventExecutionSucceeded, func(_ context.Context, _ *eventsdomain.PlatformEvent) error {
		counters.IncExecutionsSucceeded()
		return nil
	})
	svc.Subscribe(orchestrationdomain.EventExecutionFailed, func(_ context.Context, _ *eventsdomain.PlatformEvent) error {
		counters.IncExecutionsFailed()
		return nil
	})
	return svc
}

// handleExecutionJob runs RunExecution inside a single transaction (the use case's
// transaction boundary) and publishes domain events only after a successful commit.
// Requeue jobs are enqueued only after commit to avoid racing uncommitted rows.
func handleExecutionJob(ctx context.Context, pool *pgxpool.Pool, svc *orchestrationapp.Services, job queue.Job) error {
	raw, _ := job.Payload["execution_id"].(string)
	executionID, err := uuid.Parse(raw)
	if err != nil {
		return fmt.Errorf("worker: invalid execution_id in job payload: %w", err)
	}

	var result *orchestrationapp.RunExecutionResult
	err = persistence.WithinTransaction(ctx, pool, func(ctx context.Context) error {
		var innerErr error
		result, innerErr = svc.RunExecution(ctx, executionID)
		return innerErr
	})
	if err != nil {
		return err
	}
	organizationID, _ := uuid.Parse(stringField(job.Payload, "organization_id"))
	workspaceID, _ := uuid.Parse(stringField(job.Payload, "workspace_id"))
	correlationID := stringField(job.Payload, "correlation_id")
	events := orchestrationapp.EnrichEvents(result.Execution.PullEvents(), organizationID, workspaceID, correlationID)
	svc.PublishAfterCommit(ctx, events...)
	if result.RequeueJob != nil && svc.Jobs != nil {
		if err := svc.Jobs.EnqueueExecution(ctx, *result.RequeueJob); err != nil {
			return fmt.Errorf("worker: failed to re-enqueue execution after commit: %w", err)
		}
	}
	return nil
}

func stringField(payload map[string]any, key string) string {
	s, _ := payload[key].(string)
	return s
}
