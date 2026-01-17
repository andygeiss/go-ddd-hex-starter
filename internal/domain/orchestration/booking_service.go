// Package orchestration contains the application services that coordinate
// workflows across bounded contexts. This is the Saga coordinator layer
// that uses event-driven communication between contexts.
package orchestration

import (
	"context"
	"fmt"

	"github.com/andygeiss/hotel-booking/internal/domain/payment"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
	"github.com/andygeiss/hotel-booking/internal/domain/shared"
)

// BookingService coordinates the complete booking saga workflow.
// It orchestrates reservation creation, payment authorization/capture, and confirmation
// with proper compensation logic on failures.
//
// In event-driven mode:
// - InitiateBooking creates a reservation and publishes reservation.created
// - Payment context subscribes and processes payment, publishing payment.authorized/failed
// - Event handlers capture payment and confirm reservation
// - Compensation is handled via event subscriptions on failure events
type BookingService struct {
	reservationService  *reservation.Service
	paymentService      *payment.Service
	notificationService NotificationService
}

// NewBookingService creates a new orchestration service.
func NewBookingService(
	reservationSvc *reservation.Service,
	paymentSvc *payment.Service,
	notificationSvc NotificationService,
) *BookingService {
	return &BookingService{
		reservationService:  reservationSvc,
		paymentService:      paymentSvc,
		notificationService: notificationSvc,
	}
}

// InitiateBooking starts the booking saga by creating a reservation.
// This publishes a reservation.created event that triggers payment processing.
func (s *BookingService) InitiateBooking(
	ctx context.Context,
	reservationID shared.ReservationID,
	guestID reservation.GuestID,
	roomID reservation.RoomID,
	dateRange reservation.DateRange,
	amount shared.Money,
	guests []reservation.GuestInfo,
) (*reservation.Reservation, error) {
	// Create reservation (publishes reservation.created event)
	res, err := s.reservationService.CreateReservation(ctx, reservationID, guestID, roomID, dateRange, amount, guests)
	if err != nil {
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}

	// The payment context will subscribe to reservation.created and
	// initiate payment authorization automatically

	return res, nil
}

// CompleteBooking orchestrates the full booking workflow synchronously.
// This is used when direct method calls are preferred over events.
func (s *BookingService) CompleteBooking(
	ctx context.Context,
	reservationID shared.ReservationID,
	paymentID payment.PaymentID,
	guestID reservation.GuestID,
	roomID reservation.RoomID,
	dateRange reservation.DateRange,
	amount shared.Money,
	guests []reservation.GuestInfo,
	paymentMethod string,
) (*reservation.Reservation, error) {
	// Step 1: Create reservation
	res, err := s.createReservationStep(ctx, reservationID, guestID, roomID, dateRange, amount, guests)
	if err != nil {
		return nil, err
	}

	// Step 2: Authorize payment
	pay, err := s.authorizePaymentStep(ctx, paymentID, reservationID, amount, paymentMethod)
	if err != nil {
		return nil, err
	}

	// Step 3: Capture payment
	if err := s.capturePaymentStep(ctx, pay.ID, reservationID); err != nil {
		return nil, err
	}

	// Step 4: Confirm reservation
	if err := s.confirmReservationStep(ctx, reservationID, pay.ID); err != nil {
		return nil, err
	}

	// Step 5: Send notification (best effort)
	_ = s.notificationService.SendReservationConfirmation(ctx, res)

	return s.reservationService.GetReservation(ctx, reservationID)
}

// CancelBookingWithRefund cancels a reservation and refunds the payment if applicable.
func (s *BookingService) CancelBookingWithRefund(
	ctx context.Context,
	reservationID shared.ReservationID,
	reason string,
) error {
	res, err := s.reservationService.GetReservation(ctx, reservationID)
	if err != nil {
		return fmt.Errorf("failed to get reservation: %w", err)
	}

	if err := s.reservationService.CancelReservation(ctx, reservationID, reason); err != nil {
		return fmt.Errorf("failed to cancel reservation: %w", err)
	}

	_ = s.notificationService.SendCancellationNotice(ctx, res, reason)

	return nil
}

