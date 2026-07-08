package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strings"

	"github.com/Lorenta-Tech/kiosk-server/internal/models"
	"github.com/Lorenta-Tech/kiosk-server/internal/repository"
	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
	s3pkg "github.com/Lorenta-Tech/kiosk-server/pkg/s3"
	"github.com/google/uuid"
)

const razorpayBaseURL = "https://api.razorpay.com/v1"

type PaymentService struct {
	paymentrepo    repository.PaymentRepo
	filerepo       repository.FileRepo
	db             *sql.DB
	s3             *s3pkg.Client
	logger         *slog.Logger
	razorpayKey    string
	razorpaySecret string
	webhookSecret  string
}

func NewPaymentService(
	paymentrepo repository.PaymentRepo,
	filerepo repository.FileRepo,
	db *sql.DB,
	s3 *s3pkg.Client,
	logger *slog.Logger,
	razorpayKey string,
	razorpaySecret string,
	webhookSecret string,
) *PaymentService {
	return &PaymentService{
		paymentrepo:    paymentrepo,
		filerepo:       filerepo,
		db:             db,
		s3:             s3,
		logger:         logger,
		razorpayKey:    razorpayKey,
		razorpaySecret: razorpaySecret,
		webhookSecret:  webhookSecret,
	}
}

func (ps *PaymentService) CreateOrder(
	ctx context.Context,
	req models.CreatePaymentRequest,
) (models.CreatePaymentResponse, error) {

	ps.logger.Info("create payment order started",
		"session_id", req.SessionID,
		"amount_paise_frontend", req.AmountPaise,
	)

	session, err := ps.filerepo.GetSessionByID(ctx, req.SessionID)
	if err != nil {
		return models.CreatePaymentResponse{}, err
	}

	if session.Status != "priced" {
		return models.CreatePaymentResponse{}, apperror.BadRequest(
			"session_not_priced",
			fmt.Sprintf("session is in status '%s', must be 'priced' before payment", session.Status),
		)
	}

	if session.TotalAmount == nil {
		return models.CreatePaymentResponse{}, apperror.Internal(
			"session has no total amount", nil,
		)
	}
	dbAmountPaise := rupeeToP(*session.TotalAmount)

	if req.AmountPaise != dbAmountPaise {
		ps.logger.Warn("amount mismatch — possible tamper attempt",
			"session_id", req.SessionID,
			"frontend_paise", req.AmountPaise,
			"db_paise", dbAmountPaise,
		)
		return models.CreatePaymentResponse{}, apperror.BadRequest(
			"amount_mismatch",
			"payment amount does not match the order total",
		)
	}

	rzpOrderID, err := ps.createRazorpayOrder(ctx, dbAmountPaise, req.SessionID)
	if err != nil {
		return models.CreatePaymentResponse{}, err
	}

	ps.logger.Info("razorpay order created",
		"session_id", req.SessionID,
		"razorpay_order_id", rzpOrderID,
		"amount_paise", dbAmountPaise,
	)

	paymentID := uuid.NewString()
	if err := ps.paymentrepo.CreatePayment(ctx, models.Payment{
		ID:              paymentID,
		SessionID:       req.SessionID,
		RazorpayOrderID: rzpOrderID,
		Status:          "created",
	}); err != nil {
		return models.CreatePaymentResponse{}, err
	}

	ps.logger.Info("create payment order completed",
		"session_id", req.SessionID,
		"payment_id", paymentID,
		"razorpay_order_id", rzpOrderID,
	)

	return models.CreatePaymentResponse{
		PaymentID:       paymentID,
		RazorpayOrderID: rzpOrderID,
		AmountPaise:     dbAmountPaise,
		Currency:        "INR",
		SessionID:       req.SessionID,
	}, nil
}

