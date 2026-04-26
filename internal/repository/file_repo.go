package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Lorenta-Tech/kiosk-server/internal/models"
	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
	"github.com/lib/pq"
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
	WithTx(tx *sql.Tx) FileRepo
	CreateSession(ctx context.Context, session models.UploadSession) error
	CreateFiles(ctx context.Context, files []models.UploadFile) error
	GetSessionByID(ctx context.Context, sessionID string) (models.UploadSession, error)
	GetFileByID(ctx context.Context, fileID, sessionID string) (models.UploadFile, error)
	UpdateFileWithPrintOptions(ctx context.Context, file models.UploadFile) error
	UpdateSessionPriced(ctx context.Context, sessionID string, totalAmount float64, totalSheets int) error
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

func (r *PostgresFileRepo) GetSessionByID(ctx context.Context, sessionID string) (models.UploadSession, error) {
	const query = `
	    SELECT id, user_id, user_email, status, token, expires_at,created_at
		FROM upload_sessions
		WHERE id = $1
		`

	row := r.db.QueryRowContext(ctx, query, sessionID)

	var s models.UploadSession

	err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.UserEmail,
		&s.Status,
		&s.Token,
		&s.ExpiresAt,
		&s.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return models.UploadSession{}, apperror.NotFound(
			"session_not_found",
			fmt.Sprintf("session %s does not exist", sessionID),
		)
	}
	return s, nil
}

func (r *PostgresFileRepo) GetFileByID(ctx context.Context, fileID, sessionID string) (models.UploadFile, error) {
	const query = `
		SELECT id, session_id, file_name, staging_key, file_status
		FROM upload_files
		WHERE id = $1 AND session_id = $2
	`
	row := r.db.QueryRowContext(ctx, query, fileID, sessionID)
 
	var f models.UploadFile
	err := row.Scan(
		&f.ID,
		&f.SessionID,
		&f.FileName,
		&f.StagingKey,
		&f.FileStatus,
	)
	if err == sql.ErrNoRows {
		return models.UploadFile{}, apperror.NotFound(
			"file_not_found",
			fmt.Sprintf("file %s not found in session %s", fileID, sessionID),
		)
	}
	if err != nil {
		return models.UploadFile{}, apperror.Internal(
			"failed to fetch file",
			fmt.Errorf("repository.GetFileByID: %w", err),
		)
	}
	return f, nil
}

func (r *PostgresFileRepo) UpdateFileWithPrintOptions(ctx context.Context, f models.UploadFile) error {
	const query = `
		UPDATE upload_files
		SET
			printing_mode  = $1,
			printing_side  = $2,
			page_range     = $3,
			page_layout    = $4,
			copies         = $5,
			number_of_pages = $6,
			price          = $7,
			file_status    = 'confirmed'
		WHERE id = $8 AND session_id = $9
	`
	_, err := r.db.ExecContext(ctx, query,
		f.PrintingMode,
		f.PrintingSide,
		pq.Array(f.PageRange),
		f.PageLayout,
		f.Copies,
		f.NumberOfPages,
		f.Price,
		f.ID,
		f.SessionID,
	)
	if err != nil {
		return apperror.Internal(
			"failed to update file print options",
			fmt.Errorf("repository.UpdateFileWithPrintOptions file=%s: %w", f.ID, err),
		)
	}
	return nil
}

func (r *PostgresFileRepo) UpdateSessionPriced(ctx context.Context, sessionID string, totalAmount float64, totalSheets int) error {
	const query = `
		UPDATE upload_sessions
		SET
			status       = 'priced',
			total_amount = $1,
			total_sheets = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, totalAmount, totalSheets, sessionID)
	if err != nil {
		return apperror.Internal(
			"failed to update session status",
			fmt.Errorf("repository.UpdateSessionPriced session=%s: %w", sessionID, err),
		)
	}
	return nil
}

func (r *PostgresFileRepo) GetFileByToken(ctx context.Context,token string)(error){
	return nil
}

