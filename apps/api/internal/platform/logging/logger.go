package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
)

type contextKey string

// TraceIDKey is the key for the trace ID in the context
const TraceIDKey contextKey = "trace_id"

var Log *zerolog.Logger

type LoggerConfig struct {
	Filename   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	LocalTime  bool
	LogLevel   string
	IsDev      string
}

func InitLogger(config LoggerConfig) {
	Log = NewLogger(config)
}

func NewLogger(config LoggerConfig) *zerolog.Logger {
	var writer io.Writer
	zerolog.TimeFieldFormat = time.RFC3339

	// level, err := zerolog.ParseLevel(config.LogLevel)
	// if err != nil {
	// 	level = zerolog.InfoLevel
	// }

	// zerolog.SetGlobalLevel(level)

	if config.IsDev == "development" {
		if strings.Contains(config.Filename, "app.log") {
			writer = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		} else {
			writer = &PrettyJSONWriter{Writer: os.Stdout}
		}
	} else {
		writer = &lumberjack.Logger{
			Filename:   config.Filename,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
			LocalTime:  config.LocalTime,
		}
	}

	logger := zerolog.New(writer).With().Timestamp().Logger()

	return &logger
}

type PrettyJSONWriter struct {
	Writer io.Writer
}

// Override the Write method of the io.Writer interface
// This method is used to pretty print the JSON output
func (w *PrettyJSONWriter) Write(p []byte) (n int, err error) {
	var prettyJSON bytes.Buffer

	// Indent the JSON output with 2 spaces
	if err := json.Indent(&prettyJSON, p, "", "  "); err != nil {
		return 0, err
	}

	return w.Writer.Write(prettyJSON.Bytes())
}

func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}
