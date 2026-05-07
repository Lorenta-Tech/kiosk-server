package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Lorenta-Tech/kiosk-server/internal/models"
	"github.com/Lorenta-Tech/kiosk-server/internal/service"
	"github.com/Lorenta-Tech/kiosk-server/internal/validator"
	"github.com/Lorenta-Tech/kiosk-server/pkg/apperror"
	"github.com/Lorenta-Tech/kiosk-server/pkg/utils"
)

type PaymentHandler struct {
	paymentservice *service.PaymentService
	logger         *slog.Logger
}

func NewPaymentHandler(paymentservice *service.PaymentService, logger *slog.Logger) *PaymentHandler {
	return &PaymentHandler{paymentservice: paymentservice, logger: logger}
}

// HandleCreateOrder godoc
//
//	@Summary      Create a Razorpay payment order
//	@Description  Called after /files/upload/confirm. Validates that the amount
//	              the frontend sends matches what is stored in the DB, then
//	              creates a Razorpay order. The frontend passes the returned
//	              razorpay_order_id directly to the Razorpay checkout modal.
//	@Tags         payments
//	@Accept       json
//	@Produce      json
//	@Param        body  body      models.CreatePaymentRequest  true  "Session and amount"
//	@Success      201   {object}  models.CreatePaymentResponse
//	@Failure      400   {object}  utils.Envelope
//	@Failure      500   {object}  utils.Envelope
//	@Router       /payments/create [post]
func (ph *PaymentHandler) HandleCreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	req, err := utils.DecodeJSON[models.CreatePaymentRequest](r)
	if err != nil {
		utils.HandleError(w, ph.logger, err)
		return
	}

	if err := validator.Validate(req); err != nil {
		utils.HandleError(w, ph.logger, apperror.BadRequest("validation_error", err.Error()))
		return
	}

	resp, err := ph.paymentservice.CreateOrder(ctx, req)
	if err != nil {
		utils.HandleError(w, ph.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"data": resp})
}

// HandleWebhook godoc
//
//	@Summary      Razorpay webhook receiver
//	@Description  Razorpay calls this endpoint server-to-server when a payment
//	              event occurs. The signature in X-Razorpay-Signature is verified
//	              before any processing. On payment.captured, files are promoted
//	              from S3 staging to final and the session is marked paid.
//	              Always returns 200 to Razorpay — non-2xx causes Razorpay to
//	              retry for 24 hours with exponential backoff.
//	@Tags         payments
//	@Accept       json
//	@Produce      json
//	@Success      200
//	@Failure      401  {object}  utils.Envelope
//	@Failure      500  {object}  utils.Envelope
//	@Router       /webhooks/razorpay [post]
func (ph *PaymentHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 4500*time.Millisecond)
	defer cancel()

	r = r.WithContext(ctx)

	if err := ph.paymentservice.HandleWebhook(r); err != nil {
		utils.HandleError(w, ph.logger, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"status": "ok"})
}