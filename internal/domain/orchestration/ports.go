package orchestration

import (
	"context"

	"github.com/andygeiss/hotel-booking/internal/domain/payment"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
)

// NotificationService handles sending notifications to guests.
type NotificationService interface {
	// SendReservationConfirmation sends a confirmation email to the guest
	SendReservationConfirmation(ctx context.Context, r *reservation.Reservation) error
	// SendCancellationNotice sends a cancellation notice to the guest
	SendCancellationNotice(ctx context.Context, r *reservation.Reservation, reason string) error
	// SendPaymentReceipt sends a payment receipt to the guest
	SendPaymentReceipt(ctx context.Context, p *payment.Payment) error
}
