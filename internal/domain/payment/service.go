package payment

import (
	"context"
	"fmt"

	"github.com/andygeiss/cloud-native-utils/event"
	"github.com/andygeiss/hotel-booking/internal/domain/shared"
)

// Service handles payment workflows.
type Service struct {
	paymentRepo    PaymentRepository
	paymentGateway PaymentGateway
	publisher      event.EventPublisher
}

// NewService creates a new payment Service with dependencies.
func NewService(
	repo PaymentRepository,
	gateway PaymentGateway,
	pub event.EventPublisher,
) *Service {
	return &Service{
		paymentRepo:    repo,
		paymentGateway: gateway,
		publisher:      pub,
	}
}

// AuthorizePayment creates a payment and authorizes it with the gateway.
func (s *Service) AuthorizePayment(
	ctx context.Context,
	id PaymentID,
	reservationID ReservationID,
	amount Money,
	method string,
) (*Payment, error) {
	// 1. Create payment aggregate
	payment := NewPayment(id, reservationID, amount, method)

	// 2. Authorize with payment gateway
	transactionID, err := s.paymentGateway.Authorize(ctx, payment)
	if err != nil {
		// Mark payment as failed
		_ = payment.Fail("gateway_error", err.Error())

		// Persist failed payment
		if persistErr := s.paymentRepo.Create(ctx, id, *payment); persistErr != nil {
			return nil, fmt.Errorf("failed to persist failed payment: %w", persistErr)
		}

		// Publish failure event
		failEvt := NewEventFailed().
			WithPaymentID(id).
			WithReservationID(reservationID).
			WithErrorCode("gateway_error").
			WithErrorMsg(err.Error())

		_ = s.publisher.Publish(ctx, failEvt)

		return nil, fmt.Errorf("payment authorization failed: %w", err)
	}

	// 3. Update payment with transaction ID
	if err := payment.Authorize(transactionID); err != nil {
		return nil, fmt.Errorf("failed to update payment status: %w", err)
	}

	// 4. Persist to repository
	if err := s.paymentRepo.Create(ctx, id, *payment); err != nil {
		return nil, fmt.Errorf("failed to persist payment: %w", err)
	}

	// 5. Publish success event
	evt := NewEventAuthorized().
		WithPaymentID(id).
		WithReservationID(reservationID).
		WithAmount(amount).
		WithTransactionID(transactionID)

	if err := s.publisher.Publish(ctx, evt); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return payment, nil
}

// CapturePayment captures an authorized payment.
func (s *Service) CapturePayment(ctx context.Context, id PaymentID) error {
	// 1. Load payment from repository
	payment, err := s.paymentRepo.Read(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to read payment: %w", err)
	}

	// 2. Capture with payment gateway
	if err := s.paymentGateway.Capture(ctx, payment.TransactionID, payment.Amount); err != nil {
		// Mark as failed
		_ = payment.Fail("capture_failed", err.Error())
		_ = s.paymentRepo.Update(ctx, id, *payment)

		// Publish failure event
		failEvt := NewEventFailed().
			WithPaymentID(id).
			WithReservationID(payment.ReservationID).
			WithErrorCode("capture_failed").
			WithErrorMsg(err.Error())

		_ = s.publisher.Publish(ctx, failEvt)

		return fmt.Errorf("payment capture failed: %w", err)
	}

	// 3. Update payment status
	if err := payment.Capture(); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// 4. Update repository
	if err := s.paymentRepo.Update(ctx, id, *payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// 5. Publish success event
	evt := NewEventCaptured().
		WithPaymentID(id).
		WithReservationID(payment.ReservationID).
		WithAmount(payment.Amount)

	if err := s.publisher.Publish(ctx, evt); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// RefundPayment processes a refund for a captured payment.
func (s *Service) RefundPayment(ctx context.Context, id PaymentID) error {
	// 1. Load payment from repository
	payment, err := s.paymentRepo.Read(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to read payment: %w", err)
	}

	// 2. Refund with payment gateway
	if err := s.paymentGateway.Refund(ctx, payment.TransactionID, payment.Amount); err != nil {
		return fmt.Errorf("payment refund failed: %w", err)
	}

	// 3. Update payment status
	if err := payment.Refund(); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// 4. Update repository
	if err := s.paymentRepo.Update(ctx, id, *payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// 5. Publish event
	evt := NewEventRefunded().
		WithPaymentID(id).
		WithReservationID(payment.ReservationID).
		WithAmount(payment.Amount)

	if err := s.publisher.Publish(ctx, evt); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// GetPayment retrieves a payment by ID.
func (s *Service) GetPayment(ctx context.Context, id PaymentID) (*Payment, error) {
	payment, err := s.paymentRepo.Read(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to read payment: %w", err)
	}
	return payment, nil
}

// AuthorizePaymentForReservation is called when a reservation.created event is received.
// This initiates the payment authorization process for a new reservation.
func (s *Service) AuthorizePaymentForReservation(
	ctx context.Context,
	paymentID PaymentID,
	reservationID ReservationID,
	amount Money,
	method string,
) (*Payment, error) {
	return s.AuthorizePayment(ctx, paymentID, reservationID, amount, method)
}

// CapturePaymentOnAuthorization is called when a payment.authorized event is received
// by the orchestration layer to capture the authorized payment.
func (s *Service) CapturePaymentOnAuthorization(ctx context.Context, paymentID PaymentID) error {
	return s.CapturePayment(ctx, paymentID)
}

// NewMoney is a convenience function to create Money using the shared package.
func NewMoney(amount int64, currency string) Money {
	return shared.NewMoney(amount, currency)
}
