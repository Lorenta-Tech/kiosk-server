package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Lorenta-Tech/kiosk-server/internal/models"
	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
)

// DBTX Initialization

// DBTX is satisfied by both *sql.DB and *sql.Tx.
// This means the same repository methods work for regular queries
// and transactional queries without any code duplication.
type DBTX interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type FileRepo interface {
	CreateSession(ctx context.Context, session models.UploadSession) error
	CreateFiles(ctx context.Context, files []models.UploadFile) error

	// WithTx returns a new repo instance that runs queries on the given
	// transaction. Use this when the service layer needs to coordinate
	// multiple repository calls inside one transaction.
	WithTx(tx *sql.Tx) FileRepo
}


type PostgresFileRepo struct {
	db DBTX
}

func NewFileRepository(db DBTX) *PostgresFileRepo {
	return &PostgresFileRepo{db: db}
}

func (r *PostgresFileRepo) WithTx(tx *sql.Tx) FileRepo {
	return &PostgresFileRepo{db: tx}
}

func (r *PostgresFileRepo) CreateSession(ctx context.Context, session models.UploadSession) error {
	const query = `
		INSERT INTO upload_sessions (id, user_id, user_email, status, token, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.UserEmail,
		session.Status,
		session.Token,
		session.ExpiresAt,
	)
	if err != nil {
		return apperror.Internal(
			"failed to create upload session",
			fmt.Errorf("repository.CreateSession: %w", err),
		)
	}
	return nil
}

func (r *PostgresFileRepo) CreateFiles(ctx context.Context, files []models.UploadFile) error {
	const query = `
		INSERT INTO upload_files (id, session_id, file_name, staging_key, file_status)
		VALUES ($1, $2, $3, $4, $5)
	`
	for _, f := range files {
		if _, err := r.db.ExecContext(ctx, query,
			f.ID,
			f.SessionID,
			f.FileName,
			f.StagingKey,
			f.FileStatus,
		); err != nil {
			return apperror.Internal(
				"failed to save file record",
				fmt.Errorf("repository.CreateFiles file=%s: %w", f.FileName, err),
			)
		}
	}
	return nil
}