package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/SneaX-23/GoServices/internal/config"
	"github.com/SneaX-23/GoServices/internal/database"
)

func main() {
	var handler slog.Handler
	if os.Getenv("APP_ENV") == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	logger := slog.New(handler)
	slog.SetDefault(logger) // Set as global logger

	dbCfg := config.LoadDatabaseConfig()

	db, err := database.New(dbCfg, logger)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Ensure the pool is closed when the application shuts down
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var now time.Time
	err = db.Pool.QueryRow(ctx, "SELECT NOW()").Scan(&now)
	if err != nil {
		logger.Error("failed to execute test query", "error", err)
		os.Exit(1)
	}

	logger.Info("application started successfully", "db_time", now)
}
