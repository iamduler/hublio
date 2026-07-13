package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"shopping-cart/internal/config"
	"shopping-cart/internal/db"
	"shopping-cart/internal/db/sqlc"
	"shopping-cart/internal/routes"
	"shopping-cart/internal/utils"
	"shopping-cart/internal/validation"
	"shopping-cart/pkg/auth"
	"shopping-cart/pkg/cache"
	"shopping-cart/pkg/logger"
	"shopping-cart/pkg/mail"
	"shopping-cart/pkg/rabbitmq"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Module interface {
	Routes() routes.Routes
}

type Application struct {
	config  *config.Config
	router  *gin.Engine
	modules []Module
}

type ModuleContext struct {
	DB    sqlc.Querier
	Redis *redis.Client
}

func NewApplication(cfg *config.Config) (*Application, error) {
	if err := validation.InitValidator(); err != nil {
		logger.Log.Fatal().Err(err).Msgf("❌ Failed to initialize validator: %v", err)
		return nil, err
	}

	router := gin.Default()

	// Initialize database
	if err := db.InitDB(); err != nil {
		logger.Log.Fatal().Err(err).Msgf("❌ Failed to initialize database: %v", err)
		return nil, err
	}

	// Initialize Redis
	redisClient := config.NewRedisClient()

	// Initialize cache
	cacheService := cache.NewRedisCacheService(redisClient)

	// Initialize token service
	tokenService := auth.NewJWTService(cacheService)

	// Initialize mail logger
	mailLogger := utils.NewLoggerWithPath("mail.log", "info")

	// Initialize mail service
	factory, err := mail.NewProviderFactory(mail.ProviderMailtrap)

	if err != nil {
		mailLogger.Error().Err(err).Msgf("❌ Failed to initialize mail provider factory: %v", err)
		return nil, err
	}

	mailService, err := mail.NewMailService(cfg, mailLogger, factory)

	if err != nil {
		mailLogger.Error().Err(err).Msgf("❌ Failed to initialize mail service: %v", err)
		return nil, err
	}

	// Initialize RabbitMQ
	rabbitMQLogger := utils.NewLoggerWithPath("rabbitmq.log", "info")
	rabbitMQURL := utils.GetEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
	rabbitMQService, err := rabbitmq.NewRabbitMQService(rabbitMQURL, rabbitMQLogger)

	if err != nil {
		rabbitMQLogger.Error().Err(err).Msgf("❌ Failed to initialize RabbitMQ service: %v", err)
		return nil, err
	}

	ctx := &ModuleContext{
		DB:    db.DB,
		Redis: redisClient,
	}

	modules := []Module{
		NewUserModule(ctx),
		NewAuthModule(ctx, tokenService, cacheService, mailService, rabbitMQService),
	}

	routes.RegisterRoutes(router, tokenService, cacheService, getModuleRoutes(modules)...)

	return &Application{
		config:  cfg,
		router:  router,
		modules: modules,
	}, nil
}

func (a *Application) Run() error {
	server := &http.Server{
		Addr:    a.config.ServerAddress,
		Handler: a.router,
	}

	quit := make(chan os.Signal, 1)

	// Wait for interrupt signal to gracefully shutdown the server
	// syscall.SIGINT: Interrupt signal (Ctrl+C)
	// syscall.SIGTERM: Termination signal
	// syscall.SIGHUP: Reload signal
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		logger.Log.Info().Msgf("🚀 Starting server on %s", a.config.ServerAddress)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Log.Fatal().Err(err).Msg("❌ Failed to start server")
		}
	}()

	<-quit
	logger.Log.Info().Msg("🙋‍♂️ Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Log.Fatal().Err(err).Msg("❌ Failed to shutdown server")
	}

	logger.Log.Info().Msg("👋 Server shutdown successfully")

	return nil
}

func getModuleRoutes(modules []Module) []routes.Routes {
	routes := make([]routes.Routes, len(modules))

	for i, module := range modules {
		routes[i] = module.Routes()
	}

	return routes
}
