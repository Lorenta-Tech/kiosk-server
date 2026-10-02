package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/Lorenta-Tech/kiosk-server/internal/repository"
	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestSessionTokenStrIsAlwaysSixDigits(t *testing.T) {
	for _, token := range []int{sessionTokenMin, sessionTokenMax, 1, 42, 99999} {
		got := sessionTokenStr(token)
		if len(got) != sessionTokenDigits {
			t.Fatalf("sessionTokenStr(%d) = %q, want %d characters", token, got, sessionTokenDigits)
		}
	}
}

func TestGenerateTokenStaysInSixDigitRange(t *testing.T) {
	for i := 0; i < 1000; i++ {
		token, err := generateToken()
		if err != nil {
			t.Fatalf("generateToken: %v", err)
		}
		if token < sessionTokenMin || token > sessionTokenMax {
			t.Fatalf("generateToken() = %d, want %d..%d", token, sessionTokenMin, sessionTokenMax)
		}
	}
}

func TestWithUniqueSessionTokenSucceedsOnFirstAttempt(t *testing.T) {
	attempts := 0
	token, err := withUniqueSessionToken(context.Background(), discardLogger(),
		func(context.Context, int) error {
			attempts++
			return nil
		})
	if err != nil {
		t.Fatalf("withUniqueSessionToken: %v", err)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
	if token < sessionTokenMin || token > sessionTokenMax {
		t.Fatalf("token = %d, want %d..%d", token, sessionTokenMin, sessionTokenMax)
	}
}

func TestWithUniqueSessionTokenRetriesCollision(t *testing.T) {
	attempts := 0
	// Same shape as CreateSession: an AppError wrapping ErrSessionTokenTaken.
	tokenTaken := func() error {
		return apperror.Internal("failed to create upload session",
			errors.Join(repository.ErrSessionTokenTaken, errors.New("duplicate key")))
	}

	token, err := withUniqueSessionToken(context.Background(), discardLogger(),
		func(context.Context, int) error {
			attempts++
			if attempts < 3 {
				return tokenTaken()
			}
			return nil
		})
	if err != nil {
		t.Fatalf("withUniqueSessionToken: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
	if token < sessionTokenMin || token > sessionTokenMax {
		t.Fatalf("token = %d, want %d..%d", token, sessionTokenMin, sessionTokenMax)
	}
}

func TestWithUniqueSessionTokenRetriesBareSentinel(t *testing.T) {
	attempts := 0
	_, err := withUniqueSessionToken(context.Background(), discardLogger(),
		func(context.Context, int) error {
			attempts++
			if attempts == 1 {
				return repository.ErrSessionTokenTaken
			}
			return nil
		})
	if err != nil {
		t.Fatalf("withUniqueSessionToken: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
}

func TestWithUniqueSessionTokenDoesNotRetryOtherErrors(t *testing.T) {
	attempts := 0
	want := apperror.BadRequest("validation_error", "nope")

	_, err := withUniqueSessionToken(context.Background(), discardLogger(),
		func(context.Context, int) error {
			attempts++
			return want
		})
	if !errors.Is(err, want) {
		t.Fatalf("err = %v, want %v", err, want)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}

func TestWithUniqueSessionTokenGivesUpAfterMaxAttempts(t *testing.T) {
	attempts := 0
	_, err := withUniqueSessionToken(context.Background(), discardLogger(),
		func(context.Context, int) error {
			attempts++
			return errors.Join(repository.ErrSessionTokenTaken, errors.New("duplicate key"))
		})
	if err == nil {
		t.Fatal("expected an error once every attempt collided")
	}
	if attempts != sessionTokenMaxAttempts {
		t.Fatalf("attempts = %d, want %d", attempts, sessionTokenMaxAttempts)
	}
	if !apperror.As(err, new(*apperror.AppError)) {
		t.Fatalf("err = %v, want an *apperror.AppError", err)
	}
}
