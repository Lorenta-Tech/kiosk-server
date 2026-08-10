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
	AdminGetRevenueFromDouble_Sided_Prints(ctx context.Context) (int, error)
	AdminGetRevenueFromSingle_Sided_Prints(ctx context.Context) (int, error)
	AdminGetTotalSheetsPrintedByDouble_Sided_Prints(ctx context.Context) (int, error)
	AdminGetTotalSheetsPrintedBySingle_Sided_Prints(ctx context.Context) (int, error)
	AdminGetRevenueLast24Hours(ctx context.Context) (float64, error)
	AdminGetSheetsPrintedLast24Hours(ctx context.Context) (int, error)
	AdminGetColorSheetsPrintedLast24Hours(ctx context.Context) (int, error)
	AdminGetBlackAndWhiteSheetsPrintedLast24Hours(ctx context.Context) (int, error)
	AdminGetTotalSheetsPrintedInLast24HoursByDouble_Sided_Prints(ctx context.Context) (int, error)
	AdminGetTotalSheetsPrintedInLast24HoursBySingle_Sided_Prints(ctx context.Context) (int, error)
	AdminGetLast24HoursRevenueFromDouble_Sided_Prints(ctx context.Context) (int, error)
	AdminGetLast24HoursRevenueFromSingle_Sided_Prints(ctx context.Context) (int, error)
	AdminFetchPrintHistoryfor24H(ctx context.Context) ([]models.PrintJob, error)
	AdminFetchPrintJobsOnlyPaidInLast24H(ctx context.Context) ([]models.PrintJob, error)
	AdminGetTotalSessionCountOfToday(ctx context.Context) (int, error)
	AdminGetTodaysPaymentHistory(ctx context.Context) ([]models.Payment, error)
	AdminGetPrintJobsForPricedStatus(ctx context.Context) ([]models.PrintJob, error)
	AdminGetRevenueForDate(ctx context.Context, date string) (float64, error)
	AdminGetTotalSheetsPrintedForDate(ctx context.Context, date string) (int, error)
	AdminFetchPrintHistoryForDate(ctx context.Context, date string) ([]models.PrintJob, error)
	//AdminGetTotalSheestPrintedInLast24Hours(ctx context.Context) (int, error)
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

