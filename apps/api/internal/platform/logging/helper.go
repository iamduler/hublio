package logging

import (
	"os"
	"path/filepath"

	"hublio/internal/platform/env"

	"github.com/rs/zerolog"
)

func NewLoggerWithPath(filename string, logLevel string) *zerolog.Logger {
	cwd, err := os.Getwd()
	if err != nil {
		if Log != nil {
			Log.Fatal().Err(err).Msg("failed to get current working directory")
		}
		panic(err)
	}

	logPath := filepath.Join(cwd, "logs", filename)

	config := LoggerConfig{
		LogLevel:   logLevel,
		Filename:   logPath,
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     30,
		Compress:   true,
		LocalTime:  true,
		IsDev:      env.GetEnv("DEVELOPMENT_MODE", "production"),
	}

	return NewLogger(config)
}
