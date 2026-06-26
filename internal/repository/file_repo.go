package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Lorenta-Tech/kiosk-server/internal/models"
	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
	"github.com/lib/pq"
)

// DBTX
type DBTX interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

// Interface

type FileRepo interface {
	WithTx(tx *sql.Tx) FileRepo

	// Init
	CreateSession(ctx context.Context, session models.UploadSession) error
	CreateFiles(ctx context.Context, files []models.UploadFile) error

	// Confirm
	GetSessionByID(ctx context.Context, sessionID string) (models.UploadSession, error)
	GetFileByID(ctx context.Context, fileID, sessionID string) (models.UploadFile, error)
	UpdateFileWithPrintOptions(ctx context.Context, file models.UploadFile) error
	UpdateSessionPriced(ctx context.Context, sessionID string, totalAmount float64, totalSheets int) error

	// Recent print jobs
	GetRecentPrintJobs(ctx context.Context, userID string, limit int) ([]models.UploadSession, error)
	GetActivePrintJobs(ctx context.Context, userID string) ([]models.UploadSession, error)
	GetFilesBySessionID(ctx context.Context, sessionID string) ([]models.UploadFile, error)

	// Token lookup
	GetSessionByToken(ctx context.Context, token string) (models.UploadSession, error)
	MarkFilePromoted(ctx context.Context, fileID, finalKey string) error
	UpdateSessionPaid(ctx context.Context, sessionID string) error

	ExpireSessionAfterPrinting(ctx context.Context, sessionID string) error

	//admin 
	// NEED TO BE FIX
	AdminFetchPrintHistory(ctx context.Context) ([]models.PrintJob, error)
}

// Implementation
type PostgresFileRepo struct {
	db DBTX
}

func NewFileRepository(db DBTX) *PostgresFileRepo {
	return &PostgresFileRepo{db: db}
}

func (r *PostgresFileRepo) WithTx(tx *sql.Tx) FileRepo {
	return &PostgresFileRepo{db: tx}
}

// Init
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
			f.ID, f.SessionID, f.FileName, f.StagingKey, f.FileStatus,
		); err != nil {
			return apperror.Internal(
				"failed to save file record",
				fmt.Errorf("repository.CreateFiles file=%s: %w", f.FileName, err),
			)
		}
	}
	return nil
}

