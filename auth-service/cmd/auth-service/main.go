package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/SneaX-23/GoServices/auth-service/internal/config"
	"github.com/SneaX-23/GoServices/auth-service/internal/database"
	"github.com/SneaX-23/GoServices/auth-service/internal/handlers"
	"github.com/SneaX-23/GoServices/auth-service/internal/messaging"
	"github.com/SneaX-23/GoServices/auth-service/internal/repository"
	"github.com/SneaX-23/GoServices/auth-service/internal/service"
	"github.com/go-playground/validator/v10"
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
	rdb := config.NewRedisClient()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// ping db to check connection
	var now time.Time
	err = db.Pool.QueryRow(ctx, "SELECT NOW()").Scan(&now)
	if err != nil {
		logger.Error("failed to execute test query", "error", err)
		os.Exit(1)
	}
	logger.Info("application started successfully", "db_time", now)

	// Initialize repository
	userRepo := repository.NewUserRepository(db)

	// Initialize redis
	redisOtpRepo := repository.NewRedisOtpRepo(rdb)

	// Initialize producer
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	producer := messaging.NewKafkaProducer(kafkaBroker, "auth-events")
	defer producer.Close()

	//
	service := service.NewAuthService(userRepo, redisOtpRepo, producer)

	// go-playground validator to validate user input
	v := validator.New()
	userhandler := handlers.NewUserHandler(service, v)

	// Setup routes
	mux := http.NewServeMux()

	mux.HandleFunc("POST /verify-email", userhandler.VerifyEmail)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	// Listen on port :8080
	slog.Info("Server starting on :8080")
	if err := server.ListenAndServe(); err != nil {
		slog.Error("Server failed to start", "err", err)
	}
}
