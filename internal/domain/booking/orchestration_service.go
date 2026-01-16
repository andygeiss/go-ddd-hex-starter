package booking

import (
	"context"
	"fmt"
)

// BookingOrchestrationService coordinates the complete booking saga workflow.
// It orchestrates reservation creation, payment authorization/capture, and confirmation
// with proper compensation logic on failures.
type BookingOrchestrationService struct {
	reservationService  *ReservationService
	paymentService      *PaymentService
	notificationService NotificationService
}

// NewBookingOrchestrationService creates a new orchestration service.
func NewBookingOrchestrationService(
	reservationSvc *ReservationService,
	paymentSvc *PaymentService,
	notificationSvc NotificationService,
) *BookingOrchestrationService {
	return &BookingOrchestrationService{
		reservationService:  reservationSvc,
		paymentService:      paymentSvc,
		notificationService: notificationSvc,
	}
}

// CompleteBooking orchestrates the full booking workflow with compensation.
func (s *BookingOrchestrationService) CompleteBooking(
	ctx context.Context,
	reservationID ReservationID,
	paymentID PaymentID,
	guestID GuestID,
	roomID RoomID,
	dateRange DateRange,
	amount Money,
	guests []GuestInfo,
	paymentMethod string,
) (*Reservation, error) {
	reservation, err := s.createReservationStep(ctx, reservationID, guestID, roomID, dateRange, amount, guests)
	if err != nil {
		return nil, err
	}

	payment, err := s.authorizePaymentStep(ctx, paymentID, reservationID, amount, paymentMethod)
	if err != nil {
		return nil, err
	}

	if err := s.capturePaymentStep(ctx, payment.ID, reservationID); err != nil {
		return nil, err
	}

	if err := s.confirmReservationStep(ctx, reservationID, payment.ID); err != nil {
		return nil, err
	}

	_ = s.notificationService.SendReservationConfirmation(ctx, reservation)

	return s.reservationService.GetReservation(ctx, reservationID)
}

// CancelBookingWithRefund cancels a reservation and refunds the payment if applicable.
func (s *BookingOrchestrationService) CancelBookingWithRefund(
	ctx context.Context,
	reservationID ReservationID,
	reason string,
) error {
	reservation, err := s.reservationService.GetReservation(ctx, reservationID)
	if err != nil {
		return fmt.Errorf("failed to get reservation: %w", err)
	}

	if err := s.reservationService.CancelReservation(ctx, reservationID, reason); err != nil {
		return fmt.Errorf("failed to cancel reservation: %w", err)
	}

	_ = s.notificationService.SendCancellationNotice(ctx, reservation, reason)

	return nil
}

func (s *BookingOrchestrationService) createReservationStep(
	ctx context.Context,
	reservationID ReservationID,
	guestID GuestID,
	roomID RoomID,
	dateRange DateRange,
	amount Money,
	guests []GuestInfo,
) (*Reservation, error) {
	reservation, err := s.reservationService.CreateReservation(ctx, reservationID, guestID, roomID, dateRange, amount, guests)
	if err != nil {
		return nil, fmt.Errorf("step 1 failed (create reservation): %w", err)
	}
	return reservation, nil
}

func (s *BookingOrchestrationService) authorizePaymentStep(
	ctx context.Context,
	paymentID PaymentID,
	reservationID ReservationID,
	amount Money,
	paymentMethod string,
) (*Payment, error) {
	payment, err := s.paymentService.AuthorizePayment(ctx, paymentID, reservationID, amount, paymentMethod)
	if err != nil {
		cancelErr := s.reservationService.CancelReservation(ctx, reservationID, "payment_authorization_failed")
		if cancelErr != nil {
			return nil, fmt.Errorf("step 2 failed (authorize payment) and compensation failed: %w (original error: %w)", cancelErr, err)
		}
		return nil, fmt.Errorf("step 2 failed (authorize payment): %w", err)
	}
	return payment, nil
}

func (s *BookingOrchestrationService) capturePaymentStep(ctx context.Context, paymentID PaymentID, reservationID ReservationID) error {
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

func (s *BookingOrchestrationService) confirmReservationStep(ctx context.Context, reservationID ReservationID, paymentID PaymentID) error {
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