func (ps *PaymentService) GetPaymentStatus(
	ctx context.Context,
	sessionId string,
) (models.TokenJobResponse, error) {
	ps.logger.Info("get job by session ID started", "session_id", sessionId)

	session, err := ps.filerepo.GetSessionByID(ctx, sessionId)
	if err != nil {
		return models.TokenJobResponse{}, err
	}

	ps.logger.Info("Token Debugging", "Token:", session.Token)

	files, err := ps.filerepo.GetFilesBySessionID(ctx, session.ID)
	if err != nil {
		return models.TokenJobResponse{}, err
	}

	printJobFiles := make([]models.PrintJobFile, 0, len(files))
	for _, f := range files {
		pf := models.PrintJobFile{
			FileID:        f.ID,
			FileName:      f.FileName,
			PrintingMode:  f.PrintingMode,
			PrintingSide:  f.PrintingSide,
			PageRange:     f.PageRange,
			PageLayout:    f.PageLayout,
			Copies:        f.Copies,
			NumberOfPages: f.NumberOfPages,
			Price:         f.Price,
			FileStatus:    f.FileStatus,
		}
		printJobFiles = append(printJobFiles, pf)
	}

	ps.logger.Info("get job by session ID completed",
		"session_id", sessionId,
		"file_count", len(files),
	)

	return models.TokenJobResponse{
		Job: models.PrintJob{
			SessionID:   session.ID,
			Status:      session.Status,
			Token:       session.Token,
			TotalAmount: session.TotalAmount,
			TotalSheets: session.TotalSheets,
			CreatedAt:   session.CreatedAt,
			Files:       printJobFiles,
		},
	}, nil
}

// HandleWebhook processes inbound Razorpay webhook events.
func (ps *PaymentService) HandleWebhook(r *http.Request) error {
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		return apperror.Internal("failed to read webhook body", err)
	}

	signature := r.Header.Get("X-Razorpay-Signature")
	if !ps.verifyWebhookSignature(rawBody, signature) {
		ps.logger.Warn("webhook signature verification failed",
			"signature", signature,
		)
		return apperror.Unauthorized("webhook signature is invalid")
	}

	var payload models.RazorpayWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return apperror.Internal("failed to parse webhook payload", err)
	}

	ps.logger.Info("webhook received",
		"event", payload.Event,
	)

	switch payload.Event {
	case "payment.captured", "order.paid":
		return ps.handlePaymentSuccess(r.Context(), payload)
	case "payment.failed":
		return ps.handlePaymentFailed(r.Context(), payload)
	default:
		// Acknowledge unknown events — never return non-2xx to Razorpay
		// or it will retry indefinitely.
		ps.logger.Info("webhook event ignored", "event", payload.Event)
		return nil
	}
}