// Confirm
func (r *PostgresFileRepo) GetSessionByID(ctx context.Context, sessionID string) (models.UploadSession, error) {
	const query = `
		SELECT id, user_id, user_email, status, token, total_amount, expires_at, created_at
		FROM upload_sessions
		WHERE id = $1
	`
	var s models.UploadSession
	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&s.ID, &s.UserID, &s.UserEmail,
		&s.Status, &s.Token, &s.TotalAmount, &s.ExpiresAt, &s.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return models.UploadSession{}, apperror.NotFound(
			"session_not_found",
			fmt.Sprintf("session %s does not exist", sessionID),
		)
	}
	if err != nil {
		return models.UploadSession{}, apperror.Internal(
			"failed to fetch session",
			fmt.Errorf("repository.GetSessionByID: %w", err),
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
	var f models.UploadFile
	err := r.db.QueryRowContext(ctx, query, fileID, sessionID).Scan(
		&f.ID, &f.SessionID, &f.FileName, &f.StagingKey, &f.FileStatus,
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
			printing_mode   = $1,
			printing_side   = $2,
			page_range      = $3,
			page_layout     = $4,
			copies          = $5,
			number_of_pages = $6,
			price           = $7,
			file_status     = 'confirmed'
		WHERE id = $8 AND session_id = $9
	`

	fmt.Println("DEBUG: PageRange in UpdateFileWithPrintOptions:", f.PageRange)
	_, err := r.db.ExecContext(ctx, query,
		f.PrintingMode, f.PrintingSide,
		pq.Array(f.PageRange),
		f.PageLayout, f.Copies, f.NumberOfPages, f.Price,
		f.ID, f.SessionID,
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
		SET status = 'priced', total_amount = $1, total_sheets = $2
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

// Recent print jobs
func (r *PostgresFileRepo) GetRecentPrintJobs(ctx context.Context, userID string, limit int) ([]models.UploadSession, error) {
	const query = `
		SELECT id, user_id, user_email, status, total_amount, total_sheets, token, expires_at, created_at
		FROM upload_sessions
		WHERE user_id = $1
		  AND status = 'completed'
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, apperror.Internal(
			"failed to fetch recent print jobs",
			fmt.Errorf("repository.GetRecentPrintJobs: %w", err),
		)
	}
	defer rows.Close()

	sessions := make([]models.UploadSession, 0, limit)

	for rows.Next() {
		var s models.UploadSession

		if err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.UserEmail,
			&s.Status,
			&s.TotalAmount,
			&s.TotalSheets,
			&s.Token,
			&s.ExpiresAt,
			&s.CreatedAt,
		); err != nil {
			return nil, apperror.Internal(
				"failed to scan print job row",
				fmt.Errorf("repository.GetRecentPrintJobs scan: %w", err),
			)
		}

		sessions = append(sessions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, apperror.Internal(
			"error reading print job rows",
			fmt.Errorf("repository.GetRecentPrintJobs rows.Err: %w", err),
		)
	}

	return sessions, nil
}

func (r *PostgresFileRepo) GetFilesBySessionID(ctx context.Context, sessionID string) ([]models.UploadFile, error) {
	const query = `
		SELECT
			id, session_id, file_name,staging_key,final_key,
			printing_mode, printing_side, page_range,
			page_layout, copies, number_of_pages,
			price, file_status, created_at
		FROM upload_files
		WHERE session_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, apperror.Internal(
			"failed to fetch session files",
			fmt.Errorf("repository.GetFilesBySessionID: %w", err),
		)
	}
	defer rows.Close()

	files := make([]models.UploadFile, 0)
	for rows.Next() {
		var f models.UploadFile
		if err := rows.Scan(
			&f.ID, &f.SessionID, &f.FileName, &f.StagingKey, &f.FinalKey,
			&f.PrintingMode, &f.PrintingSide,
			pq.Array(&f.PageRange),
			&f.PageLayout, &f.Copies, &f.NumberOfPages,
			&f.Price, &f.FileStatus, &f.CreatedAt,
		); err != nil {
			return nil, apperror.Internal(
				"failed to scan file row",
				fmt.Errorf("repository.GetFilesBySessionID scan: %w", err),
			)
		}
		files = append(files, f)
	}
	if err := rows.Err(); err != nil {
		return nil, apperror.Internal(
			"error reading file rows",
			fmt.Errorf("repository.GetFilesBySessionID rows.Err: %w", err),
		)
	}
	return files, nil
}

// Token lookup
func (r *PostgresFileRepo) GetSessionByToken(ctx context.Context, token string) (models.UploadSession, error) {
	const query = `
		SELECT id, user_id, user_email, status, token, total_amount, total_sheets, expires_at, created_at
		FROM upload_sessions
		WHERE token = $1
	`
	var s models.UploadSession
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&s.ID,
		&s.UserID,
		&s.UserEmail,
		&s.Status,
		&s.Token,
		&s.TotalAmount,
		&s.TotalSheets,
		&s.ExpiresAt,
		&s.CreatedAt,
	)
	if err == sql.ErrNoRows {
		// Return a generic message — do not reveal whether the token
		// exists or the session is in a different state.
		return models.UploadSession{}, apperror.NotFound(
			"invalid_token",
			"the token you entered is invalid",
		)
	}
	if err != nil {
		return models.UploadSession{}, apperror.Internal(
			"failed to fetch session by token",
			fmt.Errorf("repository.GetSessionByToken: %w", err),
		)
	}
	return s, nil
}

// MarkFilePromoted sets final_key and file_status="promoted" after S3 promotion.
func (r *PostgresFileRepo) MarkFilePromoted(ctx context.Context, fileID, finalKey string) error {
	const query = `
		UPDATE upload_files
		SET final_key = $1, file_status = 'promoted'
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, finalKey, fileID)
	if err != nil {
		return apperror.Internal(
			"failed to mark file promoted",
			fmt.Errorf("repository.MarkFilePromoted file=%s: %w", fileID, err),
		)
	}
	return nil
}

// UpdateSessionPaid sets status="paid" on an upload_sessions row.
func (r *PostgresFileRepo) UpdateSessionPaid(ctx context.Context, sessionID string) error {
	const query = `
		UPDATE upload_sessions
		SET status = 'paid'
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return apperror.Internal(
			"failed to mark session paid",
			fmt.Errorf("repository.UpdateSessionPaid session=%s: %w", sessionID, err),
		)
	}
	return nil
}

func (r *PostgresFileRepo) ExpireSessionAfterPrinting(
	ctx context.Context,
	sessionID string,
) error {

	const query = `
		UPDATE upload_sessions
		SET
			status = 'completed',
			expires_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return apperror.Internal(
			"failed to expire session after printing",
			fmt.Errorf("repository.ExpireSessionAfterPrinting session=%s: %w", sessionID, err),
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperror.Internal(
			"failed to verify session update",
			fmt.Errorf("repository.ExpireSessionAfterPrinting rows: %w", err),
		)
	}

	if rowsAffected == 0 {
		return apperror.NotFound(
			"session_not_found",
			fmt.Sprintf("session %s does not exist", sessionID),
		)
	}

	return nil
}

