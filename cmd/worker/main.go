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
	"hublio/internal/platform/queue"

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
	redisClient := config.NewRedisClient()

	queueLogger := logging.NewLoggerWithPath("queue.log", "info")
	workQueue := queue.NewRedisQueue(redisClient, queueLogger)

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
	_ = redisClient.Close()
	_ = shutdownCtx

	logging.Log.Info().Msg("worker shutdown successfully")
}