// OnPaymentAuthorized handles the payment.authorized event.
// It captures the payment and confirms the reservation.
func (s *BookingService) OnPaymentAuthorized(ctx context.Context, paymentID payment.PaymentID, reservationID shared.ReservationID) error {
	// Capture the payment
	if err := s.paymentService.CapturePayment(ctx, paymentID); err != nil {
		// Compensation: cancel the reservation
		_ = s.reservationService.CancelReservation(ctx, reservationID, "payment_capture_failed")
		return fmt.Errorf("failed to capture payment: %w", err)
	}

	// Note: The payment.captured event will trigger reservation confirmation
	// via the event handler, or we can do it here directly
	return nil
}

// OnPaymentCaptured handles the payment.captured event.
// It confirms the reservation.
func (s *BookingService) OnPaymentCaptured(ctx context.Context, reservationID shared.ReservationID) error {
	if err := s.reservationService.ConfirmReservation(ctx, reservationID); err != nil {
		return fmt.Errorf("failed to confirm reservation: %w", err)
	}

	res, err := s.reservationService.GetReservation(ctx, reservationID)
	if err == nil {
		_ = s.notificationService.SendReservationConfirmation(ctx, res)
	}

	return nil
}

// OnPaymentFailed handles the payment.failed event.
// It cancels the reservation as compensation.
func (s *BookingService) OnPaymentFailed(ctx context.Context, reservationID shared.ReservationID, reason string) error {
	return s.reservationService.CancelReservation(ctx, reservationID, reason)
}

func (s *BookingService) createReservationStep(
	ctx context.Context,
	reservationID shared.ReservationID,
	guestID reservation.GuestID,
	roomID reservation.RoomID,
	dateRange reservation.DateRange,
	amount shared.Money,
	guests []reservation.GuestInfo,
) (*reservation.Reservation, error) {
	res, err := s.reservationService.CreateReservation(ctx, reservationID, guestID, roomID, dateRange, amount, guests)
	if err != nil {
		return nil, fmt.Errorf("step 1 failed (create reservation): %w", err)
	}
	return res, nil
}

func (s *BookingService) authorizePaymentStep(
	ctx context.Context,
	paymentID payment.PaymentID,
	reservationID shared.ReservationID,
	amount shared.Money,
	paymentMethod string,
) (*payment.Payment, error) {
	pay, err := s.paymentService.AuthorizePayment(ctx, paymentID, reservationID, amount, paymentMethod)
	if err != nil {
		cancelErr := s.reservationService.CancelReservation(ctx, reservationID, "payment_authorization_failed")
		if cancelErr != nil {
			return nil, fmt.Errorf("step 2 failed (authorize payment) and compensation failed: %w (original error: %w)", cancelErr, err)
		}
		return nil, fmt.Errorf("step 2 failed (authorize payment): %w", err)
	}
	return pay, nil
}

func (s *BookingService) capturePaymentStep(ctx context.Context, paymentID payment.PaymentID, reservationID shared.ReservationID) error {
	captureErr := s.paymentService.CapturePayment(ctx, paymentID)
	if captureErr != nil {
		cancelErr := s.reservationService.CancelReservation(ctx, reservationID, "payment_capture_failed")
		if cancelErr != nil {
			return fmt.Errorf("step 3 failed (capture payment) and compensation failed: %w (original error: %w)", cancelErr, captureErr)
		}
		return fmt.Errorf("step 3 failed (capture payment): %w", captureErr)
	}
	return nil
}

func (s *BookingService) confirmReservationStep(ctx context.Context, reservationID shared.ReservationID, paymentID payment.PaymentID) error {
	confirmErr := s.reservationService.ConfirmReservation(ctx, reservationID)
	if confirmErr != nil {
		refundErr := s.paymentService.RefundPayment(ctx, paymentID)
		cancelErr := s.reservationService.CancelReservation(ctx, reservationID, "confirmation_failed")
		if refundErr != nil || cancelErr != nil {
			return fmt.Errorf("step 4 failed (confirm reservation) and compensation failed (refund: %w, cancel: %w): %w", refundErr, cancelErr, confirmErr)
		}
		return fmt.Errorf("step 4 failed (confirm reservation): %w", confirmErr)
	}
	return nil
}
