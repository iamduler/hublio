package main

import (
	"context"
	"encoding/json"
	"os/signal"
	"path/filepath"
	"shopping-cart/internal/config"
	"shopping-cart/internal/utils"
	"shopping-cart/pkg/logger"
	"shopping-cart/pkg/mail"
	"shopping-cart/pkg/rabbitmq"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

type Worker struct {
	rabbitMQService rabbitmq.RabbitMQService
	mailService     mail.EmailProviderService
	config          *config.Config
	logger          *zerolog.Logger
}

func NewWorker(cfg *config.Config) *Worker {
	// Initialize RabbitMQ
	logger := utils.NewLoggerWithPath("rabbitmq.log", "info")
	rabbitMQURL := utils.GetEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
	rabbitMQService, err := rabbitmq.NewRabbitMQService(rabbitMQURL, logger)

	if err != nil {
		logger.Error().Err(err).Msgf("❌ Failed to initialize RabbitMQ service: %v", err)
		return nil
	}

	// Initialize mail service
	factory, err := mail.NewProviderFactory(mail.ProviderMailtrap)
	mailLogger := utils.NewLoggerWithPath("mail.log", "info")

	if err != nil {
		mailLogger.Error().Err(err).Msgf("❌ Failed to initialize mail provider factory: %v", err)
		return nil
	}

	mailService, err := mail.NewMailService(cfg, mailLogger, factory)

	if err != nil {
		mailLogger.Error().Err(err).Msgf("❌ Failed to initialize mail service: %v", err)
		return nil
	}

	return &Worker{
		rabbitMQService: rabbitMQService,
		mailService:     mailService,
		config:          cfg,
		logger:          logger,
	}
}

func (w *Worker) Start(ctx context.Context) error {
	const emailQueue = "auth_email_queue"

	handler := func(body []byte) error {
		w.logger.Debug().Msgf("Received email: %s", string(body))

		var email mail.Email

		// Unmarshal the email
		if err := json.Unmarshal(body, &email); err != nil {
			w.logger.Error().Err(err).Msgf("❌ Failed to unmarshal email: %v", err)
			return err
		}

		// Send the email
		if err := w.mailService.SendMail(ctx, &email); err != nil {
			w.logger.Error().Err(err).Msgf("❌ Failed to send email: %v", err)
			return utils.NewError("Failed to send email", utils.ErrCodeInternal)
		}

		w.logger.Info().Msgf("✅ Email sent successfully")

		return nil
	}

	// Consume the email queue
	err := w.rabbitMQService.Consume(ctx, emailQueue, handler)

	if err != nil {
		w.logger.Error().Err(err).Msgf("❌ Failed to consume email queue: %v", err)
		return err
	}

	w.logger.Info().Msgf("✅ Email queue consumed successfully")

	// Wait for the context to be done
	<-ctx.Done()
	w.logger.Info().Msgf("Worker stopped due to context done")

	return ctx.Err()
}

func (w *Worker) Shutdown(ctx context.Context) error {
	w.logger.Info().Msgf("🙋‍♂️ Shutting down worker...")
	err := w.rabbitMQService.Close()

	if err != nil {
		w.logger.Error().Err(err).Msgf("❌ Failed to close RabbitMQ connection: %v", err)
		return err
	}

	w.logger.Info().Msgf("✅ RabbitMQ connection closed successfully")

	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			w.logger.Info().Msgf("Worker stopped due to context deadline exceeded")
			return ctx.Err()
		}
	default:
	}

	w.logger.Info().Msgf("Worker stopped successfully")
	return nil
}

func main() {
	rootDir := utils.MustGetWorkingDir()

	logFilePath := filepath.Join(rootDir, "internal", "logs", "worker.log")

	logger.InitLogger(logger.LoggerConfig{
		LogLevel:   "info",
		Filename:   logFilePath,
		MaxSize:    10, // MB
		MaxBackups: 3,  // backups
		MaxAge:     30, // days
		Compress:   true,
		LocalTime:  true, // Use local time instead of UTC
		IsDev:      utils.GetEnv("DEVELOPMENT_MODE", "development"),
	})

	if err := godotenv.Load(filepath.Join(rootDir, ".env")); err != nil {
		logger.Log.Warn().Msgf("❌ Failed to load environment variables: %v", err)
	} else {
		logger.Log.Info().Msg("Environment variables loaded successfully for worker")
	}

	// Initialize config
	config := config.NewConfig()

	// Initialize worker
	worker := NewWorker(config)

	if worker == nil {
		logger.Log.Error().Msgf("❌ Failed to initialize worker")
		return
	}

	// Wait for interrupt signal to gracefully shutdown the worker
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := worker.Start(ctx); err != nil && err != context.Canceled {
			logger.Log.Error().Err(err).Msgf("❌ Failed to start worker: %v", err)
		}
	}()

	<-ctx.Done()
	logger.Log.Info().Msgf("🙋‍♂️ Shutting down worker...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := worker.Shutdown(ctx); err != nil {
		logger.Log.Error().Err(err).Msgf("❌ Failed to stop worker: %v", err)
	}

	wg.Wait()
	logger.Log.Info().Msgf("👋 Worker shutdown successfully")
}
