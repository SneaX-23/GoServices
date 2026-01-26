package repository

import (
	"context"
	"fmt"

	"github.com/SneaX-23/GoServices/auth-service/internal/database"
	"github.com/SneaX-23/GoServices/auth-service/internal/domain"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
}

type postgresUserRepository struct {
	db     *database.Database
	tracer trace.Tracer
}

func NewUserRepository(db *database.Database, tracer trace.Tracer) UserRepository {
	return &postgresUserRepository{
		db:     db,
		tracer: tracer,
	}
}

func (r *postgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	ctx, span := r.tracer.Start(ctx, "repository.Create")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.eamil", user.Email),
		attribute.String("user.username", user.Username),
	)

	query := `INSERT INTO users (email, name, password) values ($1, $2, $3) returning id`

	err := r.db.Pool.QueryRow(ctx, query, user.Email, user.Username, user.Password).Scan(&user.ID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to innsert user")
		return fmt.Errorf("failed to insert user: %w", err)
	}

	span.SetAttributes(
		attribute.String("user.ID", user.ID),
	)
	span.SetStatus(codes.Ok, "user created successfully")

	return nil
}

func (r *postgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.GetByEmail")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	query := `SELECT id, email, name FROM users WHERE email =$1`

	var user domain.User
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.Username)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch user")
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	span.SetStatus(codes.Ok, "user found")
	return &user, nil
}

func (r *postgresUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.GetByUsername")
	defer span.End()

	span.SetAttributes(attribute.String("usre.username", username))

	query := `SELECT * FROM users WHERE username = $1`

	var user domain.User
	err := r.db.Pool.QueryRow(ctx, query, username).Scan(&user.ID, &user.Email, &user.Username)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch user by username")
		return nil, fmt.Errorf("failed to fetch user by username: %w", err)
	}

	span.SetStatus(codes.Ok, "user found")
	return &user, nil
}
