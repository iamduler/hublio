package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	eventsapp "hublio/internal/events/application"
	eventsdomain "hublio/internal/events/domain"
	eventsinfra "hublio/internal/events/infrastructure"
	eventshttp "hublio/internal/events/interfaces"
	identityapp "hublio/internal/identity/application"
	identitydomain "hublio/internal/identity/domain"
	identityinfra "hublio/internal/identity/infrastructure"
	identityhttp "hublio/internal/identity/interfaces"
	integrationapp "hublio/internal/integration/application"
	"hublio/internal/integration/connectors"
	"hublio/internal/integration/connectors/fake"
	"hublio/internal/integration/connectors/misa"
	"hublio/internal/integration/connectors/nhanh"
	integrationinfra "hublio/internal/integration/infrastructure"
	integrationhttp "hublio/internal/integration/interfaces"
	orchestrationapp "hublio/internal/orchestration/application"
	orchestrationdomain "hublio/internal/orchestration/domain"
	orchestrationinfra "hublio/internal/orchestration/infrastructure"
	orchestrationhttp "hublio/internal/orchestration/interfaces"
	"hublio/internal/platform/apikey"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/auth"
	"hublio/internal/platform/cache"
	"hublio/internal/platform/config"
	"hublio/internal/platform/docsui"
	"hublio/internal/platform/env"
	"hublio/internal/platform/logging"
	"hublio/internal/platform/metrics"
	"hublio/internal/platform/middleware"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/queue"
	"hublio/internal/platform/requestctx"
	"hublio/internal/platform/validation"
	transformationapp "hublio/internal/transformation/application"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Application struct {
	config        *config.Config
	router        *gin.Engine
	db            *persistence.Database
	redis         *redis.Client
	tokens        auth.TokenService
	cacheSvc      cache.RedisCacheService
	workQueue     queue.Queue
	apiKeyAuth    apikey.Authenticator
	identity      *identityapp.Services
	integration   *integrationapp.Services
	orchestration *orchestrationapp.Services
	events        *eventsapp.Services
	metrics       *metrics.Counters
}

func NewApplication(cfg *config.Config) (*Application, error) {
	if err := validation.InitValidator(); err != nil {
		return nil, err
	}

	db, err := persistence.NewDatabase(cfg)
	if err != nil {
		return nil, err
	}

	redisClient := config.NewRedisClient()
	cacheSvc := cache.NewRedisCacheService(redisClient)
	tokenSvc := auth.NewJWTService(cacheSvc)

	queueLogger := logging.NewLoggerWithPath("queue.log", "info")
	workQueue := queue.NewRedisQueue(redisClient, queueLogger)

	orgRepo := identityinfra.NewOrganizationRepository(db.Pool)
	wsRepo := identityinfra.NewWorkspaceRepository(db.Pool)
	userRepo := identityinfra.NewUserRepository(db.Pool)
	memRepo := identityinfra.NewMembershipRepository(db.Pool)
	keyRepo := identityinfra.NewAPIKeyRepository(db.Pool)

	identitySvc := &identityapp.Services{
		Orgs:        orgRepo,
		Workspaces:  wsRepo,
		Users:       userRepo,
		Memberships: memRepo,
		APIKeys:     keyRepo,
		Passwords:   identityinfra.NewBcryptPasswordHasher(),
	}

	dbAuth := identityinfra.NewDBAuthenticator(keyRepo, wsRepo, orgRepo)
	apiKeyAuth := newAPIKeyAuthenticator(dbAuth)

	integrationSvc, err := newIntegrationServices(db.Pool)
	if err != nil {
		return nil, err
	}

	orchestrationSvc := newOrchestrationServices(db.Pool, workQueue, integrationSvc, identitySvc)

	metricsCounters := metrics.New()
	eventsSvc := newEventsServices(db.Pool, metricsCounters)
	wireEventsBridges(eventsSvc, identitySvc, integrationSvc, orchestrationSvc)

	seedIntegrationConnectors(db.Pool, integrationSvc)

	app := &Application{
		config:        cfg,
		db:            db,
		redis:         redisClient,
		tokens:        tokenSvc,
		cacheSvc:      cacheSvc,
		workQueue:     workQueue,
		apiKeyAuth:    apiKeyAuth,
		identity:      identitySvc,
		integration:   integrationSvc,
		orchestration: orchestrationSvc,
		events:        eventsSvc,
		metrics:       metricsCounters,
	}

	router := gin.New()
	app.registerMiddleware(router)
	app.registerRoutes(router)
	app.router = router

	return app, nil
}

