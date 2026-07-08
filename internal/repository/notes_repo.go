package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Lorenta-Tech/kiosk-server/internal/models"
	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
)

// ================================================================
// Interface
// ================================================================

type NotesRepo interface {
	WithTx(tx *sql.Tx) NotesRepo

	// Branch
	GetActiveBranches(ctx context.Context) ([]models.Branch, error)
	GetBranchByID(ctx context.Context, branchID string) (models.Branch, error)

	// Branch Semester
	GetSemestersByBranchID(ctx context.Context, branchID string) ([]models.BranchSemester, error)
	GetBranchSemesterByID(ctx context.Context, semesterID string) (models.BranchSemester, error)

	// Subject
	GetSubjectsBySemesterID(ctx context.Context, semesterID string) ([]models.Subject, error)
	GetSubjectByID(ctx context.Context, subjectID string) (models.Subject, error)
	CreateSubject(ctx context.Context, subject models.Subject) error

	// Module
	GetModulesBySubjectID(ctx context.Context, subjectID string) ([]models.Module, error)
	GetModuleByID(ctx context.Context, moduleID string) (models.Module, error)

	// Note
	CreateNote(ctx context.Context, note models.Note) error
	GetNoteByID(ctx context.Context, noteID string) (models.Note, error)
	GetNotesByModuleID(ctx context.Context, moduleID string) ([]models.Note, error)
	UpdateNoteStatus(ctx context.Context, noteID string, status string) error
	UpdateNote(ctx context.Context, noteID string, title string, description string) error
	DeleteNote(ctx context.Context, noteID string) error

	// Dept Admin
	GetDeptAdminByID(ctx context.Context, adminID string) (models.DeptAdmin, error)
}

// ================================================================
// Implementation
// ================================================================

type PostgresNotesRepo struct {
	db DBTX
}

func NewNotesRepository(db DBTX) *PostgresNotesRepo {
	return &PostgresNotesRepo{db: db}
}

func (r *PostgresNotesRepo) WithTx(tx *sql.Tx) NotesRepo {
	return &PostgresNotesRepo{db: tx}
}

// ================================================================
// Branch
// ================================================================

