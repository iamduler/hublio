package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hublio/internal/platform/auth"
	"hublio/internal/platform/cache"
	"hublio/internal/platform/config"
	"hublio/internal/platform/logging"
	"hublio/internal/platform/middleware"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/validation"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Application struct {
	config   *config.Config
	router   *gin.Engine
	db       *persistence.Database
	redis    *redis.Client
	tokens   auth.TokenService
	cacheSvc cache.RedisCacheService
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

	router := gin.New()
	registerMiddleware(router, tokenSvc, cacheSvc)
	registerRoutes(router)

	return &Application{
		config:   cfg,
		router:   router,
		db:       db,
		redis:    redisClient,
		tokens:   tokenSvc,
		cacheSvc: cacheSvc,
	}, nil
}

func registerMiddleware(router *gin.Engine, tokenSvc auth.TokenService, cacheSvc cache.RedisCacheService) {
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

	middleware.InitAuthMiddleware(tokenSvc, cacheSvc)
}

func registerRoutes(router *gin.Engine) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "hublio-api",
		})
	})

	api := router.Group("/api/v1")
	api.Use(middleware.ApiKeyMiddleware())
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found", "path": c.Request.URL.Path})
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