func newAPIKeyAuthenticator(dbAuth apikey.Authenticator) apikey.Authenticator {
	bootstrap := env.GetEnv("API_KEY", "")
	if bootstrap == "" {
		return dbAuth
	}
	return &fallbackAuthenticator{
		primary:  dbAuth,
		fallback: apikey.NewStaticAuthenticator(bootstrap),
	}
}

// fallbackAuthenticator tries Workspace DB keys first, then optional bootstrap API_KEY.
type fallbackAuthenticator struct {
	primary  apikey.Authenticator
	fallback apikey.Authenticator
}

func (f *fallbackAuthenticator) Authenticate(ctx context.Context, plaintextKey string) (apikey.Principal, error) {
	principal, err := f.primary.Authenticate(ctx, plaintextKey)
	if err == nil {
		return principal, nil
	}
	return f.fallback.Authenticate(ctx, plaintextKey)
}

func newIntegrationServices(pool *pgxpool.Pool) (*integrationapp.Services, error) {
	key := env.GetEnv("CREDENTIAL_ENCRYPTION_KEY", "")
	if key == "" {
		key = "dev-only-insecure-32-byte-key!!!"
	}
	encryptor, err := integrationinfra.NewAESSecretEncryptor([]byte(key))
	if err != nil {
		return nil, err
	}

	runtimeRegistry := connectors.NewRegistry(fake.New(), misa.New(), nhanh.New())

	return &integrationapp.Services{
		Connectors:  integrationinfra.NewConnectorRepository(pool),
		Connections: integrationinfra.NewConnectionRepository(pool),
		Credentials: integrationinfra.NewCredentialRepository(pool),
		SyncRoutes:  integrationinfra.NewSyncRouteRepository(pool),
		Watermarks:  integrationinfra.NewSyncRouteWatermarkRepository(pool),
		Runtimes:    runtimeRegistry,
		Secrets:     encryptor,
	}, nil
}

// newOrchestrationServices wires the Orchestration Application layer. ConnectionGateway and
// ConnectorGateway adapt Integration + Identity Domain repositories (Infrastructure-level
// composition) so Orchestration's own Domain/Application never import Integration/Identity.
func newOrchestrationServices(
	pool *pgxpool.Pool,
	workQueue queue.Queue,
	integrationSvc *integrationapp.Services,
	identitySvc *identityapp.Services,
) *orchestrationapp.Services {
	connectionGateway := orchestrationinfra.NewConnectionGateway(
		integrationSvc.Connections,
		integrationSvc.Connectors,
		integrationSvc.Credentials,
		identitySvc.Workspaces,
		identitySvc.Orgs,
		integrationSvc.Secrets,
	)
	connectorGateway := orchestrationinfra.NewConnectorGateway(integrationSvc.Runtimes)
	syncRouteGateway := orchestrationinfra.NewSyncRouteGateway(
		integrationSvc.SyncRoutes,
		integrationSvc.Watermarks,
		identitySvc.Workspaces,
		integrationSvc.Secrets,
	)
	transformer := orchestrationinfra.NewTransformerAdapter(transformationapp.NewServices())

	return &orchestrationapp.Services{
		Intents:     orchestrationinfra.NewIntentRepository(pool),
		Executions:  orchestrationinfra.NewExecutionRepository(pool),
		Idempotency: orchestrationinfra.NewIdempotencyRepository(pool),
		Connections: connectionGateway,
		Connectors:  connectorGateway,
		SyncRoutes:  syncRouteGateway,
		Transformer: transformer,
		Jobs:        orchestrationinfra.NewQueueJobEnqueuer(workQueue),
	}
}

