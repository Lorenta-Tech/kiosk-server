package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/Lorenta-Tech/kiosk-server/internal/models"
	"github.com/jackc/pgconn"
)

// failingDBTX fails every Exec with err; the query methods are never reached.
type failingDBTX struct {
	err error
}

func (f failingDBTX) ExecContext(context.Context, string, ...any) (sql.Result, error) {
	return nil, f.err
}
func (f failingDBTX) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	return nil, f.err
}
func (f failingDBTX) QueryRowContext(context.Context, string, ...any) *sql.Row {
	return nil
}

func testSession() models.UploadSession {
	return models.UploadSession{
		ID:        "6b1f0d3e-0f1a-4a5b-9c1d-2e3f4a5b6c7d",
		UserID:    "1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d",
		UserEmail: "user@example.com",
		Status:    "created",
		Token:     "123456",
	}
}

func TestCreateSessionReportsTokenTaken(t *testing.T) {
	repo := NewFileRepository(failingDBTX{err: &pgconn.PgError{
		Code:           pgUniqueViolation,
		ConstraintName: sessionTokenConstraint,
		Message:        "duplicate key value violates unique constraint",
	}})

	err := repo.CreateSession(context.Background(), testSession())
	if !errors.Is(err, ErrSessionTokenTaken) {
		t.Fatalf("err = %v, want it to match ErrSessionTokenTaken", err)
	}
}

func TestCreateSessionIgnoresOtherUniqueViolations(t *testing.T) {
	repo := NewFileRepository(failingDBTX{err: &pgconn.PgError{
		Code:           pgUniqueViolation,
		ConstraintName: "upload_sessions_pkey",
		Message:        "duplicate key value violates unique constraint",
	}})

	err := repo.CreateSession(context.Background(), testSession())
	if errors.Is(err, ErrSessionTokenTaken) {
		t.Fatal("a violation on another constraint must not be reported as a token collision")
	}
}

func TestCreateSessionIgnoresOtherErrors(t *testing.T) {
	repo := NewFileRepository(failingDBTX{err: errors.New("connection refused")})

	err := repo.CreateSession(context.Background(), testSession())
	if err == nil {
		t.Fatal("expected an error")
	}
	if errors.Is(err, ErrSessionTokenTaken) {
		t.Fatal("a connection failure must not be reported as a token collision")
	}
}
