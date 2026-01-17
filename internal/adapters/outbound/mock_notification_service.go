package outbound

import (
	"context"
	"errors"
	"log/slog"

	"github.com/andygeiss/hotel-booking/internal/domain/payment"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
)

// MockNotificationService implements NotificationService by logging to console.
type MockNotificationService struct {
	logger *slog.Logger
}

// NewMockNotificationService creates a new mock notification service.
func NewMockNotificationService(logger *slog.Logger) *MockNotificationService {
	return &MockNotificationService{
		logger: logger,
	}
}

// SendReservationConfirmation logs a confirmation message.
func (s *MockNotificationService) SendReservationConfirmation(
	ctx context.Context,
	res *reservation.Reservation,
) error {
	if len(res.Guests) == 0 {
		return errors.New("no guests found in reservation")
	}

	primaryGuest := res.Guests[0]

	s.logger.Info("sending reservation confirmation email",
		"reservation_id", res.ID,
		"guest_email", primaryGuest.Email,
		"guest_name", primaryGuest.Name,
		"room_id", res.RoomID,
		"check_in", res.DateRange.CheckIn.Format("2006-01-02"),
		"check_out", res.DateRange.CheckOut.Format("2006-01-02"),
		"total_amount", res.TotalAmount.FormatAmount(),
	)

	return nil
}

// SendCancellationNotice logs a cancellation message.
func (s *MockNotificationService) SendCancellationNotice(
	ctx context.Context,
	res *reservation.Reservation,
	reason string,
) error {
	if len(res.Guests) == 0 {
		return errors.New("no guests found in reservation")
	}

	primaryGuest := res.Guests[0]

	s.logger.Info("sending cancellation notice email",
		"reservation_id", res.ID,
		"guest_email", primaryGuest.Email,
		"guest_name", primaryGuest.Name,
		"reason", reason,
	)

	return nil
}

// SendPaymentReceipt logs a payment receipt message.
func (s *MockNotificationService) SendPaymentReceipt(
	ctx context.Context,
	pay *payment.Payment,
) error {
	s.logger.Info("sending payment receipt email",
		"payment_id", pay.ID,
		"reservation_id", pay.ReservationID,
		"amount", pay.Amount.FormatAmount(),
		"payment_method", pay.PaymentMethod,
		"transaction_id", pay.TransactionID,
	)

	return nil
}
