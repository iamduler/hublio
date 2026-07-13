package main

import (
	"path/filepath"
	"shopping-cart/internal/app"
	"shopping-cart/internal/config"
	"shopping-cart/internal/utils"
	"shopping-cart/pkg/logger"

	"github.com/joho/godotenv"
)

func main() {
	rootDir := utils.MustGetWorkingDir()

	logFilePath := filepath.Join(rootDir, "internal", "logs", "app.log")

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
		logger.Log.Info().Msg("Environment variables loaded successfully for api")
	}

	// Initialize config
	config := config.NewConfig()

	// Initialize application
	app, err := app.NewApplication(config)

	if err != nil {
		logger.Log.Fatal().Err(err).Msgf("❌ Failed to initialize application: %v", err)
		return
	}

	// Initialize server
	if err := app.Run(); err != nil {
		logger.Log.Fatal().Err(err).Msgf("❌ Failed to run application: %v", err)
	}
}
