package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/SneaX-23/GoServices/auth-service/internal/database"
	"github.com/SneaX-23/GoServices/auth-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	StoreRefreshToken(ctx context.Context, userID, hashedToken string) error
	FindTokenByHash(ctx context.Context, hashedToken string) (*domain.ExistingRefreshToken, error)
	RevokeAllTokens(ctx context.Context, userID string) error
	DeleteToken(ctx context.Context, id string) error
	RotateRefreshToken(ctx context.Context, userID, oldTokenID, newHashedToken string) error
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

func (r *postgresUserRepository) StoreRefreshToken(ctx context.Context, userID, hashedToken string) error {
	ctx, span := r.tracer.Start(ctx, "repository.StoreRefreshToken")
	defer span.End()

	span.SetAttributes(attribute.String("user.userID", userID))

	query := `INSERT into refresh_tokens (hashed_token, user_id, expires_at, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`
	expires_at := time.Now().Add(7 * 24 * time.Hour)

	_, err := r.db.Pool.Exec(ctx, query, hashedToken, userID, expires_at, time.Now(), time.Now())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to store hash refresh token")
		return fmt.Errorf("failed to store hashed refresh token: %w", err)
	}
	span.SetStatus(codes.Ok, "token stored")
	return nil
}

func (r *postgresUserRepository) FindTokenByHash(ctx context.Context, hashedToken string) (*domain.ExistingRefreshToken, error) {
	ctx, span := r.tracer.Start(ctx, "repository.FindTokenByHash")
	defer span.End()

	var existingToken domain.ExistingRefreshToken

	query := `SELECT id, hashed_token, user_id, replaced_by, expires_at, created_at, updated_at
			  FROM refresh_tokens 
			  WHERE hashed_token = $1`

	err := r.db.Pool.QueryRow(ctx, query, hashedToken).
		Scan(&existingToken.ID,
			&existingToken.HashedToken,
			&existingToken.UserID,
			&existingToken.ReplacedBy,
			&existingToken.ExpiresAt,
			&existingToken.CreatedAt,
			&existingToken.UpdatedAt,
		)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			span.SetStatus(codes.Ok, "token not found")
			return nil, nil
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to find token by hash: %w", err)
	}
	span.SetStatus(codes.Ok, "refresh token found")

	return &existingToken, nil
}

func (r *postgresUserRepository) RevokeAllTokens(ctx context.Context, userID string) error {
	ctx, span := r.tracer.Start(ctx, "repository.RevokeAllTokens")
	defer span.End()

	span.SetAttributes(attribute.String("user.userID", userID))

	query := `DELETE FROM refresh_tokens WHERE user_id = $1`

	_, err := r.db.Pool.Exec(ctx, query, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "couldnt revoke tokens")
		return fmt.Errorf("failed to revoke user tokens: %w", err)
	}

	span.SetStatus(codes.Ok, "revoked user tokens")

	return nil
}

func (r *postgresUserRepository) DeleteToken(ctx context.Context, id string) error {
	ctx, span := r.tracer.Start(ctx, "repository.DeleteToken")
	defer span.End()
	span.SetAttributes(attribute.String("refresh_tokens.id", id))

	query := `DELETE FROM refresh_tokens WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete token")
		return fmt.Errorf("failed to delete token: %w", err)
	}
	span.SetStatus(codes.Ok, "token deleted")
	return nil
}

func (r *postgresUserRepository) RotateRefreshToken(ctx context.Context, userID, oldTokenID, newHashedToken string) error {
	ctx, span := r.tracer.Start(ctx, "repository.RotateRefreshToken")
	defer span.End()

	span.SetAttributes(attribute.String("refresh_tokens.user_id", userID))

	tx, err := r.db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	insertQuery := `
    INSERT INTO refresh_tokens
        (hashed_token, user_id, expires_at, created_at, updated_at)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING id
    `

	var newTokenID string
	if err := tx.QueryRow(
		ctx,
		insertQuery,
		newHashedToken,
		userID,
		expiresAt,
		time.Now(),
		time.Now(),
	).Scan(&newTokenID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to insert new token: %w", err)
	}

	updateQuery := `
    UPDATE refresh_tokens
    SET replaced_by = $1
    WHERE id = $2 AND user_id = $3
    `

	cmd, err := tx.Exec(ctx, updateQuery, newTokenID, oldTokenID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to update old token: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("old refresh token not found or unauthorized")
	}

	if err := tx.Commit(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	span.SetStatus(codes.Ok, "refresh token rotated")
	return nil
}
