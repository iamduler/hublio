package main

import (
	"path/filepath"

	"hublio/internal/platform/config"
	"hublio/internal/platform/env"
	"hublio/internal/platform/logging"
	"hublio/internal/platform/server"

	"github.com/joho/godotenv"
)

func main() {
	rootDir := env.MustGetWorkingDir()
	logFilePath := filepath.Join(rootDir, "logs", "app.log")

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
		logging.Log.Info().Msg("environment variables loaded for api")
	}

	cfg := config.NewConfig()

	app, err := server.NewApplication(cfg)
	if err != nil {
		logging.Log.Fatal().Err(err).Msg("failed to initialize application")
	}

	if err := app.Run(); err != nil {
		logging.Log.Fatal().Err(err).Msg("failed to run application")
	}
}
