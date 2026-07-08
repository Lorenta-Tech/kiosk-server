package models

// CreatePaymentRequest is sent by the frontend after /upload/confirm.
// session_id identifies which priced session to create an order for.
// amount_paise is the amount the frontend believes is owed, in paise (INR × 100).
// The server validates this against the DB before creating the Razorpay order.
type CreatePaymentRequest struct {
	SessionID   string `json:"session_id"   validate:"required,uuid4"`
	AmountPaise int64  `json:"amount_paise" validate:"required,min=1"`
}

type Payment struct {
	ID                string
	SessionID         string
	RazorpayOrderID   string
	RazorpayPaymentID *string
	Status            string // created | success | failed
}

type CreatePaymentResponse struct {
	PaymentID       string `json:"payment_id"`
	RazorpayOrderID string `json:"razorpay_order_id"`
	AmountPaise     int64  `json:"amount_paise"`
	Currency        string `json:"currency"`
	SessionID       string `json:"session_id"`
}

// RazorpayWebhookPayload is the top-level body Razorpay POSTs to your server.
type RazorpayWebhookPayload struct {
	Entity    string               `json:"entity"`
	AccountID string               `json:"account_id"`
	Event     string               `json:"event"`
	Payload   RazorpayEventPayload `json:"payload"`
}

type RazorpayEventPayload struct {
	Payment *RazorpayPaymentWrapper `json:"payment"`
	Order   *RazorpayOrderWrapper   `json:"order"`
}

type RazorpayPaymentWrapper struct {
	Entity RazorpayPaymentEntity `json:"entity"`
}

type RazorpayOrderWrapper struct {
	Entity RazorpayOrderEntity `json:"entity"`
}

type RazorpayPaymentEntity struct {
	ID      string `json:"id"`
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"` //in paise
	Status  string `json:"status"`
}
type RazorpayOrderEntity struct {
	ID     string `json:"id"`
	Amount int64  `json:"amount"` // in paise
	Status string `json:"status"`
}
