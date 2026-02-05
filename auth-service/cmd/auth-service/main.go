package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/SneaX-23/GoServices/auth-service/internal/config"
	"github.com/SneaX-23/GoServices/auth-service/internal/database"
	"github.com/SneaX-23/GoServices/auth-service/internal/genproto"
	"github.com/SneaX-23/GoServices/auth-service/internal/handlers"
	"github.com/SneaX-23/GoServices/auth-service/internal/messaging"
	"github.com/SneaX-23/GoServices/auth-service/internal/repository"
	"github.com/SneaX-23/GoServices/auth-service/internal/service"
	"github.com/SneaX-23/GoServices/auth-service/internal/telemetry"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
)

func main() {
	var handler slog.Handler
	if os.Getenv("APP_ENV") == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Initialize OpenTelemetry
	telemetryCfg := telemetry.LoadTelemetryConfig()
	shutdown, err := telemetry.InitTracer(telemetryCfg)
	if err != nil {
		logger.Error("failed to initialize tracer", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			logger.Error("failed to shutdown tracer", "error", err)
		}
	}()

	tracer := telemetry.GetTracer()

	// Load database configuration
	dbCfg := config.LoadDatabaseConfig()

	db, err := database.New(dbCfg, logger, tracer)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize Redis
	rdb := config.NewRedisClient()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Ping database to check connection
	var now time.Time
	err = db.Pool.QueryRow(ctx, "SELECT NOW()").Scan(&now)
	if err != nil {
		logger.Error("failed to execute test query", "error", err)
		os.Exit(1)
	}
	logger.Info("application started successfully", "db_time", now)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db, tracer)
	redisOtpRepo := repository.NewRedisOtpRepo(rdb, tracer)

	// Initialize producer
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	producer := messaging.NewKafkaProducer(kafkaBroker, "auth-events", tracer)
	defer producer.Close()

	secret := os.Getenv("JWT_ACCESS_SECRET")
	if secret == "" {
		slog.Error("JWT_ACCESS_SECRET not set", "secret", secret)
		os.Exit(1)
	}
	// Initialize service
	authService := service.NewAuthService(userRepo, redisOtpRepo, producer, tracer, secret)

	v := validator.New()

	port, _ := strconv.Atoi(os.Getenv("PORT"))
	server := handlers.NewUserHandler(authService, v, tracer, port)

	// servet for grpc server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal("failed to to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	grpcHandlerServer := handlers.NewGRPCServer(authService)
	genproto.RegisterAuthServiceServer(grpcServer, grpcHandlerServer)

	go func() {
		slog.Info("gRPC server starting on :50001")
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("gRPC failed to serve", "err", err)
		}
	}()

	// Start server in goroutine
	go func() {
		slog.Info("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "err", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "err", err)
	}

	slog.Info("Server exited")
}
