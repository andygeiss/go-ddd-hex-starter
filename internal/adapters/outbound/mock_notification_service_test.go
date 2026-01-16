package outbound_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/outbound"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/booking"
)

// ============================================================================
// MockNotificationService Tests
// ============================================================================

func createTestReservation() *booking.Reservation {
	checkIn := time.Now().AddDate(0, 0, 7)
	checkOut := time.Now().AddDate(0, 0, 10)

	return &booking.Reservation{
		ID:        "res-001",
		GuestID:   "guest-001",
		RoomID:    "room-101",
		DateRange: booking.NewDateRange(checkIn, checkOut),
		Status:    booking.StatusPending,
		TotalAmount: booking.NewMoney(30000, "USD"),
		Guests: []booking.GuestInfo{
			booking.NewGuestInfo("John Doe", "john@example.com", "+1234567890"),
		},
	}
}

func createTestPayment() *booking.Payment {
	return booking.NewPayment(
		"pay-001",
		"res-001",
		booking.NewMoney(30000, "USD"),
		"credit_card",
	)
}

func Test_MockNotificationService_SendReservationConfirmation_Should_Succeed(t *testing.T) {
	// Arrange
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := outbound.NewMockNotificationService(logger)
	ctx := context.Background()
	reservation := createTestReservation()

	// Act
	err := svc.SendReservationConfirmation(ctx, reservation)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
}

func Test_MockNotificationService_SendReservationConfirmation_No_Guests_Should_Return_Error(t *testing.T) {
	// Arrange
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := outbound.NewMockNotificationService(logger)
	ctx := context.Background()
	reservation := createTestReservation()
	reservation.Guests = []booking.GuestInfo{}

	// Act
	err := svc.SendReservationConfirmation(ctx, reservation)

	// Assert
	assert.That(t, "error must not be nil for no guests", err != nil, true)
}

func Test_MockNotificationService_SendCancellationNotice_Should_Succeed(t *testing.T) {
	// Arrange
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := outbound.NewMockNotificationService(logger)
	ctx := context.Background()
	reservation := createTestReservation()

	// Act
	err := svc.SendCancellationNotice(ctx, reservation, "guest requested cancellation")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
}

func Test_MockNotificationService_SendCancellationNotice_No_Guests_Should_Return_Error(t *testing.T) {
	// Arrange
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := outbound.NewMockNotificationService(logger)
	ctx := context.Background()
	reservation := createTestReservation()
	reservation.Guests = []booking.GuestInfo{}

	// Act
	err := svc.SendCancellationNotice(ctx, reservation, "test")

	// Assert
	assert.That(t, "error must not be nil for no guests", err != nil, true)
}

func Test_MockNotificationService_SendPaymentReceipt_Should_Succeed(t *testing.T) {
	// Arrange
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := outbound.NewMockNotificationService(logger)
	ctx := context.Background()
	payment := createTestPayment()

	// Act
	err := svc.SendPaymentReceipt(ctx, payment)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
}
