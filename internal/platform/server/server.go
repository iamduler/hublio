package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	identityapp "hublio/internal/identity/application"
	identitydomain "hublio/internal/identity/domain"
	identityinfra "hublio/internal/identity/infrastructure"
	identityhttp "hublio/internal/identity/interfaces"
	integrationapp "hublio/internal/integration/application"
	"hublio/internal/integration/connectors"
	"hublio/internal/integration/connectors/fake"
	integrationinfra "hublio/internal/integration/infrastructure"
	integrationhttp "hublio/internal/integration/interfaces"
	"hublio/internal/platform/apikey"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/auth"
	"hublio/internal/platform/cache"
	"hublio/internal/platform/config"
	"hublio/internal/platform/docsui"
	"hublio/internal/platform/env"
	"hublio/internal/platform/logging"
	"hublio/internal/platform/middleware"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/queue"
	"hublio/internal/platform/requestctx"
	"hublio/internal/platform/validation"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Application struct {
	config      *config.Config
	router      *gin.Engine
	db          *persistence.Database
	redis       *redis.Client
	tokens      auth.TokenService
	cacheSvc    cache.RedisCacheService
	workQueue   queue.Queue
	apiKeyAuth  apikey.Authenticator
	identity    *identityapp.Services
	integration *integrationapp.Services
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
		Events:      identityapp.NoopPublisher{},
	}

	dbAuth := identityinfra.NewDBAuthenticator(keyRepo, wsRepo, orgRepo)
	apiKeyAuth := newAPIKeyAuthenticator(dbAuth)

	integrationSvc, err := newIntegrationServices(db.Pool)
	if err != nil {
		return nil, err
	}
	seedIntegrationConnectors(db.Pool, integrationSvc)

	app := &Application{
		config:      cfg,
		db:          db,
		redis:       redisClient,
		tokens:      tokenSvc,
		cacheSvc:    cacheSvc,
		workQueue:   workQueue,
		apiKeyAuth:  apiKeyAuth,
		identity:    identitySvc,
		integration: integrationSvc,
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

	runtimeRegistry := connectors.NewRegistry(fake.New())

	return &integrationapp.Services{
		Connectors:  integrationinfra.NewConnectorRepository(pool),
		Connections: integrationinfra.NewConnectionRepository(pool),
		Credentials: integrationinfra.NewCredentialRepository(pool),
		Runtimes:    runtimeRegistry,
		Secrets:     encryptor,
		Events:      integrationapp.NoopPublisher{},
	}, nil
}

// seedIntegrationConnectors registers the built-in "fake" Connector on startup so Orchestration
// can rely on it being available. Failures are logged but never block application startup
// (e.g. first boot before migrations have run).
func seedIntegrationConnectors(pool *pgxpool.Pool, svc *integrationapp.Services) {
	if svc == nil {
		return
	}
	ctx := context.Background()
	if err := persistence.WithinTransaction(ctx, pool, func(ctx context.Context) error {
		_, err := svc.SeedFakeConnector(ctx)
		return err
	}); err != nil && logging.Log != nil {
		logging.Log.Warn().Err(err).Msg("failed to seed fake connector (will retry lazily on demand)")
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

	machine := api.Group("")
	machine.Use(middleware.APIKeyMiddleware(a.apiKeyAuth))
	{
		machine.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		machine.POST("/platform/queue/health", a.enqueueHealthHandler)
	}

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
