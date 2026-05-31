package domain

import "time"

// Payment tracks payment for an appointment
type Payment struct {
	ID                 string     `json:"id"`
	AppointmentID      string     `json:"appointment_id"`
	PatientID          string     `json:"patient_id"`
	Amount             float64    `json:"amount"`
	Currency           string     `json:"currency"`
	Status             string     `json:"status"` // pending, successful, failed, refunded
	PaystackReference  *string    `json:"paystack_reference"`
	PaystackAccessCode *string    `json:"paystack_access_code"`
	PaidAt             *time.Time `json:"paid_at"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type InitializePaymentRequest struct {
	AppointmentID string `json:"appointment_id" binding:"required"`
}

type InitializePaymentResponse struct {
	AuthorizationURL string  `json:"authorization_url"`
	Reference        string  `json:"reference"`
	Payment          Payment `json:"payment"`
}

type PaystackWebhookEvent struct {
	Event string                 `json:"event"`
	Data  map[string]interface{} `json:"data"`
}

var (
	ErrPaymentNotFound = NewAppError("payment_not_found", "payment not found", 404)
	ErrPaymentFailed   = NewAppError("payment_failed", "payment could not be processed", 402)
	ErrInvalidWebhook  = NewAppError("invalid_webhook", "invalid webhook signature", 401)
)