func (r *PostgresNotesRepo) GetActiveBranches(ctx context.Context) ([]models.Branch, error) {
	const query = `
		SELECT id, name, code, is_active, created_at
		FROM branches
		WHERE is_active = TRUE
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, apperror.Internal(
			"failed to fetch branches",
			fmt.Errorf("repository.GetActiveBranches: %w", err),
		)
	}
	defer rows.Close()

	branches := make([]models.Branch, 0)

	for rows.Next() {
		var b models.Branch
		if err := rows.Scan(
			&b.ID,
			&b.Name,
			&b.Code,
			&b.IsActive,
			&b.CreatedAt,
		); err != nil {
			return nil, apperror.Internal(
				"failed to scan branch row",
				fmt.Errorf("repository.GetActiveBranches scan: %w", err),
			)
		}
		branches = append(branches, b)
	}

	if err := rows.Err(); err != nil {
		return nil, apperror.Internal(
			"error reading branch rows",
			fmt.Errorf("repository.GetActiveBranches rows.Err: %w", err),
		)
	}

	return branches, nil
}

func (r *PostgresNotesRepo) GetBranchByID(ctx context.Context, branchID string) (models.Branch, error) {
	const query = `
		SELECT id, name, code, is_active, created_at
		FROM branches
		WHERE id = $1
	`

	var b models.Branch
	err := r.db.QueryRowContext(ctx, query, branchID).Scan(
		&b.ID,
		&b.Name,
		&b.Code,
		&b.IsActive,
		&b.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return models.Branch{}, apperror.NotFound(
			"branch_not_found",
			fmt.Sprintf("branch %s does not exist", branchID),
		)
	}
	if err != nil {
		return models.Branch{}, apperror.Internal(
			"failed to fetch branch",
			fmt.Errorf("repository.GetBranchByID: %w", err),
		)
	}

	return b, nil
}

// ================================================================
// Branch Semester
// ================================================================

func (r *PostgresNotesRepo) GetSemestersByBranchID(ctx context.Context, branchID string) ([]models.BranchSemester, error) {
	const query = `
		SELECT id, branch_id, semester_number, is_active, created_at
		FROM branch_semesters
		WHERE branch_id = $1 AND is_active = TRUE
		ORDER BY semester_number ASC
	`

	rows, err := r.db.QueryContext(ctx, query, branchID)
	if err != nil {
		return nil, apperror.Internal(
			"failed to fetch semesters",
			fmt.Errorf("repository.GetSemestersByBranchID: %w", err),
		)
	}
	defer rows.Close()

	semesters := make([]models.BranchSemester, 0)

	for rows.Next() {
		var s models.BranchSemester
		if err := rows.Scan(
			&s.ID,
			&s.BranchID,
			&s.SemesterNumber,
			&s.IsActive,
			&s.CreatedAt,
		); err != nil {
			return nil, apperror.Internal(
				"failed to scan semester row",
				fmt.Errorf("repository.GetSemestersByBranchID scan: %w", err),
			)
		}
		semesters = append(semesters, s)
	}

	if err := rows.Err(); err != nil {
		return nil, apperror.Internal(
			"error reading semester rows",
			fmt.Errorf("repository.GetSemestersByBranchID rows.Err: %w", err),
		)
	}

	return semesters, nil
}

func (r *PostgresNotesRepo) GetBranchSemesterByID(ctx context.Context, semesterID string) (models.BranchSemester, error) {
	const query = `
		SELECT id, branch_id, semester_number, is_active, created_at
		FROM branch_semesters
		WHERE id = $1
	`

	var s models.BranchSemester
	err := r.db.QueryRowContext(ctx, query, semesterID).Scan(
		&s.ID,
		&s.BranchID,
		&s.SemesterNumber,
		&s.IsActive,
		&s.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return models.BranchSemester{}, apperror.NotFound(
			"semester_not_found",
			fmt.Sprintf("semester %s does not exist", semesterID),
		)
	}
	if err != nil {
		return models.BranchSemester{}, apperror.Internal(
			"failed to fetch semester",
			fmt.Errorf("repository.GetBranchSemesterByID: %w", err),
		)
	}

	return s, nil
}

// ================================================================
// Subject
// ================================================================

func (r *PostgresNotesRepo) GetSubjectsBySemesterID(ctx context.Context, semesterID string) ([]models.Subject, error) {
	const query = `
		SELECT id, branch_semester_id, name, subject_code, is_active, created_at
		FROM subjects
		WHERE branch_semester_id = $1 AND is_active = TRUE
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, semesterID)
	if err != nil {
		return nil, apperror.Internal(
			"failed to fetch subjects",
			fmt.Errorf("repository.GetSubjectsBySemesterID: %w", err),
		)
	}
	defer rows.Close()

	subjects := make([]models.Subject, 0)

	for rows.Next() {
		var s models.Subject
		if err := rows.Scan(
			&s.ID,
			&s.BranchSemesterID,
			&s.Name,
			&s.SubjectCode,
			&s.IsActive,
			&s.CreatedAt,
		); err != nil {
			return nil, apperror.Internal(
				"failed to scan subject row",
				fmt.Errorf("repository.GetSubjectsBySemesterID scan: %w", err),
			)
		}
		subjects = append(subjects, s)
	}

	if err := rows.Err(); err != nil {
		return nil, apperror.Internal(
			"error reading subject rows",
			fmt.Errorf("repository.GetSubjectsBySemesterID rows.Err: %w", err),
		)
	}

	return subjects, nil
}

func (r *PostgresNotesRepo) GetSubjectByID(ctx context.Context, subjectID string) (models.Subject, error) {
	const query = `
		SELECT id, branch_semester_id, name, subject_code, is_active, created_at
		FROM subjects
		WHERE id = $1
	`

	var s models.Subject
	err := r.db.QueryRowContext(ctx, query, subjectID).Scan(
		&s.ID,
		&s.BranchSemesterID,
		&s.Name,
		&s.SubjectCode,
		&s.IsActive,
		&s.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return models.Subject{}, apperror.NotFound(
			"subject_not_found",
			fmt.Sprintf("subject %s does not exist", subjectID),
		)
	}
	if err != nil {
		return models.Subject{}, apperror.Internal(
			"failed to fetch subject",
			fmt.Errorf("repository.GetSubjectByID: %w", err),
		)
	}

	return s, nil
}

