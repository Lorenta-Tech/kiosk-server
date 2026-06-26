package repository

type PostgresAdminRepo struct {
	db DBTX
}
func NewAdminRepository(db DBTX) *PostgresAdminRepo {
	return &PostgresAdminRepo{db: db}
}

type AdminRepo interface {
}