func (r *PostgresAdminRepo) AdminFetchPrintHistoryForDate(ctx context.Context, date string) ([]models.PrintJob, error) {
	const sessionsQuery = `
		SELECT id, token, status, total_amount, total_sheets, created_at
		FROM upload_sessions
		WHERE status = 'completed'
		  AND (created_at AT TIME ZONE 'Asia/Kolkata')::date = $1::date
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, sessionsQuery, date)
	if err != nil {
		return nil, apperror.Internal(
			"failed to fetch print history for date",
			fmt.Errorf("repository.AdminFetchPrintHistoryForDate: %w", err),
		)
	}
	defer rows.Close()

	jobs := make([]models.PrintJob, 0)
	jobIndex := make(map[string]int)
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
				fmt.Errorf("repository.AdminFetchPrintHistoryForDate scan: %w", err),
			)
		}
		job.Files = make([]models.PrintJobFile, 0)
		jobIndex[job.SessionID] = len(jobs)
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, apperror.Internal(
			"error reading print history rows",
			fmt.Errorf("repository.AdminFetchPrintHistoryForDate rows.Err: %w", err),
		)
	}
	if len(jobs) == 0 {
		return jobs, nil
	}

	sessionIDs := make([]string, 0, len(jobs))
	for _, job := range jobs {
		sessionIDs = append(sessionIDs, job.SessionID)
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
			fmt.Errorf("repository.AdminFetchPrintHistoryForDate files: %w", err),
		)
	}
	defer fileRows.Close()
	for fileRows.Next() {
		var sessionID string
		var file models.PrintJobFile
		if err := fileRows.Scan(
			&sessionID,
			&file.FileID,
			&file.FileName,
			&file.PrintingMode,
			&file.PrintingSide,
			pq.Array(&file.PageRange),
			&file.PageLayout,
			&file.Copies,
			&file.NumberOfPages,
			&file.Price,
			&file.FileStatus,
		); err != nil {
			return nil, apperror.Internal(
				"failed to scan print job file row",
				fmt.Errorf("repository.AdminFetchPrintHistoryForDate file scan: %w", err),
			)
		}
		if index, ok := jobIndex[sessionID]; ok {
			jobs[index].Files = append(jobs[index].Files, file)
		}
	}
	if err := fileRows.Err(); err != nil {
		return nil, apperror.Internal(
			"error reading print job file rows",
			fmt.Errorf("repository.AdminFetchPrintHistoryForDate fileRows.Err: %w", err),
		)
	}

	return jobs, nil
}

func (r *PostgresAdminRepo) AdminGetTotalRevenue(ctx context.Context) (float64, error) {
	const query = `
		SELECT COALESCE(SUM(total_amount), 0) AS total_amount
        FROM upload_sessions
        WHERE status IN ('completed', 'paid')
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

func (r *PostgresAdminRepo) AdminGetRevenueForDate(ctx context.Context, date string) (float64, error) {
	const query = `
		SELECT COALESCE(SUM(total_amount), 0)
		FROM upload_sessions
		WHERE status IN ('completed', 'paid')
		  AND (created_at AT TIME ZONE 'Asia/Kolkata')::date = $1::date
	`

	var total float64
	if err := r.db.QueryRowContext(ctx, query, date).Scan(&total); err != nil {
		return 0, apperror.Internal(
			"failed to calculate revenue for date",
			fmt.Errorf("repository.AdminGetRevenueForDate: %w", err),
		)
	}

	return total, nil
}

func (r *PostgresAdminRepo) AdminGetTotalSheetsPrinted(ctx context.Context) (int, error) {
	const query = `
SELECT COALESCE(SUM(total_sheets), 0) AS total_sheets_printed
FROM upload_sessions
WHERE status IN ('completed')
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

func (r *PostgresAdminRepo) AdminGetTotalSheetsPrintedForDate(ctx context.Context, date string) (int, error) {
	const query = `
		SELECT COALESCE(SUM(total_sheets), 0)
		FROM upload_sessions
		WHERE status = 'completed'
		  AND (created_at AT TIME ZONE 'Asia/Kolkata')::date = $1::date
	`

	var total int
	if err := r.db.QueryRowContext(ctx, query, date).Scan(&total); err != nil {
		return 0, apperror.Internal(
			"failed to calculate sheets printed for date",
			fmt.Errorf("repository.AdminGetTotalSheetsPrintedForDate: %w", err),
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

// Not Working as expected, need to check the query
func (r *PostgresAdminRepo) AdminGetColorSheetsPrintedLast24Hours(ctx context.Context) (int, error) {
	const query = `
		SELECT COALESCE(SUM(uf.number_of_pages), 0)
		FROM upload_files uf
		JOIN upload_sessions us ON us.id = uf.session_id
		WHERE us.status = 'completed'
		  AND uf.printing_mode = 'color'
		    AND (us.created_at AT TIME ZONE 'Asia/Kolkata')::date =
            (NOW() AT TIME ZONE 'Asia/Kolkata')::date
	`

	var total int
	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, apperror.Internal(
			"failed to calculate color sheets printed for last 24 hours",
			fmt.Errorf("repository.AdminGetColorSheetsPrintedLast24Hours: %w", err),
		)
	}

	return total, nil
}

// Not Working as expected, need to check the query
func (r *PostgresAdminRepo) AdminGetBlackAndWhiteSheetsPrintedLast24Hours(ctx context.Context) (int, error) {
	const query = `
		SELECT COALESCE(SUM(uf.number_of_pages), 0)
		FROM upload_files uf
		JOIN upload_sessions us ON us.id = uf.session_id
		WHERE us.status = 'completed'
		  AND uf.printing_mode = 'monochromatic'
		    AND (uf.created_at AT TIME ZONE 'Asia/Kolkata')::date =
            (NOW() AT TIME ZONE 'Asia/Kolkata')::date
	`

	var total int
	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, apperror.Internal(
			"failed to calculate black and white sheets printed for last 24 hours",
			fmt.Errorf("repository.AdminGetBlackAndWhiteSheetsPrintedLast24Hours: %w", err),
		)
	}

	return total, nil
}

func (r *PostgresAdminRepo) AdminGetRevenueLast24Hours(ctx context.Context) (float64, error) {
	const query = `
		SELECT COALESCE(SUM(total_amount), 0) AS total_amount
        FROM upload_sessions
        WHERE status IN ('completed', 'paid')
        AND (created_at AT TIME ZONE 'Asia/Kolkata')::date =
        (NOW() AT TIME ZONE 'Asia/Kolkata')::date
	`

	var total float64
	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, apperror.Internal(
			"failed to calculate revenue for last 24 hours",
			fmt.Errorf("repository.AdminGetRevenueLast24Hours: %w", err),
		)
	}

	return total, nil
}

func (r *PostgresAdminRepo) AdminGetSheetsPrintedLast24Hours(ctx context.Context) (int, error) {
	const query = `
		SELECT COALESCE(SUM(total_sheets), 0) AS total_sheets
        FROM upload_sessions
        WHERE status IN ('completed')
        AND (created_at AT TIME ZONE 'Asia/Kolkata')::date =
        (NOW() AT TIME ZONE 'Asia/Kolkata')::date
	`

	var total int
	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, apperror.Internal(
			"failed to calculate sheets printed for last 24 hours",
			fmt.Errorf("repository.AdminGetSheetsPrintedLast24Hours: %w", err),
		)
	}

	return total, nil
}

func (r *PostgresAdminRepo) AdminGetRevenueFromSingle_Sided_Prints(ctx context.Context) (int, error) {
	const query = `
SELECT COALESCE(SUM(uf.price),0) AS double_side_revenue
FROM upload_sessions us
JOIN upload_files uf
ON us.id = uf.session_id
WHERE us.status IN ('completed','paid')
AND uf.printing_side = 'double_side'
  `

	var total int
	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, apperror.Internal(
			"failed to calculate revenue from single-sided prints",
			fmt.Errorf("repository.AdminGetRevenueFromSingle_Sided_Prints: %w", err),
		)
	}

	return total, nil
}

func (r *PostgresAdminRepo) AdminGetRevenueFromDouble_Sided_Prints(ctx context.Context) (int, error) {
	const query = `
SELECT
    COALESCE(SUM(uf.price), 0) AS double_side_revenue
FROM upload_sessions us
JOIN upload_files uf
ON us.id = uf.session_id
WHERE us.status = 'completed'
  AND uf.printing_side = 'double_side'`

	var total int
	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, apperror.Internal(
			"failed to calculate revenue from double-sided prints",
			fmt.Errorf("repository.AdminGetRevenueFromDouble_Sided_Prints: %w", err),
		)
	}

	return total, nil
}

func (r *PostgresAdminRepo) AdminGetLast24HoursRevenueFromSingle_Sided_Prints(ctx context.Context) (int, error) {
	const query = `
SELECT COALESCE(SUM(uf.price),0)
FROM upload_sessions us
JOIN upload_files uf
ON us.id = uf.session_id
WHERE us.status IN ('completed','paid')
AND uf.printing_side='single_side'
AND (created_at AT TIME ZONE 'Asia/Kolkata')::date =
(NOW() AT TIME ZONE 'Asia/Kolkata')::date`
	var total int
	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, apperror.Internal(
			"failed to calculate last 24 hours revenue from single-sided prints",
			fmt.Errorf("repository.AdminGetLast24HoursRevenueFromSingle_Sided_Prints: %w", err),
		)
	}

	return total, nil
}

func (r *PostgresAdminRepo) AdminGetLast24HoursRevenueFromDouble_Sided_Prints(ctx context.Context) (int, error) {
	const query = `
SELECT COALESCE(SUM(uf.price),0)
FROM upload_sessions us
JOIN upload_files uf
ON us.id = uf.session_id
WHERE us.status IN ('completed','paid')
AND uf.printing_side='double_side'
AND (us.created_at AT TIME ZONE 'Asia/Kolkata')::date =
(NOW() AT TIME ZONE 'Asia/Kolkata')::date`

	var total int
	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, apperror.Internal(
			"failed to calculate last 24 hours revenue from double-sided prints",
			fmt.Errorf("repository.AdminGetLast24HoursRevenueFromDouble_Sided_Prints: %w", err),
		)
	}

	return total, nil
}

// func (r *PostgresAdminRepo) AdminGetTotalSheestPrintedInLast24Hours(ctx context.Context) (int, error) {
// 	const query = `	SELECT COALESCE(SUM(total_sheets), 0)
// 	AS total_sheets FROM upload_sessions
// 	WHERE status IN ('completed', 'paid')
//     AND (created_at AT TIME ZONE 'Asia/Kolkata')::date = DATE '2026-07-15'
// 	`
// 	var total int
// 	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
// 		return 0, apperror.Internal(
// 			"failed to calculate total sheets printed in last 24 hours",
// 			fmt.Errorf("repository.AdminGetTotalSheestPrintedInLast24Hours: %w", err),
// 		)
// 	}

// 	return total, nil
// }

func (r *PostgresAdminRepo) AdminGetTotalSheetsPrintedInLast24HoursBySingle_Sided_Prints(ctx context.Context) (int, error) {
	query := `SELECT COALESCE(
    SUM(
        CEIL(uf.number_of_pages::numeric / uf.page_layout) * uf.copies
    ),
    0
) AS single_side_sheets
FROM upload_sessions us
JOIN upload_files uf
ON us.id = uf.session_id
WHERE us.status IN ('completed','paid')
  AND uf.printing_side = 'single_side'
  AND (us.created_at AT TIME ZONE 'Asia/Kolkata')::date =
      (NOW() AT TIME ZONE 'Asia/Kolkata')::date`
	var total int
	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, apperror.Internal(
			"failed to calculate total sheets printed in last 24 hours by single-sided prints",
			fmt.Errorf("repository.AdminGetTotalSheetsPrintedInLast24HoursBySingle_Sided_Prints: %w", err),
		)
	}

	return total, nil
}

func (r *PostgresAdminRepo) AdminGetTotalSheetsPrintedInLast24HoursByDouble_Sided_Prints(ctx context.Context) (int, error) {
	query := `SELECT COALESCE(
    SUM(
        CEIL(uf.number_of_pages::numeric / (uf.page_layout * 2)) * uf.copies
    ),
    0
) AS double_side_sheets
FROM upload_sessions us
JOIN upload_files uf
ON us.id = uf.session_id
WHERE us.status IN ('completed','paid')
AND uf.printing_side = 'double_side'
AND (us.created_at AT TIME ZONE 'Asia/Kolkata')::date =
(NOW() AT TIME ZONE 'Asia/Kolkata')::date`
	var total int
	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, apperror.Internal(
			"failed to calculate total sheets printed in last 24 hours by double-sided prints",
			fmt.Errorf("repository.AdminGetTotalSheetsPrintedInLast24HoursByDouble_Sided_Prints: %w", err),
		)
	}

	return total, nil
}

