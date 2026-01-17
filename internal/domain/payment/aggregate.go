// Package payment contains the Payment bounded context.
// It handles all payment-related domain logic including authorization,
// capture, refund, and payment attempt tracking.
package payment

import (
	"errors"
	"fmt"
	"time"

	"github.com/andygeiss/hotel-booking/internal/domain/shared"
)

// Type aliases for shared types
type ReservationID = shared.ReservationID
type Money = shared.Money

// PaymentID is a strongly-typed identifier for payments.
type PaymentID string

// PaymentStatus represents the state of a payment.
type PaymentStatus string

const (
	StatusPending    PaymentStatus = "pending"
	StatusAuthorized PaymentStatus = "authorized"
	StatusCaptured   PaymentStatus = "captured"
	StatusFailed     PaymentStatus = "failed"
	StatusRefunded   PaymentStatus = "refunded"
)

// Payment is the aggregate root for payment processing.
type Payment struct {
	ID            PaymentID
	ReservationID ReservationID
	Amount        Money
	Status        PaymentStatus
	PaymentMethod string
	TransactionID string // External payment gateway transaction ID
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Attempts      []PaymentAttempt
}

// Payment errors.
var (
	ErrInvalidPaymentTransition = errors.New("invalid payment state transition")
	ErrAlreadyAuthorized        = errors.New("payment already authorized")
	ErrNotAuthorized            = errors.New("payment not authorized")
	ErrAlreadyCaptured          = errors.New("payment already captured")
	ErrNotCaptured              = errors.New("payment not captured")
	ErrAlreadyRefunded          = errors.New("payment already refunded")
	ErrCannotRefund             = errors.New("can only refund captured payments")
)

// NewPayment creates a new payment in pending status.
func NewPayment(id PaymentID, reservationID ReservationID, amount Money, method string) *Payment {
	return &Payment{
		ID:            id,
		ReservationID: reservationID,
		Amount:        amount,
		Status:        StatusPending,
		PaymentMethod: method,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Attempts:      []PaymentAttempt{},
	}
}

// Authorize transitions the payment to authorized status.
func (p *Payment) Authorize(transactionID string) error {
	if p.Status == StatusAuthorized {
		return ErrAlreadyAuthorized
	}

	if p.Status != StatusPending && p.Status != StatusFailed {
		return fmt.Errorf("%w: cannot authorize from %s", ErrInvalidPaymentTransition, p.Status)
	}

	p.Status = StatusAuthorized
	p.TransactionID = transactionID
	p.UpdatedAt = time.Now()
	p.addAttempt(StatusAuthorized, "", "")

	return nil
}

// Capture transitions the payment to captured status (finalizes the payment).
func (p *Payment) Capture() error {
	if p.Status == StatusCaptured {
		return ErrAlreadyCaptured
	}

	if p.Status != StatusAuthorized {
		return ErrNotAuthorized
	}

	p.Status = StatusCaptured
	p.UpdatedAt = time.Now()
	p.addAttempt(StatusCaptured, "", "")

	return nil
}

// Fail marks the payment as failed with error details.
func (p *Payment) Fail(errorCode, errorMsg string) error {
	if p.Status == StatusCaptured || p.Status == StatusRefunded {
		return fmt.Errorf("%w: cannot fail from %s", ErrInvalidPaymentTransition, p.Status)
	}

	p.Status = StatusFailed
	p.UpdatedAt = time.Now()
	p.addAttempt(StatusFailed, errorCode, errorMsg)

	return nil
}

// Refund transitions the payment to refunded status.
func (p *Payment) Refund() error {
	if p.Status == StatusRefunded {
		return ErrAlreadyRefunded
	}

	if p.Status != StatusCaptured {
		return ErrCannotRefund
	}

	p.Status = StatusRefunded
	p.UpdatedAt = time.Now()
	p.addAttempt(StatusRefunded, "", "")

	return nil
}

// IsSuccessful returns true if the payment was successfully captured.
func (p *Payment) IsSuccessful() bool {
	return p.Status == StatusCaptured
}

// CanBeRetried returns true if the payment can be retried.
func (p *Payment) CanBeRetried() bool {
	if p.Status != StatusFailed && p.Status != StatusPending {
		return false
	}

	failedAttempts := 0
	for _, attempt := range p.Attempts {
		if attempt.Status == StatusFailed {
			failedAttempts++
		}
	}

	return failedAttempts < 3
}

// addAttempt adds a payment attempt to the history.
func (p *Payment) addAttempt(status PaymentStatus, errorCode, errorMsg string) {
	attempt := PaymentAttempt{
		AttemptedAt: time.Now(),
		Status:      status,
		ErrorCode:   errorCode,
		ErrorMsg:    errorMsg,
	}
	p.Attempts = append(p.Attempts, attempt)
}