// Active print jobs (status = paid)
func (r *PostgresFileRepo) GetActivePrintJobs(
	ctx context.Context,
	userID string,
) ([]models.UploadSession, error) {

	const query = `
		SELECT id, user_id, user_email, status,
		       total_amount, total_sheets, token,
		       expires_at, created_at
		FROM upload_sessions
		WHERE user_id = $1
		  AND status = 'paid'
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, apperror.Internal(
			"failed to fetch active print jobs",
			fmt.Errorf("repository.GetActivePrintJobs: %w", err),
		)
	}
	defer rows.Close()

	sessions := make([]models.UploadSession, 0)

	for rows.Next() {
		var s models.UploadSession

		if err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.UserEmail,
			&s.Status,
			&s.TotalAmount,
			&s.TotalSheets,
			&s.Token,
			&s.ExpiresAt,
			&s.CreatedAt,
		); err != nil {
			return nil, apperror.Internal(
				"failed to scan active print job row",
				fmt.Errorf("repository.GetActivePrintJobs scan: %w", err),
			)
		}

		sessions = append(sessions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, apperror.Internal(
			"error reading active print job rows",
			fmt.Errorf("repository.GetActivePrintJobs rows.Err: %w", err),
		)
	}

	return sessions, nil
}
func (r *PostgresFileRepo) AdminFetchPrintHistory(ctx context.Context) ([]models.PrintJob, error) {
	const sessionsQuery = `
		SELECT id, token, status, total_amount, total_sheets, created_at
		FROM upload_sessions
		WHERE status = 'completed'
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, sessionsQuery)
	if err != nil {
		return nil, apperror.Internal(
			"failed to fetch print history",
			fmt.Errorf("repository.AdminFetchPrintHistory: %w", err),
		)
	}
	defer rows.Close()

	jobs := make([]models.PrintJob, 0)
	jobIndex := make(map[string]int) // session_id -> index in jobs, so we can attach files after

	for rows.Next() {
		var job models.PrintJob

		if err := rows.Scan(
			&job.SessionID,
			&job.Token,
			&job.Status,
			&job.TotalAmount,
			&job.TotalSheets,
			&job.CreatedAt,
		); err != nil {
			return nil, apperror.Internal(
				"failed to scan print history row",
				fmt.Errorf("repository.AdminFetchPrintHistory scan: %w", err),
			)
		}

		job.Files = make([]models.PrintJobFile, 0)
		jobIndex[job.SessionID] = len(jobs)
		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, apperror.Internal(
			"error reading print history rows",
			fmt.Errorf("repository.AdminFetchPrintHistory rows.Err: %w", err),
		)
	}

	if len(jobs) == 0 {
		return jobs, nil
	}

	// Collect session IDs to fetch files in one query (avoid N+1)
	sessionIDs := make([]string, 0, len(jobs))
	for _, j := range jobs {
		sessionIDs = append(sessionIDs, j.SessionID)
	}

	const filesQuery = `
		SELECT session_id, id, file_name, printing_mode, printing_side,
		       page_range, page_layout, copies, number_of_pages, price, file_status
		FROM upload_files
		WHERE session_id = ANY($1)
		ORDER BY session_id, file_name
	`

	fileRows, err := r.db.QueryContext(ctx, filesQuery, sessionIDs)
	if err != nil {
		return nil, apperror.Internal(
			"failed to fetch print job files",
			fmt.Errorf("repository.AdminFetchPrintHistory files: %w", err),
		)
	}
	defer fileRows.Close()

	for fileRows.Next() {
		var sessionID string
		var f models.PrintJobFile

		if err := fileRows.Scan(
			&sessionID,
			&f.FileID,
			&f.FileName,
			&f.PrintingMode,
			&f.PrintingSide,
			pq.Array(&f.PageRange),
			&f.PageLayout,
			&f.Copies,
			&f.NumberOfPages,
			&f.Price,
			&f.FileStatus,
		); err != nil {
			return nil, apperror.Internal(
				"failed to scan print job file row",
				fmt.Errorf("repository.AdminFetchPrintHistory file scan: %w", err),
			)
		}

		if idx, ok := jobIndex[sessionID]; ok {
			jobs[idx].Files = append(jobs[idx].Files, f)
		}
	}

	if err := fileRows.Err(); err != nil {
		return nil, apperror.Internal(
			"error reading print job file rows",
			fmt.Errorf("repository.AdminFetchPrintHistory fileRows.Err: %w", err),
		)
	}

	return jobs, nil
}