func (r *PostgresAdminRepo) AdminGetTotalSheetsPrintedBySingle_Sided_Prints(ctx context.Context) (int, error) {
	query := `
	SELECT COALESCE(
    SUM(
        CEIL(uf.number_of_pages::numeric / uf.page_layout) * uf.copies
    ),
    0
) AS single_side_sheets
FROM upload_sessions us
JOIN upload_files uf
ON us.id = uf.session_id
WHERE us.status IN ('completed','paid')
  AND uf.printing_side = 'single_side'
	`
	var total int
	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, apperror.Internal(
			"failed to calculate total sheets printed by single-sided prints",
			fmt.Errorf("repository.AdminGetTotalSheetsPrintedBySingle_Sided_Prints: %w", err),
		)
	}

	return total, nil
}

func (r *PostgresAdminRepo) AdminGetTotalSheetsPrintedByDouble_Sided_Prints(ctx context.Context) (int, error) {
	query := `SELECT COALESCE(
    SUM(
        CEIL(uf.number_of_pages::numeric / (uf.page_layout * 2)) * uf.copies
    ),
    0
) AS double_side_sheets
FROM upload_sessions us
JOIN upload_files uf
ON us.id = uf.session_id
WHERE us.status IN ('completed','paid')
AND uf.printing_side = 'double_side'`
	var total int
	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, apperror.Internal(
			"failed to calculate total sheets printed by double-sided prints",
			fmt.Errorf("repository.AdminGetTotalSheetsPrintedByDouble_Sided_Prints: %w", err),
		)
	}

	return total, nil
}

