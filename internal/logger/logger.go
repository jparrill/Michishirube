package logger

import (
	"context"
	"log/slog"
	"os"
)

type contextKey string

const loggerKey contextKey = "logger"

// WithLogger adds a logger to the context
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext extracts the logger from context, returns default if not found
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// NewLogger creates a new structured logger with the specified level
func NewLogger(level slog.Level) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	
	handler := slog.NewTextHandler(os.Stdout, opts)
	return slog.New(handler)
}

// NewJSONLogger creates a new JSON logger for production
func NewJSONLogger(level slog.Level) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	
	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}

// WithFields adds structured fields to the logger in context
func WithFields(ctx context.Context, fields ...any) context.Context {
	logger := FromContext(ctx)
	newLogger := logger.With(fields...)
	return WithLogger(ctx, newLogger)
}