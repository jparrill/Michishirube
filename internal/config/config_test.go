package config

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"michishirube/internal/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		valid  bool
	}{
		{
			name: "valid config",
			config: Config{
				Port:     "8080",
				DBPath:   "test.db",
				LogLevel: "info",
			},
			valid: true,
		},
		{
			name: "empty port",
			config: Config{
				DBPath:   "test.db",
				LogLevel: "info",
			},
			valid: false,
		},
		{
			name: "empty db_path",
			config: Config{
				Port:     "8080",
				LogLevel: "info",
			},
			valid: false,
		},
		{
			name: "invalid log level",
			config: Config{
				Port:     "8080",
				DBPath:   "test.db",
				LogLevel: "invalid",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.config.validateAndFix(slog.Default())
			
			if tt.valid {
				// Config should remain valid
				assert.NotEmpty(t, tt.config.Port)
				assert.NotEmpty(t, tt.config.DBPath)
				assert.True(t, isValidLogLevel(tt.config.LogLevel))
			} else {
				// After validateAndFix, config should be valid (fixed with defaults)
				assert.NotEmpty(t, tt.config.Port, "Port should be fixed with default")
				assert.NotEmpty(t, tt.config.DBPath, "DBPath should be fixed with default")
				assert.True(t, isValidLogLevel(tt.config.LogLevel), "LogLevel should be fixed with default")
			}
		})
	}
}

func TestConfig_GetSlogLevel(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		expected slog.Level
	}{
		{"debug level", "debug", slog.LevelDebug},
		{"info level", "info", slog.LevelInfo},
		{"warn level", "warn", slog.LevelWarn},
		{"error level", "error", slog.LevelError},
		{"uppercase debug", "DEBUG", slog.LevelDebug},
		{"mixed case info", "Info", slog.LevelInfo},
		{"invalid defaults to info", "invalid", slog.LevelInfo},
		{"empty defaults to info", "", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{LogLevel: tt.logLevel}
			assert.Equal(t, tt.expected, config.GetSlogLevel())
		})
	}
}

func TestIsValidLogLevel(t *testing.T) {
	tests := []struct {
		name  string
		level string
		valid bool
	}{
		{"valid debug", "debug", true},
		{"valid info", "info", true},
		{"valid warn", "warn", true},
		{"valid error", "error", true},
		{"valid uppercase", "DEBUG", true},
		{"valid mixed case", "Info", true},
		{"invalid level", "invalid", false},
		{"empty level", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, isValidLogLevel(tt.level))
		})
	}
}

func TestLoad_WithDefaults(t *testing.T) {
	// Create temporary config file for testing
	tempFile := "test_config.yaml"
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			t.Logf("failed to remove temp file: %v", err)
		}
	}()
	
	// Test with no config file (should use defaults)
	baseLogger := logger.NewLogger(slog.LevelInfo)
	ctx := logger.WithLogger(context.Background(), baseLogger)
	
	config, err := Load(ctx)
	require.NoError(t, err)
	
	assert.Equal(t, "8080", config.Port)
	assert.Equal(t, "michishirube.db", config.DBPath)
	assert.Equal(t, "info", config.LogLevel)
}

func TestLoad_WithConfigFile(t *testing.T) {
	// Create temporary config file
	tempConfigContent := `port: "9090"
db_path: "custom.db"
log_level: "debug"
`
	
	// Save current directory and change back after test
	originalConfig := "config.yaml"
	var originalContent []byte
	var hadOriginal bool
	
	// Backup original config if exists
	if data, err := os.ReadFile(originalConfig); err == nil {
		originalContent = data
		hadOriginal = true
	}
	
	// Write test config
	err := os.WriteFile(originalConfig, []byte(tempConfigContent), 0644)
	require.NoError(t, err)
	
	defer func() {
		if hadOriginal {
			if err := os.WriteFile(originalConfig, originalContent, 0644); err != nil {
				t.Logf("failed to restore original config: %v", err)
			}
		} else {
			if err := os.Remove(originalConfig); err != nil {
				t.Logf("failed to remove test config: %v", err)
			}
		}
	}()
	
	baseLogger := logger.NewLogger(slog.LevelInfo)
	ctx := logger.WithLogger(context.Background(), baseLogger)
	
	config, err := Load(ctx)
	require.NoError(t, err)
	
	assert.Equal(t, "9090", config.Port)
	assert.Equal(t, "custom.db", config.DBPath)
	assert.Equal(t, "debug", config.LogLevel)
}

func TestLoad_WithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	originalPort := os.Getenv("PORT")
	originalDBPath := os.Getenv("DB_PATH")
	originalLogLevel := os.Getenv("LOG_LEVEL")
	
	require.NoError(t, os.Setenv("PORT", "3000"))
	require.NoError(t, os.Setenv("DB_PATH", "env.db"))
	require.NoError(t, os.Setenv("LOG_LEVEL", "error"))
	
	defer func() {
		if originalPort == "" {
			if err := os.Unsetenv("PORT"); err != nil {
				t.Logf("failed to unset PORT: %v", err)
			}
		} else {
			if err := os.Setenv("PORT", originalPort); err != nil {
				t.Logf("failed to restore PORT: %v", err)
			}
		}
		if originalDBPath == "" {
			if err := os.Unsetenv("DB_PATH"); err != nil {
				t.Logf("failed to unset DB_PATH: %v", err)
			}
		} else {
			if err := os.Setenv("DB_PATH", originalDBPath); err != nil {
				t.Logf("failed to restore DB_PATH: %v", err)
			}
		}
		if originalLogLevel == "" {
			if err := os.Unsetenv("LOG_LEVEL"); err != nil {
				t.Logf("failed to unset LOG_LEVEL: %v", err)
			}
		} else {
			if err := os.Setenv("LOG_LEVEL", originalLogLevel); err != nil {
				t.Logf("failed to restore LOG_LEVEL: %v", err)
			}
		}
	}()
	
	baseLogger := logger.NewLogger(slog.LevelInfo)
	ctx := logger.WithLogger(context.Background(), baseLogger)
	
	config, err := Load(ctx)
	require.NoError(t, err)
	
	assert.Equal(t, "3000", config.Port)
	assert.Equal(t, "env.db", config.DBPath)
	assert.Equal(t, "error", config.LogLevel)
}

func TestPredefinedErrors(t *testing.T) {
	// Test that predefined errors have expected messages
	assert.Equal(t, "port cannot be empty", ErrPortEmpty.Error())
	assert.Equal(t, "db_path cannot be empty", ErrDBPathEmpty.Error())
	assert.Equal(t, "failed to parse config.yaml", ErrConfigParse.Error())
	assert.Equal(t, "invalid log level", ErrInvalidLogLevel.Error())
}