func (r *PostgresAdminRepo) AdminFetchPrintHistoryfor24H(ctx context.Context) ([]models.PrintJob, error) {
	const sessionsQuery = `
	SELECT id, token, status, total_amount, total_sheets, created_at
	FROM upload_sessions
	WHERE status = ('completed','paid')
	  AND (created_at AT TIME ZONE 'Asia/Kolkata')::date =
	      (NOW() AT TIME ZONE 'Asia/Kolkata')::date
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

func (r *PostgresAdminRepo) AdminFetchPrintJobsOnlyPaidInLast24H(ctx context.Context) ([]models.PrintJob, error) {
	const sessionsQuery = `
	SELECT id, token, status, total_amount, total_sheets, created_at
	FROM upload_sessions
	WHERE status = 'paid'
	  AND (created_at AT TIME ZONE 'Asia/Kolkata')::date =
	      (NOW() AT TIME ZONE 'Asia/Kolkata')::date
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

func (r *PostgresAdminRepo) AdminGetTotalSessionCountOfToday(ctx context.Context) (int, error) {
	query := `
	SELECT COUNT(*) AS total_session
	FROM upload_sessions
	WHERE status IN ('completed','paid')
	AND (created_at AT TIME ZONE 'Asia/Kolkata')::date =
	(NOW() AT TIME ZONE 'Asia/Kolkata')::date
	`

	var total int
	if err := r.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, apperror.Internal(
			"failed to calculate today's session count",
			fmt.Errorf("repository.AdminGetTotalSessionCountOfToday: %w", err),
		)
	}

	return total, nil
}

