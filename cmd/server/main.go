package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"michishirube/internal/config"
	"michishirube/internal/logger"
	"michishirube/internal/server"
	"michishirube/internal/storage/sqlite"
)

func main() {
	// Setup base logger with INFO level initially
	baseLogger := logger.NewLogger(slog.LevelInfo)
	ctx := logger.WithLogger(context.Background(), baseLogger)
	
	log := logger.FromContext(ctx)
	log.Info("Starting Michishirube application")

	// Load configuration
	cfg, err := config.Load(ctx)
	if err != nil {
		log.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Reconfigure logger with the actual log level from config
	actualLogger := logger.NewLogger(cfg.GetSlogLevel())
	ctx = logger.WithLogger(ctx, actualLogger)
	log = logger.FromContext(ctx)

	// Add config info to context for future logging
	ctx = logger.WithFields(ctx, "port", cfg.Port, "db_path", cfg.DBPath, "log_level", cfg.LogLevel)
	log.Info("Logger reconfigured with config level")

	// Ensure database directory exists
	if err := ensureDBDirectory(ctx, cfg.DBPath); err != nil {
		log.Error("Failed to create database directory", "error", err)
		os.Exit(1)
	}

	// Initialize storage
	log.Info("Initializing storage", "db_path", cfg.DBPath)
	storage, err := sqlite.New(cfg.DBPath)
	if err != nil {
		log.Error("Failed to initialize storage", "error", err)
		os.Exit(1)
	}
	defer func() {
		log.Info("Closing storage connection")
		if err := storage.Close(); err != nil {
			log.Error("Failed to close storage", "error", err)
		}
	}()

	log.Info("Storage initialized successfully")

	// Initialize and start server
	srv := server.New(cfg, storage, log)
	log.Info("Starting HTTP server", "port", cfg.Port)
	
	if err := srv.Start(); err != nil {
		log.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

func ensureDBDirectory(ctx context.Context, dbPath string) error {
	log := logger.FromContext(ctx)
	
	dir := filepath.Dir(dbPath)
	if dir == "." {
		log.Debug("Database in current directory, no directory creation needed")
		return nil
	}
	
	log.Info("Creating database directory", "directory", dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Error("Failed to create database directory", "directory", dir, "error", err)
		return err
	}
	
	log.Debug("Database directory created successfully", "directory", dir)
	return nil
}