// newEventsServices wires the Events BC application layer (F1/F2/F3): Postgres-backed
// EventRepository/EventReader/AuditRepository, in-memory metrics counters, and a logging
// hook for best-effort in-process subscriber failures. It also subscribes the metrics
// counters to the two Runtime facts the Phase F exit criteria care about
// (ExecutionSucceeded/ExecutionFailed).
func newEventsServices(pool *pgxpool.Pool, counters *metrics.Counters) *eventsapp.Services {
	eventRepo := eventsinfra.NewEventRepository(pool)
	svc := &eventsapp.Services{
		Events:  eventRepo,
		Reader:  eventRepo,
		Audit:   eventsinfra.NewAuditRepository(pool),
		Metrics: counters,
		OnSubscriberError: func(event *eventsdomain.PlatformEvent, err error) {
			if logging.Log != nil {
				logging.Log.Warn().Err(err).
					Str("event_name", event.EventName()).
					Str("event_id", event.ID().String()).
					Msg("events: in-process subscriber failed")
			}
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

// wireEventsBridges replaces every BC's default (nil -> Noop) EventPublisher/AuditRecorder
// with a thin bridge over the Events BC Services (internal/events/infrastructure), so
// Identity/Integration/Orchestration never import the Events BC directly (AGENTS.md package
// boundaries). Must run after every *app.Services value has been constructed.
func wireEventsBridges(
	eventsSvc *eventsapp.Services,
	identitySvc *identityapp.Services,
	integrationSvc *integrationapp.Services,
	orchestrationSvc *orchestrationapp.Services,
) {
	identitySvc.Events = eventsinfra.NewIdentityEventBridge(eventsSvc)
	identitySvc.Audit = eventsinfra.NewIdentityAuditBridge(eventsSvc)

	integrationSvc.Events = eventsinfra.NewIntegrationEventBridge(eventsSvc)
	integrationSvc.Audit = eventsinfra.NewIntegrationAuditBridge(eventsSvc)

	orchestrationSvc.Events = eventsinfra.NewOrchestrationEventBridge(eventsSvc)
	orchestrationSvc.Audit = eventsinfra.NewOrchestrationAuditBridge(eventsSvc)
}

// seedIntegrationConnectors registers built-in Connectors (fake, misa, nhanh) on startup so
// Orchestration can resolve runtimes by catalog code. Failures are logged but never block
// application startup (e.g. first boot before migrations have run).
func seedIntegrationConnectors(pool *pgxpool.Pool, svc *integrationapp.Services) {
	if svc == nil {
		return
	}
	ctx := context.Background()
	if err := persistence.WithinTransaction(ctx, pool, func(ctx context.Context) error {
		return svc.SeedBuiltInConnectors(ctx)
	}); err != nil && logging.Log != nil {
		logging.Log.Warn().Err(err).Msg("failed to seed built-in connectors (will retry lazily on demand)")
	}
}

// identityMembershipChecker adapts the Identity BC's MembershipRepository to the
// integrationhttp.MembershipChecker port required by Integration handlers.
type identityMembershipChecker struct {
	memberships identitydomain.MembershipRepository
}

func (c *identityMembershipChecker) Check(ctx context.Context, workspaceID, userID uuid.UUID) error {
	if _, err := c.memberships.Find(ctx, workspaceID, userID); err != nil {
		if err == identitydomain.ErrNotFound {
			return apperr.New("forbidden", apperr.ErrCodeForbidden)
		}
		return apperr.Wrap(err, "membership lookup failed", apperr.ErrCodeInternal)
	}
	return nil
}

func (a *Application) registerMiddleware(router *gin.Engine) {
	httpLogger := logging.NewLoggerWithPath("access.log", "info")
	recoveryLogger := logging.NewLoggerWithPath("recovery.log", "error")
	rateLimiterLogger := logging.NewLoggerWithPath("rate_limiter.log", "warn")

	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(
		middleware.RateLimitMiddleware(rateLimiterLogger),
		middleware.CORSMiddleware(),
		middleware.TraceMiddleware(),
		middleware.LoggerMiddleware(httpLogger),
		middleware.RecoveryMiddleware(recoveryLogger),
	)

	middleware.InitAuthMiddleware(a.tokens, a.cacheSvc)
}

func (a *Application) registerRoutes(router *gin.Engine) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "hublio-api",
		})
	})

	router.GET("/ready", a.readyHandler)

	if docsui.Enabled() {
		docsui.Register(router)
	}

	api := router.Group("/api/v1")

	identityHandler := identityhttp.NewHandler(a.identity, a.db.Pool, a.tokens)
	identityHandler.RegisterRoutes(api, middleware.AuthMiddleware())

	membershipChecker := &identityMembershipChecker{memberships: a.identity.Memberships}
	integrationHandler := integrationhttp.NewHandler(a.integration, a.db.Pool, membershipChecker)
	integrationHandler.RegisterRoutes(api, middleware.AuthMiddleware())

	orchestrationHandler := orchestrationhttp.NewHandler(a.orchestration, a.db.Pool)
	orchestrationHandler.RegisterRoutes(api, middleware.APIKeyMiddleware(a.apiKeyAuth))
	orchestrationHandler.RegisterWebhookRoutes(api)

	eventsHandler := eventshttp.NewHandler(a.events)
	eventsHandler.RegisterRoutes(api, middleware.APIKeyMiddleware(a.apiKeyAuth))

	machine := api.Group("")
	machine.Use(middleware.APIKeyMiddleware(a.apiKeyAuth))
	{
		machine.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		machine.POST("/platform/queue/health", a.enqueueHealthHandler)
		machine.GET("/platform/metrics", a.metricsHandler)
	}

	router.GET("/metrics", a.metricsHandler)

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found", "path": c.Request.URL.Path})
	})
}

