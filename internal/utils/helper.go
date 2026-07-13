package utils

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"os"
	"path/filepath"
	"shopping-cart/pkg/logger"
	"strconv"

	"github.com/rs/zerolog"
)

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

func GetIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)

	if err != nil {
		return defaultValue
	}

	return intValue
}

func NewLoggerWithPath(path string, logLevel string) *zerolog.Logger {
	cwd, err := os.Getwd()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("❌ Failed to get current working directory")
	}

	logPath := filepath.Join(cwd, "internal", "logs", path)

	config := logger.LoggerConfig{
		LogLevel:   logLevel,
		Filename:   logPath,
		MaxSize:    10, // MB
		MaxBackups: 3,  // backups
		MaxAge:     30, // days
		Compress:   true,
		LocalTime:  true, // Use local time instead of UTC
		IsDev:      GetEnv("DEVELOPMENT_MODE", "production"),
	}

	return logger.NewLogger(config)
}

func GenerateRandomString(length int) (string, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)

	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(randomBytes), nil
}

func MustGetWorkingDir() string {
	cwd, err := os.Getwd()

	if err != nil {
		log.Fatalf("❌ Failed to get current working directory: %v", err)
	}

	return cwd
}
