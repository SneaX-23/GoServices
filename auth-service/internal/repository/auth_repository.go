package repository

import (
	"context"
	"fmt"

	"github.com/SneaX-23/GoServices/auth-service/internal/database"
)

type User struct {
	Id    int32
	Email string
	Name  string
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
}

type postgresUserRepository struct {
	db *database.Database
}

func NewUserRepository(db *database.Database) UserRepository {
	return &postgresUserRepository{
		db: db,
	}
}

func (r *postgresUserRepository) Create(ctx context.Context, user *User) error {
	query := `INSERT INTO users (email, name) values ($1, $2) returning id`

	err := r.db.Pool.QueryRow(ctx, query, user.Email, user.Name).Scan(&user.Id)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

func (r *postgresUserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, email, name FROM users WHERE email =$1`

	var user User
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(&user.Id, &user.Email, &user.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	return &user, nil
}
