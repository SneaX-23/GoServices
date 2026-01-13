package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/SneaX-23/GoServices/auth-service/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

type SlogAdapter struct {
	l *slog.Logger
}

func (s *SlogAdapter) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	attrs := make([]slog.Attr, 0, len(data))
	for k, v := range data {
		attrs = append(attrs, slog.Any(k, v))
	}

	var slogLevel slog.Level
	switch level {
	case tracelog.LogLevelTrace, tracelog.LogLevelDebug:
		slogLevel = slog.LevelDebug
	case tracelog.LogLevelInfo:
		slogLevel = slog.LevelInfo
	case tracelog.LogLevelWarn:
		slogLevel = slog.LevelWarn
	case tracelog.LogLevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	s.l.LogAttrs(ctx, slogLevel, msg, attrs...)
}

type Database struct {
	Pool *pgxpool.Pool
}

func New(cfg *config.DatabaseConfig, logger *slog.Logger) (*Database, error) {
	// Build Connection String
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode)

	// Parse Config
	pgxConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx config: %w", err)
	}

	// Set Production Pool Limits
	pgxConfig.MaxConns = cfg.MaxConns
	pgxConfig.MinConns = cfg.MinConns
	pgxConfig.MaxConnLifetime = cfg.MaxConnLifetime
	pgxConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	pgxConfig.HealthCheckPeriod = cfg.HealthCheckPeriod

	// Attach simple slog tracer
	pgxConfig.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   &SlogAdapter{l: logger},
		LogLevel: tracelog.LogLevelInfo, // Set to LogLevelDebug to see all SQL queries
	}

	// Create the Pool
	pool, err := pgxpool.NewWithConfig(context.Background(), pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Health Check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db ping failed: %w", err)
	}

	logger.Info("database connection established", "host", cfg.Host, "db", cfg.Name)

	return &Database{Pool: pool}, nil
}

func (db *Database) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
