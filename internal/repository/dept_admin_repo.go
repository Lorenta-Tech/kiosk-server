package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Lorenta-Tech/kiosk-server/internal/models"
	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
)

type AdminRepo interface {
	CreateDeptAdmin(
		ctx context.Context,
		admin models.DeptAdmin,
	) error

	GetDeptAdminByEmail(
		ctx context.Context,
		email string,
	) (models.DeptAdmin, error)

	GetDeptAdminByID(
		ctx context.Context,
		adminID string,
	) (models.DeptAdmin, error)

	ListDeptAdmins(
		ctx context.Context,
	) ([]models.DeptAdmin, error)
}

type PostgresAdminRepo struct {
	db DBTX
}

func NewAdminRepository(db DBTX) *PostgresAdminRepo {
	return &PostgresAdminRepo{
		db: db,
	}
}

func (r *PostgresAdminRepo) CreateDeptAdmin(
	ctx context.Context,
	admin models.DeptAdmin,
) error {

	const query = `
		INSERT INTO dept_admins (
			id,
			branch_id,
			name,
			email,
			password_hash,
			role
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		admin.ID,
		admin.BranchID,
		admin.Name,
		admin.Email,
		admin.PasswordHash,
		admin.Role,
	)

	if err != nil {
		return apperror.Internal(
			"failed to create department admin",
			fmt.Errorf("repository.CreateDeptAdmin: %w", err),
		)
	}

	return nil
}

func (r *PostgresAdminRepo) GetDeptAdminByEmail(
	ctx context.Context,
	email string,
) (models.DeptAdmin, error) {

	const query = `
		SELECT
			id,
			branch_id,
			name,
			email,
			password_hash,
			role,
			created_at
		FROM dept_admins
		WHERE email = $1
	`

	var admin models.DeptAdmin

	err := r.db.QueryRowContext(
		ctx,
		query,
		email,
	).Scan(
		&admin.ID,
		&admin.BranchID,
		&admin.Name,
		&admin.Email,
		&admin.PasswordHash,
		&admin.Role,
		&admin.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return models.DeptAdmin{}, apperror.NotFound(
			"dept_admin_not_found",
			fmt.Sprintf("department admin with email %s not found", email),
		)
	}

	if err != nil {
		return models.DeptAdmin{}, apperror.Internal(
			"failed to fetch department admin",
			fmt.Errorf("repository.GetDeptAdminByEmail: %w", err),
		)
	}

	return admin, nil
}

func (r *PostgresAdminRepo) GetDeptAdminByID(
	ctx context.Context,
	adminID string,
) (models.DeptAdmin, error) {

	const query = `
		SELECT
			id,
			branch_id,
			name,
			email,
			password_hash,
			role,
			created_at
		FROM dept_admins
		WHERE id = $1
	`

	var admin models.DeptAdmin

	err := r.db.QueryRowContext(
		ctx,
		query,
		adminID,
	).Scan(
		&admin.ID,
		&admin.BranchID,
		&admin.Name,
		&admin.Email,
		&admin.PasswordHash,
		&admin.Role,
		&admin.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return models.DeptAdmin{}, apperror.NotFound(
			"dept_admin_not_found",
			fmt.Sprintf("department admin %s not found", adminID),
		)
	}

	if err != nil {
		return models.DeptAdmin{}, apperror.Internal(
			"failed to fetch department admin",
			fmt.Errorf("repository.GetDeptAdminByID: %w", err),
		)
	}

	return admin, nil
}

func (r *PostgresAdminRepo) ListDeptAdmins(
	ctx context.Context,
) ([]models.DeptAdmin, error) {

	const query = `
		SELECT
			id,
			branch_id,
			name,
			email,
			password_hash,
			role,
			created_at
		FROM dept_admins
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
	)
	if err != nil {
		return nil, apperror.Internal(
			"failed to fetch department admins",
			fmt.Errorf("repository.ListDeptAdmins: %w", err),
		)
	}
	defer rows.Close()

	admins := make([]models.DeptAdmin, 0)

	for rows.Next() {

		var admin models.DeptAdmin

		if err := rows.Scan(
			&admin.ID,
			&admin.BranchID,
			&admin.Name,
			&admin.Email,
			&admin.PasswordHash,
			&admin.Role,
			&admin.CreatedAt,
		); err != nil {
			return nil, apperror.Internal(
				"failed to scan department admin row",
				fmt.Errorf("repository.ListDeptAdmins scan: %w", err),
			)
		}

		admins = append(admins, admin)
	}

	if err := rows.Err(); err != nil {
		return nil, apperror.Internal(
			"error reading department admin rows",
			fmt.Errorf("repository.ListDeptAdmins rows.Err: %w", err),
		)
	}

	return admins, nil
}