// handlePaymentSuccess
func (ps *PaymentService) handlePaymentSuccess(ctx context.Context, payload models.RazorpayWebhookPayload) error {
	if payload.Payload.Payment == nil {
		return apperror.Internal("payment.captured event missing payment entity", nil)
	}

	entity := payload.Payload.Payment.Entity
	orderID := entity.OrderID
	paymentID := entity.ID

	ps.logger.Info("processing payment success",
		"razorpay_order_id", orderID,
		"razorpay_payment_id", paymentID,
	)

	payment, err := ps.paymentrepo.GetPaymentByOrderID(ctx, orderID)
	if err != nil {
		return err
	}

	if payment.Status == "success" {
		ps.logger.Info("duplicate webhook ignored — payment already succeeded",
			"razorpay_order_id", orderID,
		)
		return nil
	}

	session, err := ps.filerepo.GetSessionByID(ctx, payment.SessionID)
	if err != nil {
		return err
	}

	if session.Status == "paid" {
		ps.logger.Info("session already paid — skipping promotion",
			"session_id", session.ID,
		)
		return nil
	}

	files, err := ps.filerepo.GetFilesBySessionID(ctx, session.ID)
	if err != nil {
		return err
	}

	//  Promote each file staging → final
	type promotedFile struct {
		fileID   string
		finalKey string
	}
	promoted := make([]promotedFile, 0, len(files))

	for _, f := range files {
		finalKey, err := ps.s3.PromoteFile(ctx, f.StagingKey)
		if err != nil {
			ps.logger.Error("failed to promote file",
				"session_id", session.ID,
				"file_id", f.ID,
				"staging_key", f.StagingKey,
				"error", err,
			)
			return apperror.Internal(
				fmt.Sprintf("failed to promote file %s to final storage", f.FileName), err,
			)
		}

		ps.logger.Info("file promoted",
			"session_id", session.ID,
			"file_id", f.ID,
			"final_key", finalKey,
		)

		promoted = append(promoted, promotedFile{fileID: f.ID, finalKey: finalKey})
	}

	tx, err := ps.db.BeginTx(ctx, nil)
	if err != nil {
		return apperror.Internal("failed to begin transaction", err)
	}
	defer tx.Rollback()

	txPayment := ps.paymentrepo.WithTx(tx)
	txFile := ps.filerepo.WithTx(tx)

	for _, pf := range promoted {
		if err := txFile.MarkFilePromoted(ctx, pf.fileID, pf.finalKey); err != nil {
			return err
		}
	}

	// Mark session as paid
	if err := txFile.UpdateSessionPaid(ctx, session.ID); err != nil {
		return err
	}

	// Mark payment as success
	if err := txPayment.UpdatePaymentSuccess(ctx, orderID, paymentID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return apperror.Internal("failed to commit payment transaction", err)
	}

	ps.logger.Info("payment success committed",
		"session_id", session.ID,
		"razorpay_order_id", orderID,
		"razorpay_payment_id", paymentID,
		"files_promoted", len(promoted),
	)

	return nil
}

// handlePaymentFailed
func (ps *PaymentService) handlePaymentFailed(ctx context.Context, payload models.RazorpayWebhookPayload) error {
	if payload.Payload.Payment == nil {
		return apperror.Internal("payment.failed event missing payment entity", nil)
	}

	orderID := payload.Payload.Payment.Entity.OrderID

	ps.logger.Warn("payment failed",
		"razorpay_order_id", orderID,
	)

	if err := ps.paymentrepo.UpdatePaymentFailed(ctx, orderID); err != nil {
		return err
	}

	ps.logger.Info("payment failure recorded", "razorpay_order_id", orderID)
	return nil
}

//	helpers
//
// verifyWebhookSignature computes HMAC-SHA256(webhookSecret, rawBody)
func (ps *PaymentService) verifyWebhookSignature(body []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(ps.webhookSecret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(strings.TrimSpace(signature)))
}

// createRazorpayOrder calls Razorpay's Orders API to create an order.
func (ps *PaymentService) createRazorpayOrder(ctx context.Context, amountPaise int64, sessionID string) (string, error) {
	body := fmt.Sprintf(
		`{"amount":%d,"currency":"INR","receipt":"%s","payment_capture":1}`,
		amountPaise, sessionID,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		razorpayBaseURL+"/orders",
		strings.NewReader(body),
	)
	if err != nil {
		return "", apperror.Internal("failed to build razorpay request", err)
	}

	req.SetBasicAuth(ps.razorpayKey, ps.razorpaySecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", apperror.Internal("razorpay api call failed", err)
	}
	defer resp.Body.Close()

	var result struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Error  *struct {
			Description string `json:"description"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", apperror.Internal("failed to decode razorpay response", err)
	}

	if resp.StatusCode != http.StatusOK || result.ID == "" {
		desc := "razorpay order creation failed"
		if result.Error != nil {
			desc = result.Error.Description
		}
		return "", apperror.Internal(desc, fmt.Errorf("razorpay status: %d", resp.StatusCode))
	}

	return result.ID, nil
}

// rupeeToP converts rupees (float64, from DB NUMERIC) to paise (int64).
func rupeeToP(rupees float64) int64 {
	return int64(math.Round(rupees * 100))
}
