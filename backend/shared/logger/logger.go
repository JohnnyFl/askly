package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var (
	// Log is the global logger instance
	Log zerolog.Logger
)

func init() {
	// Initialize global logger
	Log = NewLogger()
}

// NewLogger creates a new configured zerolog logger
func NewLogger() zerolog.Logger {
	// Configure zerolog for AWS Lambda
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.TimestampFieldName = "timestamp"
	zerolog.LevelFieldName = "level"
	zerolog.MessageFieldName = "message"

	// Set log level from environment variable (default: info)
	logLevel := getLogLevel()
	zerolog.SetGlobalLevel(logLevel)

	// Create logger with JSON output to stdout
	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger()

	// Add caller info if enabled (useful for debugging)
	if shouldEnableCaller() {
		logger = logger.With().Caller().Logger()
	}

	return logger
}

// getLogLevel returns the log level from environment variable
func getLogLevel() zerolog.Level {
	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		levelStr = "info" // Default to info for production
	}

	switch strings.ToLower(levelStr) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

// shouldEnableCaller determines if caller info should be added to logs
func shouldEnableCaller() bool {
	enable := os.Getenv("ENABLE_CALLER")
	return strings.ToLower(enable) == "true" || enable == "1"
}

// LogDuration logs the duration of an operation
func LogDuration(logger zerolog.Logger, operation string, start time.Time) {
	duration := time.Since(start)
	logger.Info().
		Dur("duration_ms", duration).
		Str("operation", operation).
		Msg("Operation completed")
}

// LogError logs an error with context
func LogError(logger zerolog.Logger, err error, operation string, context map[string]interface{}) {
	event := logger.Error().Err(err).Str("operation", operation)

	// Add additional context fields
	for key, value := range context {
		event = event.Interface(key, value)
	}

	event.Msg("Operation failed")
}

// LogDebug logs debug information with context
func LogDebug(logger zerolog.Logger, message string, context map[string]interface{}) {
	event := logger.Debug()

	// Add context fields
	for key, value := range context {
		event = event.Interface(key, value)
	}

	event.Msg(message)
}

// LogInfo logs informational message with context
func LogInfo(logger zerolog.Logger, message string, context map[string]interface{}) {
	event := logger.Info()

	// Add context fields
	for key, value := range context {
		event = event.Interface(key, value)
	}

	event.Msg(message)
}

// LogWarn logs warning message with context
func LogWarn(logger zerolog.Logger, message string, context map[string]interface{}) {
	event := logger.Warn()

	// Add context fields
	for key, value := range context {
		event = event.Interface(key, value)
	}

	event.Msg(message)
}