func (a *Application) readyHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	status := gin.H{
		"status":         "ok",
		"postgres":       "ok",
		"redis":          "ok",
		"correlation_id": requestctx.CorrelationID(c.Request.Context()),
		"request_id":     requestctx.RequestID(c.Request.Context()),
	}
	code := http.StatusOK

	if err := a.db.Ping(ctx); err != nil {
		status["postgres"] = "error"
		status["status"] = "degraded"
		code = http.StatusServiceUnavailable
	}

	if err := a.redis.Ping(ctx).Err(); err != nil {
		status["redis"] = "error"
		status["status"] = "degraded"
		code = http.StatusServiceUnavailable
	}

	c.JSON(code, status)
}

// metricsHandler exposes the in-memory Phase F Observability counters as plain JSON
// (AGENTS.md/checklist: avoid a new Prometheus client dependency for now) plus, best-effort,
// the current Redis work-queue depth. Unauthenticated on /metrics (no business data, only
// aggregate counts) and duplicated at /api/v1/platform/metrics behind API-Key auth.
func (a *Application) metricsHandler(c *gin.Context) {
	snapshot := a.metrics.Snapshot()
	body := gin.H{
		"executions_succeeded_total": snapshot.ExecutionsSucceeded,
		"executions_failed_total":    snapshot.ExecutionsFailed,
		"events_published_total":     snapshot.EventsPublished,
		"audit_records_total":        snapshot.AuditRecords,
	}
	if a.workQueue != nil {
		if depth, err := a.workQueue.Depth(c.Request.Context()); err == nil {
			body["queue_depth"] = depth
		}
	}
	c.JSON(http.StatusOK, body)
}

func (a *Application) enqueueHealthHandler(c *gin.Context) {
	if err := queue.EnqueueHealth(c.Request.Context(), a.workQueue); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"status":  "accepted",
		"message": "platform.health job enqueued",
	})
}

func (a *Application) Run() error {
	server := &http.Server{
		Addr:    a.config.ServerAddress,
		Handler: a.router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		if logging.Log != nil {
			logging.Log.Info().Msgf("starting server on %s", a.config.ServerAddress)
		}
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			if logging.Log != nil {
				logging.Log.Fatal().Err(err).Msg("failed to start server")
			}
		}
	}()

	<-quit
	if logging.Log != nil {
		logging.Log.Info().Msg("shutting down server")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return err
	}

	a.db.Close()
	_ = a.redis.Close()

	if logging.Log != nil {
		logging.Log.Info().Msg("server shutdown successfully")
	}

	return nil
}
