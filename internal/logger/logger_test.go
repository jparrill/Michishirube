package logger

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithLogger_FromContext(t *testing.T) {
	// Create a test logger
	testLogger := NewLogger(slog.LevelDebug)
	
	// Add logger to context
	ctx := context.Background()
	ctxWithLogger := WithLogger(ctx, testLogger)
	
	// Retrieve logger from context
	retrievedLogger := FromContext(ctxWithLogger)
	
	assert.Equal(t, testLogger, retrievedLogger)
}

func TestFromContext_WithoutLogger(t *testing.T) {
	// Context without logger should return default
	ctx := context.Background()
	logger := FromContext(ctx)
	
	assert.NotNil(t, logger)
	assert.Equal(t, slog.Default(), logger)
}

func TestWithFields(t *testing.T) {
	baseLogger := NewLogger(slog.LevelInfo)
	ctx := WithLogger(context.Background(), baseLogger)
	
	// Add fields to logger in context
	ctxWithFields := WithFields(ctx, "key1", "value1", "key2", "value2")
	
	// Logger should be different (has fields)
	originalLogger := FromContext(ctx)
	fieldsLogger := FromContext(ctxWithFields)
	
	assert.NotEqual(t, originalLogger, fieldsLogger)
}

func TestNewLogger(t *testing.T) {
	logger := NewLogger(slog.LevelDebug)
	
	assert.NotNil(t, logger)
	// Logger should be configured for text output (not JSON)
}

func TestNewJSONLogger(t *testing.T) {
	logger := NewJSONLogger(slog.LevelError)
	
	assert.NotNil(t, logger)
	// Logger should be configured for JSON output
}

func TestLoggerLevels(t *testing.T) {
	levels := []slog.Level{
		slog.LevelDebug,
		slog.LevelInfo,
		slog.LevelWarn,
		slog.LevelError,
	}
	
	for _, level := range levels {
		t.Run(level.String(), func(t *testing.T) {
			textLogger := NewLogger(level)
			jsonLogger := NewJSONLogger(level)
			
			assert.NotNil(t, textLogger)
			assert.NotNil(t, jsonLogger)
		})
	}
}

func TestContextKey(t *testing.T) {
	// Test that the context key is not accessible from outside
	// (it's private, so this mainly tests our implementation)
	
	logger1 := NewLogger(slog.LevelInfo)
	logger2 := NewLogger(slog.LevelDebug)
	
	ctx1 := WithLogger(context.Background(), logger1)
	ctx2 := WithLogger(context.Background(), logger2)
	
	retrievedLogger1 := FromContext(ctx1)
	retrievedLogger2 := FromContext(ctx2)
	
	assert.Equal(t, logger1, retrievedLogger1)
	assert.Equal(t, logger2, retrievedLogger2)
	assert.NotEqual(t, retrievedLogger1, retrievedLogger2)
}

func TestNestedContext(t *testing.T) {
	// Test context nesting and overwriting
	logger1 := NewLogger(slog.LevelInfo)
	logger2 := NewLogger(slog.LevelDebug)
	
	ctx := context.Background()
	ctx = WithLogger(ctx, logger1)
	ctx = WithLogger(ctx, logger2) // Override
	
	retrievedLogger := FromContext(ctx)
	assert.Equal(t, logger2, retrievedLogger)
}