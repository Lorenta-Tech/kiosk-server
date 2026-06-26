package repository

import (
	"context"
	"fmt"

	"github.com/Lorenta-Tech/kiosk-server/internal/models"
	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
	"github.com/lib/pq"
)

type PostgresAdminRepo struct {
	db DBTX
}

func NewAdminRepository(db DBTX) *PostgresAdminRepo {
	return &PostgresAdminRepo{db: db}
}

type AdminRepo interface {
	AdminFetchPrintHistory(ctx context.Context) ([]models.PrintJob, error)
	AdminGetTotalRevenue(ctx context.Context) (float64, error)
	AdminGetTotalSheetsPrinted(ctx context.Context) (int, error)
	AdminGetTotalColorSheetsPrinted(ctx context.Context) (int, error)
	AdminGetTotalBlackAndWhiteSheetsPrinted(ctx context.Context) (int, error)
}

func (r *PostgresAdminRepo) AdminFetchPrintHistory(ctx context.Context) ([]models.PrintJob, error) {
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

func (r *PostgresAdminRepo) AdminGetTotalRevenue(ctx context.Context) (float64, error) {
	const query = `
		SELECT COALESCE(SUM(total_amount), 0)
		FROM upload_sessions
		WHERE status = 'completed'
	`

	var total float64
	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, apperror.Internal(
			"failed to calculate total revenue",
			fmt.Errorf("repository.AdminGetTotalRevenue: %w", err),
		)
	}

	return total, nil
}
func (r *PostgresAdminRepo) AdminGetTotalSheetsPrinted(ctx context.Context) (int, error) {
	const query = `
		SELECT COALESCE(SUM(uf.number_of_pages), 0)
		FROM upload_files uf
		JOIN upload_sessions us ON us.id = uf.session_id
		WHERE us.status = 'completed'
	`
	var total int
	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, apperror.Internal(
			"failed to calculate total sheets printed",
			fmt.Errorf("repository.AdminGetTotalSheetsPrinted: %w", err),
		)
	}
	return total, nil
}

func (r *PostgresAdminRepo) AdminGetTotalColorSheetsPrinted(ctx context.Context) (int, error) {
	const query = `
		SELECT COALESCE(SUM(uf.number_of_pages), 0)
		FROM upload_files uf
		JOIN upload_sessions us ON us.id = uf.session_id
		WHERE us.status = 'completed'
		  AND uf.printing_mode = 'color'
	`
	var total int
	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, apperror.Internal(
			"failed to calculate total color sheets printed",
			fmt.Errorf("repository.AdminGetTotalColorSheetsPrinted: %w", err),
		)
	}
	return total, nil
}

func (r *PostgresAdminRepo) AdminGetTotalBlackAndWhiteSheetsPrinted(ctx context.Context) (int, error) {
	const query = `
		SELECT COALESCE(SUM(uf.number_of_pages), 0)
		FROM upload_files uf
		JOIN upload_sessions us ON us.id = uf.session_id
		WHERE us.status = 'completed'
		  AND uf.printing_mode = 'monochromatic'
	`
	var total int
	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, apperror.Internal(
			"failed to calculate total black and white sheets printed",
			fmt.Errorf("repository.AdminGetTotalBlackAndWhiteSheetsPrinted: %w", err),
		)
	}
	return total, nil
}