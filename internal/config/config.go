package config

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"

	"michishirube/internal/logger"
	"gopkg.in/yaml.v3"
)

var (
	ErrPortEmpty     = errors.New("port cannot be empty")
	ErrDBPathEmpty   = errors.New("db_path cannot be empty")
	ErrConfigParse   = errors.New("failed to parse config.yaml")
	ErrInvalidLogLevel = errors.New("invalid log level")
)

type Config struct {
	Port     string `yaml:"port"`
	DBPath   string `yaml:"db_path"`
	LogLevel string `yaml:"log_level"`
}

func Load(ctx context.Context) (*Config, error) {
	log := logger.FromContext(ctx)
	
	// Default values
	config := &Config{
		Port:     "8080",
		DBPath:   "michishirube.db",
		LogLevel: "info",
	}

	log.Info("Loading configuration with defaults", "port", config.Port, "db_path", config.DBPath, "log_level", config.LogLevel)

	// Try to read from config file (CONFIG_PATH env var or default config.yaml)
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}
	
	if data, err := os.ReadFile(configPath); err == nil {
		log.Info("Found config file, parsing", "path", configPath)
		if err := yaml.Unmarshal(data, config); err != nil {
			log.Error("Failed to parse config file", "path", configPath, "error", err)
			return nil, ErrConfigParse
		}
		log.Info("Configuration loaded from file", "path", configPath, "port", config.Port, "db_path", config.DBPath, "log_level", config.LogLevel)
	} else {
		log.Info("No config file found, using defaults", "tried_path", configPath)
	}

	// Environment variables override config file
	if port := os.Getenv("PORT"); port != "" {
		if port == "" {
			log.Warn("Invalid PORT from environment (empty), using default", "default", "8080")
		} else {
			log.Info("Overriding port from environment", "port", port)
			config.Port = port
		}
	}

	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		if dbPath == "" {
			log.Warn("Invalid DB_PATH from environment (empty), using default", "default", "michishirube.db")
		} else {
			log.Info("Overriding db_path from environment", "db_path", dbPath)
			config.DBPath = dbPath
		}
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		if !isValidLogLevel(logLevel) {
			log.Warn("Invalid LOG_LEVEL from environment, using default", "invalid", logLevel, "default", "info")
		} else {
			log.Info("Overriding log_level from environment", "log_level", logLevel)
			config.LogLevel = logLevel
		}
	}

	// Validate and fix configuration
	config.validateAndFix(log)

	log.Info("Configuration loaded successfully", "port", config.Port, "db_path", config.DBPath, "log_level", config.LogLevel)
	return config, nil
}

func (c *Config) validateAndFix(log *slog.Logger) {
	if c.Port == "" {
		log.Warn("Invalid port configuration (empty), using default", "default", "8080")
		c.Port = "8080"
	}
	
	if c.DBPath == "" {
		log.Warn("Invalid db_path configuration (empty), using default", "default", "michishirube.db")
		c.DBPath = "michishirube.db"
	}
	
	if !isValidLogLevel(c.LogLevel) {
		log.Warn("Invalid log_level configuration, using default", "invalid", c.LogLevel, "default", "info")
		c.LogLevel = "info"
	}
}

func isValidLogLevel(level string) bool {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}

func (c *Config) GetSlogLevel() slog.Level {
	switch strings.ToLower(c.LogLevel) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}