func (r *PostgresNotesRepo) CreateSubject(ctx context.Context, subject models.Subject) error {
	const query = `
		INSERT INTO subjects (id, branch_semester_id, name, subject_code)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(ctx, query,
		subject.ID,
		subject.BranchSemesterID,
		subject.Name,
		subject.SubjectCode,
	)
	if err != nil {
		return apperror.Internal(
			"failed to create subject",
			fmt.Errorf("repository.CreateSubject: %w", err),
		)
	}

	return nil
}

// ================================================================
// Module
// ================================================================

func (r *PostgresNotesRepo) GetModulesBySubjectID(ctx context.Context, subjectID string) ([]models.Module, error) {
	const query = `
		SELECT id, subject_id, module_number, title, created_at
		FROM modules
		WHERE subject_id = $1
		ORDER BY module_number ASC
	`

	rows, err := r.db.QueryContext(ctx, query, subjectID)
	if err != nil {
		return nil, apperror.Internal(
			"failed to fetch modules",
			fmt.Errorf("repository.GetModulesBySubjectID: %w", err),
		)
	}
	defer rows.Close()

	modules := make([]models.Module, 0)

	for rows.Next() {
		var m models.Module
		if err := rows.Scan(
			&m.ID,
			&m.SubjectID,
			&m.ModuleNumber,
			&m.Title,
			&m.CreatedAt,
		); err != nil {
			return nil, apperror.Internal(
				"failed to scan module row",
				fmt.Errorf("repository.GetModulesBySubjectID scan: %w", err),
			)
		}
		modules = append(modules, m)
	}

	if err := rows.Err(); err != nil {
		return nil, apperror.Internal(
			"error reading module rows",
			fmt.Errorf("repository.GetModulesBySubjectID rows.Err: %w", err),
		)
	}

	return modules, nil
}

func (r *PostgresNotesRepo) GetModuleByID(ctx context.Context, moduleID string) (models.Module, error) {
	const query = `
		SELECT id, subject_id, module_number, title, created_at
		FROM modules
		WHERE id = $1
	`

	var m models.Module
	err := r.db.QueryRowContext(ctx, query, moduleID).Scan(
		&m.ID,
		&m.SubjectID,
		&m.ModuleNumber,
		&m.Title,
		&m.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return models.Module{}, apperror.NotFound(
			"module_not_found",
			fmt.Sprintf("module %s does not exist", moduleID),
		)
	}
	if err != nil {
		return models.Module{}, apperror.Internal(
			"failed to fetch module",
			fmt.Errorf("repository.GetModuleByID: %w", err),
		)
	}

	return m, nil
}

// ================================================================
// Note
// ================================================================

func (r *PostgresNotesRepo) CreateNote(ctx context.Context, note models.Note) error {
	const query = `
		INSERT INTO notes (
			id, module_id, uploaded_by, title, description,
			file_key, file_type, original_filename, file_size_bytes, status
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		note.ID,
		note.ModuleID,
		note.UploadedBy,
		note.Title,
		note.Description,
		note.FileKey,
		note.FileType,
		note.OriginalFilename,
		note.FileSizeBytes,
		note.Status,
	)
	if err != nil {
		return apperror.Internal(
			"failed to create note",
			fmt.Errorf("repository.CreateNote: %w", err),
		)
	}

	return nil
}

func (r *PostgresNotesRepo) GetNoteByID(ctx context.Context, noteID string) (models.Note, error) {
	const query = `
		SELECT id, module_id, uploaded_by, title, description,
		       file_key, file_type, original_filename, file_size_bytes,
		       status, created_at
		FROM notes
		WHERE id = $1
	`

	var n models.Note
	err := r.db.QueryRowContext(ctx, query, noteID).Scan(
		&n.ID,
		&n.ModuleID,
		&n.UploadedBy,
		&n.Title,
		&n.Description,
		&n.FileKey,
		&n.FileType,
		&n.OriginalFilename,
		&n.FileSizeBytes,
		&n.Status,
		&n.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return models.Note{}, apperror.NotFound(
			"note_not_found",
			fmt.Sprintf("note %s does not exist", noteID),
		)
	}
	if err != nil {
		return models.Note{}, apperror.Internal(
			"failed to fetch note",
			fmt.Errorf("repository.GetNoteByID: %w", err),
		)
	}

	return n, nil
}

