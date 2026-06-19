package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Lorenta-Tech/kiosk-server/internal/models"
	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
)

type UserRepo interface {
	GetByGoogleID(ctx context.Context, googleID string) (models.User, error)
	Create(ctx context.Context, user models.User) (models.User, error)
	CheckUserExists(ctx context.Context, userID string) (bool,error)
}

type PostgresUserRepo struct {
	db DBTX
}

func NewUserRepository(db DBTX) *PostgresUserRepo {
	return &PostgresUserRepo{db: db}
}

func (r *PostgresUserRepo) GetByGoogleID(ctx context.Context, googleID string) (models.User, error) {
	const query = `
		SELECT id, google_id, email, name, created_at, updated_at
		FROM users
		WHERE google_id = $1
	`
	var u models.User
	err := r.db.QueryRowContext(ctx, query, googleID).Scan(
		&u.ID,
		&u.GoogleID,
		&u.Email,
		&u.Name,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return models.User{}, apperror.NotFound(
			"user_not_found",
			"user not found",
		)
	}
	if err != nil {
		return models.User{}, apperror.Internal(
			"failed to fetch user",
			fmt.Errorf("repository.GetByGoogleID: %w", err),
		)
	}
	return u, nil
}

func (r *PostgresUserRepo) Create(ctx context.Context, user models.User) (models.User, error) {
	const query = `
		INSERT INTO users (id, google_id, email, name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, google_id, email, name, created_at, updated_at
	`
	var u models.User
	err := r.db.QueryRowContext(ctx, query,
		user.ID,
		user.GoogleID,
		user.Email,
		user.Name,
	).Scan(
		&u.ID,
		&u.GoogleID,
		&u.Email,
		&u.Name,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return models.User{}, apperror.Internal(
			"failed to create user",
			fmt.Errorf("repository.Create: %w", err),
		)
	}
	return u, nil
}

func (r *PostgresUserRepo) CheckUserExists(ctx context.Context, userID string) (bool,error) {
	const query = `
		SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)
	`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&exists)
	if err != nil {
		return false, apperror.Internal(
			"failed to check user existence",
			fmt.Errorf("repository.CheckUserExists: %w", err),
		)
	}
	return exists, nil
}