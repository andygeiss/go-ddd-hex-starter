package payment

import "time"

// PaymentAttempt represents a single payment attempt (entity within Payment aggregate).
type PaymentAttempt struct {
	AttemptedAt time.Time
	Status      PaymentStatus
	ErrorCode   string
	ErrorMsg    string
}

// NewPaymentAttempt creates a new payment attempt entity.
func NewPaymentAttempt(status PaymentStatus, errorCode, errorMsg string) PaymentAttempt {
	return PaymentAttempt{
		AttemptedAt: time.Now(),
		Status:      status,
		ErrorCode:   errorCode,
		ErrorMsg:    errorMsg,
	}
}
