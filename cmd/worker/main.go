package main

import (
	"context"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"hublio/internal/platform/config"
	"hublio/internal/platform/env"
	"hublio/internal/platform/logging"
	"hublio/internal/platform/messaging"

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

	_ = config.NewConfig()

	rabbitLogger := logging.NewLoggerWithPath("rabbitmq.log", "info")
	rabbitMQURL := env.GetEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")

	mq, err := messaging.NewRabbitMQService(rabbitMQURL, rabbitLogger)
	if err != nil {
		logging.Log.Fatal().Err(err).Msg("failed to initialize messaging")
	}
	defer func() {
		if closeErr := mq.Close(); closeErr != nil {
			logging.Log.Error().Err(closeErr).Msg("failed to close messaging connection")
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	logging.Log.Info().Msg("worker started; waiting for orchestration consumers")

	<-ctx.Done()
	logging.Log.Info().Msg("shutting down worker")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = shutdownCtx

	logging.Log.Info().Msg("worker shutdown successfully")
}
