package db

import (
	"database/sql"
	"fmt"

	"github.com/Lorenta-Tech/kiosk-server/internal/env"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func Connect() (*sql.DB, error) {
	// var (
	// 	databaseHost     = env.GetString("DATABASE_HOST", "localhost")
	// 	databasePort     = env.GetString("DATABASE_PORT", "5432")
	// 	databaseUser     = env.GetString("DATABASE_USER", "postgres")
	// 	databseName      = env.GetString("DATABASE_NAME", "kiosk_db")
	// 	databasePassword = env.GetString("DATABASE_PASSWORD", "mustang1969")
	// )

	db, err := sql.Open(
		"pgx",
		env.GetString("DATABASE_URL", ""),
	)

	if err != nil {
		return nil, fmt.Errorf("Connect: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Connect: %w", err)
	}
	return db, nil
}
