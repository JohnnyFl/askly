package logger

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/rs/zerolog"
)

type contextKey string

const (
	loggerKey contextKey = "logger"
)

// ToContext adds a logger to the context
func ToContext(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, &logger)
}

// FromContext retrieves the logger from context, or returns the global logger if not found
func FromContext(ctx context.Context) zerolog.Logger {
	if ctx == nil {
		return Log
	}

	if logger, ok := ctx.Value(loggerKey).(*zerolog.Logger); ok && logger != nil {
		return *logger
	}

	return Log
}

// NewContextWithRequestID creates a context with a logger that includes request ID
func NewContextWithRequestID(ctx context.Context, requestID string) context.Context {
	logger := Log.With().Str("request_id", requestID).Logger()
	return ToContext(ctx, logger)
}

// NewContextFromAPIGatewayRequest creates a context with logger from API Gateway request
func NewContextFromAPIGatewayRequest(ctx context.Context, request events.APIGatewayProxyRequest) context.Context {
	logger := Log.With().
		Str("request_id", request.RequestContext.RequestID).
		Str("method", request.HTTPMethod).
		Str("path", request.Path).
		Str("source_ip", request.RequestContext.Identity.SourceIP).
		Logger()

	// Add path parameters if present
	if len(request.PathParameters) > 0 {
		logger = logger.With().Interface("path_params", request.PathParameters).Logger()
	}

	// Add query parameters if present
	if len(request.QueryStringParameters) > 0 {
		logger = logger.With().Interface("query_params", request.QueryStringParameters).Logger()
	}

	return ToContext(ctx, logger)
}

// WithField adds a field to the logger in context
func WithField(ctx context.Context, key string, value interface{}) context.Context {
	logger := FromContext(ctx).With().Interface(key, value).Logger()
	return ToContext(ctx, logger)
}

// WithFields adds multiple fields to the logger in context
func WithFields(ctx context.Context, fields map[string]interface{}) context.Context {
	logger := FromContext(ctx)
	loggerContext := logger.With()

	for key, value := range fields {
		loggerContext = loggerContext.Interface(key, value)
	}

	return ToContext(ctx, loggerContext.Logger())
}

// WithUserID adds user ID to the logger in context
func WithUserID(ctx context.Context, userID string) context.Context {
	logger := FromContext(ctx).With().Str("user_id", userID).Logger()
	return ToContext(ctx, logger)
}

// WithAdminID adds admin ID to the logger in context
func WithAdminID(ctx context.Context, adminID string) context.Context {
	logger := FromContext(ctx).With().Str("admin_id", adminID).Logger()
	return ToContext(ctx, logger)
}

// WithChannelID adds channel ID to the logger in context
func WithChannelID(ctx context.Context, channelID string) context.Context {
	logger := FromContext(ctx).With().Str("channel_id", channelID).Logger()
	return ToContext(ctx, logger)
}

// WithTranscriptionID adds transcription ID to the logger in context
func WithTranscriptionID(ctx context.Context, transcriptionID string) context.Context {
	logger := FromContext(ctx).With().Str("transcription_id", transcriptionID).Logger()
	return ToContext(ctx, logger)
}

// WithOperation adds operation name to the logger in context
func WithOperation(ctx context.Context, operation string) context.Context {
	logger := FromContext(ctx).With().Str("operation", operation).Logger()
	return ToContext(ctx, logger)
}