func (r *PostgresNotesRepo) GetNotesByModuleID(ctx context.Context, moduleID string) ([]models.Note, error) {
	const query = `
		SELECT id, module_id, uploaded_by, title, description,
		       file_key, file_type, original_filename, file_size_bytes,
		       status, created_at
		FROM notes
		WHERE module_id = $1 AND status = 'active'
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, moduleID)
	if err != nil {
		return nil, apperror.Internal(
			"failed to fetch notes",
			fmt.Errorf("repository.GetNotesByModuleID: %w", err),
		)
	}
	defer rows.Close()

	notes := make([]models.Note, 0)

	for rows.Next() {
		var n models.Note
		if err := rows.Scan(
			&n.ID,
			&n.ModuleID,
			&n.UploadedBy,
			&n.Title,
			&n.Description,
			&n.FileKey,
			&n.FileType,
			&n.OriginalFilename,
			&n.FileSizeBytes,
			&n.Status,
			&n.CreatedAt,
		); err != nil {
			return nil, apperror.Internal(
				"failed to scan note row",
				fmt.Errorf("repository.GetNotesByModuleID scan: %w", err),
			)
		}
		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		return nil, apperror.Internal(
			"error reading note rows",
			fmt.Errorf("repository.GetNotesByModuleID rows.Err: %w", err),
		)
	}

	return notes, nil
}

func (r *PostgresNotesRepo) UpdateNoteStatus(ctx context.Context, noteID string, status string) error {
	const query = `
		UPDATE notes
		SET status = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, status, noteID)
	if err != nil {
		return apperror.Internal(
			"failed to update note status",
			fmt.Errorf("repository.UpdateNoteStatus note=%s: %w", noteID, err),
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperror.Internal(
			"failed to verify note status update",
			fmt.Errorf("repository.UpdateNoteStatus rows: %w", err),
		)
	}

	if rowsAffected == 0 {
		return apperror.NotFound(
			"note_not_found",
			fmt.Sprintf("note %s does not exist", noteID),
		)
	}

	return nil
}

func (r *PostgresNotesRepo) UpdateNote(ctx context.Context, noteID string, title string, description string) error {
	const query = `
		UPDATE notes
		SET title = $1, description = $2
		WHERE id = $3 AND status != 'deleted'
	`

	result, err := r.db.ExecContext(ctx, query, title, description, noteID)
	if err != nil {
		return apperror.Internal(
			"failed to update note",
			fmt.Errorf("repository.UpdateNote note=%s: %w", noteID, err),
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperror.Internal(
			"failed to verify note update",
			fmt.Errorf("repository.UpdateNote rows: %w", err),
		)
	}

	if rowsAffected == 0 {
		return apperror.NotFound(
			"note_not_found",
			fmt.Sprintf("note %s does not exist or has been deleted", noteID),
		)
	}

	return nil
}

func (r *PostgresNotesRepo) DeleteNote(ctx context.Context, noteID string) error {
	const query = `
		UPDATE notes
		SET status = 'deleted'
		WHERE id = $1 AND status != 'deleted'
	`

	result, err := r.db.ExecContext(ctx, query, noteID)
	if err != nil {
		return apperror.Internal(
			"failed to delete note",
			fmt.Errorf("repository.DeleteNote note=%s: %w", noteID, err),
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperror.Internal(
			"failed to verify note deletion",
			fmt.Errorf("repository.DeleteNote rows: %w", err),
		)
	}

	if rowsAffected == 0 {
		return apperror.NotFound(
			"note_not_found",
			fmt.Sprintf("note %s does not exist or is already deleted", noteID),
		)
	}

	return nil
}

// ================================================================
// Dept Admin
// ================================================================

func (r *PostgresNotesRepo) GetDeptAdminByID(ctx context.Context, adminID string) (models.DeptAdmin, error) {
	const query = `
		SELECT id, branch_id, name, email, password_hash, role, created_at
		FROM dept_admins
		WHERE id = $1
	`

	var da models.DeptAdmin
	err := r.db.QueryRowContext(ctx, query, adminID).Scan(
		&da.ID,
		&da.BranchID,
		&da.Name,
		&da.Email,
		&da.PasswordHash,
		&da.Role,
		&da.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return models.DeptAdmin{}, apperror.NotFound(
			"admin_not_found",
			fmt.Sprintf("no department admin found with id %s", adminID),
		)
	}
	if err != nil {
		return models.DeptAdmin{}, apperror.Internal(
			"failed to fetch department admin",
			fmt.Errorf("repository.GetDeptAdminByID: %w", err),
		)
	}

	return da, nil
}