func (r *PostgresAdminRepo) AdminGetTodaysPaymentHistory(ctx context.Context) ([]models.Payment, error) {

	query := `
	SELECT id, session_id, razorpay_order_id, razorpay_payment_id, status
	FROM payments
	WHERE (created_at AT TIME ZONE 'Asia/Kolkata')::date =
	      (NOW() AT TIME ZONE 'Asia/Kolkata')::date
	ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)

	if err != nil {
		return nil, apperror.Internal(
			"failed to payment history",
			fmt.Errorf("repository.AdminGetTodaysPayment History: %w", err),
		)
	}

	defer rows.Close()

	payments := make([]models.Payment, 0)

	for rows.Next() {
		var p models.Payment

		if err := rows.Scan(
			&p.ID,
			&p.SessionID,
			&p.RazorpayOrderID,
			&p.RazorpayPaymentID,
			&p.Status,
		); err != nil {
			return nil, apperror.Internal(
				"failed to scan payment history row",
				fmt.Errorf("repository.AdminGetTodaysPaymentHistory scan: %w", err),
			)
		}
		payments = append(payments, p)
	}

	if err := rows.Err(); err != nil {
		return nil, apperror.Internal(
			"error reading payment history rows",
			fmt.Errorf("repository.AdminGetTodaysPaymentHistory rows.Err: %w", err),
		)
	}

	return payments, nil
}

func (r *PostgresAdminRepo) AdminGetPrintJobsForPricedStatus(ctx context.Context) ([]models.PrintJob, error) {
	const sessionsQuery = `
	SELECT id, token, status, total_amount, total_sheets, created_at
	FROM upload_sessions
	WHERE status = 'priced'
	  AND (created_at AT TIME ZONE 'Asia/Kolkata')::date =
	      (NOW() AT TIME ZONE 'Asia/Kolkata')::date
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
