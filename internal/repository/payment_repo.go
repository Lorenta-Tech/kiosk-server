package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Lorenta-Tech/kiosk-server/internal/models"
	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
)

type PaymentRepo interface {
	// CreatePayment inserts a new payments row with status="created".
	CreatePayment(ctx context.Context, payment models.Payment) error

	// GetPaymentByOrderID fetches a payment row by razorpay_order_id.
	// Used inside the webhook handler to find which session to update.
	GetPaymentByOrderID(ctx context.Context, razorpayOrderID string) (models.Payment, error)

	// UpdatePaymentSuccess sets razorpay_payment_id and status="success".
	UpdatePaymentSuccess(ctx context.Context, razorpayOrderID, razorpayPaymentID string) error

	// UpdatePaymentFailed sets status="failed" for a given order.
	UpdatePaymentFailed(ctx context.Context, razorpayOrderID string) error

	// WithTx returns a repo scoped to the given transaction.
	WithTx(tx *sql.Tx) PaymentRepo
}

type PostgresPaymentRepo struct {
	db DBTX
}

func NewPaymentRepository(db DBTX) *PostgresPaymentRepo {
	return &PostgresPaymentRepo{db: db}
}

func (r *PostgresPaymentRepo) WithTx(tx *sql.Tx) PaymentRepo {
	return &PostgresPaymentRepo{db: tx}
}

func (r *PostgresPaymentRepo) CreatePayment(ctx context.Context, payment models.Payment) error {
	const query = `
		INSERT INTO payments (id, session_id, razorpay_order_id, status)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query,
		payment.ID,
		payment.SessionID,
		payment.RazorpayOrderID,
		payment.Status,
	)
	if err != nil {
		return apperror.Internal(
			"failed to create payment record",
			fmt.Errorf("repository.CreatePayment: %w", err),
		)
	}
	return nil
}

func (r *PostgresPaymentRepo) GetPaymentByOrderID(ctx context.Context, razorpayOrderID string) (models.Payment, error) {
	const query = `
		SELECT id, session_id, razorpay_order_id, razorpay_payment_id, status
		FROM payments
		WHERE razorpay_order_id = $1
	`
	var p models.Payment
	err := r.db.QueryRowContext(ctx, query, razorpayOrderID).Scan(
		&p.ID,
		&p.SessionID,
		&p.RazorpayOrderID,
		&p.RazorpayPaymentID,
		&p.Status,
	)
	if err == sql.ErrNoRows {
		return models.Payment{}, apperror.NotFound(
			"payment_not_found",
			fmt.Sprintf("no payment found for order %s", razorpayOrderID),
		)
	}
	if err != nil {
		return models.Payment{}, apperror.Internal(
			"failed to fetch payment",
			fmt.Errorf("repository.GetPaymentByOrderID: %w", err),
		)
	}
	return p, nil
}

func (r *PostgresPaymentRepo) UpdatePaymentSuccess(ctx context.Context, razorpayOrderID, razorpayPaymentID string) error {
	const query = `
		UPDATE payments
		SET status = 'success', razorpay_payment_id = $1
		WHERE razorpay_order_id = $2
	`
	_, err := r.db.ExecContext(ctx, query, razorpayPaymentID, razorpayOrderID)
	if err != nil {
		return apperror.Internal(
			"failed to mark payment success",
			fmt.Errorf("repository.UpdatePaymentSuccess order=%s: %w", razorpayOrderID, err),
		)
	}
	return nil
}

func (r *PostgresPaymentRepo) UpdatePaymentFailed(ctx context.Context, razorpayOrderID string) error {
	const query = `
		UPDATE payments
		SET status = 'failed'
		WHERE razorpay_order_id = $1
	`
	_, err := r.db.ExecContext(ctx, query, razorpayOrderID)
	if err != nil {
		return apperror.Internal(
			"failed to mark payment failed",
			fmt.Errorf("repository.UpdatePaymentFailed order=%s: %w", razorpayOrderID, err),
		)
	}
	return nil
}
