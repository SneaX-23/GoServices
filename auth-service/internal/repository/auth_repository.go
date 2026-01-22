package repository

import (
	"context"
	"fmt"

	"github.com/SneaX-23/GoServices/auth-service/internal/database"
	"github.com/SneaX-23/GoServices/auth-service/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
}

type postgresUserRepository struct {
	db *database.Database
}

func NewUserRepository(db *database.Database) UserRepository {
	return &postgresUserRepository{
		db: db,
	}
}

func (r *postgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (email, name) values ($1, $2) returning id`

	err := r.db.Pool.QueryRow(ctx, query, user.Email, user.Username).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

func (r *postgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, name FROM users WHERE email =$1`

	var user domain.User
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	return &user, nil
}

func (r *postgresUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `SELECT * FROM users WHERE username = $1`

	var user domain.User
	err := r.db.Pool.QueryRow(ctx, query, username).Scan(&user.ID, &user.Email, &user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user by username: %w", err)
	}

	return &user, nil
}
