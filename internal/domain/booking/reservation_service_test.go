package booking_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/booking"
)

// ============================================================================
// ReservationService Tests
// ============================================================================

func Test_ReservationService_CreateReservation_Should_Succeed(t *testing.T) {
	// Arrange
	resRepo := newMockReservationRepository()
	availChecker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	svc := booking.NewReservationService(resRepo, availChecker, publisher)
	ctx := context.Background()

	checkIn := time.Now().AddDate(0, 0, 7)
	checkOut := time.Now().AddDate(0, 0, 10)
	dateRange := booking.NewDateRange(checkIn, checkOut)
	guests := []booking.GuestInfo{booking.NewGuestInfo("John Doe", "john@example.com", "+1234567890")}
	amount := booking.NewMoney(30000, "USD")

	// Act
	reservation, err := svc.CreateReservation(
		ctx,
		"res-001",
		"guest-001",
		"room-101",
		dateRange,
		amount,
		guests,
	)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "reservation must not be nil", reservation != nil, true)
	assert.That(t, "status must be pending", reservation.Status, booking.StatusPending)

	persisted, readErr := resRepo.Read(ctx, "res-001")
	assert.That(t, "reservation must be persisted", readErr == nil, true)
	assert.That(t, "persisted ID must match", string(persisted.ID), "res-001")
	assert.That(t, "one event must be published", len(publisher.events), 1)
}

func Test_ReservationService_CreateReservation_Room_Unavailable_Should_Fail(t *testing.T) {
	// Arrange
	resRepo := newMockReservationRepository()
	availChecker := &mockAvailabilityChecker{available: false}
	publisher := &mockEventPublisher{}
	svc := booking.NewReservationService(resRepo, availChecker, publisher)
	ctx := context.Background()

	checkIn := time.Now().AddDate(0, 0, 7)
	checkOut := time.Now().AddDate(0, 0, 10)
	dateRange := booking.NewDateRange(checkIn, checkOut)
	guests := []booking.GuestInfo{booking.NewGuestInfo("John Doe", "john@example.com", "+1234567890")}
	amount := booking.NewMoney(30000, "USD")

	// Act
	_, err := svc.CreateReservation(
		ctx,
		"res-001",
		"guest-001",
		"room-101",
		dateRange,
		amount,
		guests,
	)

	// Assert
	assert.That(t, "error must not be nil for unavailable room", err != nil, true)
}

func Test_ReservationService_CreateReservation_Availability_Check_Error_Should_Fail(t *testing.T) {
	// Arrange
	resRepo := newMockReservationRepository()
	availChecker := &mockAvailabilityChecker{checkErr: errors.New("availability check failed")}
	publisher := &mockEventPublisher{}
	svc := booking.NewReservationService(resRepo, availChecker, publisher)
	ctx := context.Background()

	checkIn := time.Now().AddDate(0, 0, 7)
	checkOut := time.Now().AddDate(0, 0, 10)
	dateRange := booking.NewDateRange(checkIn, checkOut)
	guests := []booking.GuestInfo{booking.NewGuestInfo("John Doe", "john@example.com", "+1234567890")}
	amount := booking.NewMoney(30000, "USD")

	// Act
	_, err := svc.CreateReservation(
		ctx,
		"res-001",
		"guest-001",
		"room-101",
		dateRange,
		amount,
		guests,
	)

	// Assert
	assert.That(t, "error must not be nil for availability check failure", err != nil, true)
}

func Test_ReservationService_CreateReservation_Invalid_Data_Should_Fail(t *testing.T) {
	// Arrange
	resRepo := newMockReservationRepository()
	availChecker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	svc := booking.NewReservationService(resRepo, availChecker, publisher)
	ctx := context.Background()

	checkIn := time.Now().AddDate(0, 0, 10)
	checkOut := time.Now().AddDate(0, 0, 7)
	dateRange := booking.NewDateRange(checkIn, checkOut)
	guests := []booking.GuestInfo{booking.NewGuestInfo("John Doe", "john@example.com", "+1234567890")}
	amount := booking.NewMoney(30000, "USD")

	// Act
	_, err := svc.CreateReservation(
		ctx,
		"res-001",
		"guest-001",
		"room-101",
		dateRange,
		amount,
		guests,
	)

	// Assert
	assert.That(t, "error must not be nil for invalid date range", err != nil, true)
}

func Test_ReservationService_ConfirmReservation_Should_Succeed(t *testing.T) {
	// Arrange
	resRepo := newMockReservationRepository()
	availChecker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	svc := booking.NewReservationService(resRepo, availChecker, publisher)
	ctx := context.Background()

	checkIn := time.Now().AddDate(0, 0, 7)
	checkOut := time.Now().AddDate(0, 0, 10)
	dateRange := booking.NewDateRange(checkIn, checkOut)
	guests := []booking.GuestInfo{booking.NewGuestInfo("John Doe", "john@example.com", "+1234567890")}
	amount := booking.NewMoney(30000, "USD")

	_, setupErr := svc.CreateReservation(ctx, "res-001", "guest-001", "room-101", dateRange, amount, guests)
	if setupErr != nil {
		t.Fatalf("setup failed: %v", setupErr)
	}

	// Act
	err := svc.ConfirmReservation(ctx, "res-001")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)

	persisted, readErr := resRepo.Read(ctx, "res-001")
	assert.That(t, "reservation must exist", readErr == nil, true)
	assert.That(t, "status must be confirmed", persisted.Status, booking.StatusConfirmed)
}

func Test_ReservationService_ConfirmReservation_Not_Found_Should_Fail(t *testing.T) {
	// Arrange
	resRepo := newMockReservationRepository()
	availChecker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	svc := booking.NewReservationService(resRepo, availChecker, publisher)
	ctx := context.Background()

	// Act
	err := svc.ConfirmReservation(ctx, "nonexistent")

	// Assert
	assert.That(t, "error must not be nil for nonexistent reservation", err != nil, true)
}

func Test_ReservationService_CancelReservation_Should_Succeed(t *testing.T) {
	// Arrange
	resRepo := newMockReservationRepository()
	availChecker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	svc := booking.NewReservationService(resRepo, availChecker, publisher)
	ctx := context.Background()

	checkIn := time.Now().AddDate(0, 0, 7)
	checkOut := time.Now().AddDate(0, 0, 10)
	dateRange := booking.NewDateRange(checkIn, checkOut)
	guests := []booking.GuestInfo{booking.NewGuestInfo("John Doe", "john@example.com", "+1234567890")}
	amount := booking.NewMoney(30000, "USD")

	_, setupErr := svc.CreateReservation(ctx, "res-001", "guest-001", "room-101", dateRange, amount, guests)
	if setupErr != nil {
		t.Fatalf("setup failed: %v", setupErr)
	}

	// Act
	err := svc.CancelReservation(ctx, "res-001", "guest requested")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)

	persisted, readErr := resRepo.Read(ctx, "res-001")
	assert.That(t, "reservation must exist", readErr == nil, true)
	assert.That(t, "status must be cancelled", persisted.Status, booking.StatusCancelled)
	assert.That(t, "cancellation reason must match", persisted.CancellationReason, "guest requested")
}

func Test_ReservationService_GetReservation_Should_Return_Reservation(t *testing.T) {
	// Arrange
	resRepo := newMockReservationRepository()
	availChecker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	svc := booking.NewReservationService(resRepo, availChecker, publisher)
	ctx := context.Background()

	checkIn := time.Now().AddDate(0, 0, 7)
	checkOut := time.Now().AddDate(0, 0, 10)
	dateRange := booking.NewDateRange(checkIn, checkOut)
	guests := []booking.GuestInfo{booking.NewGuestInfo("John Doe", "john@example.com", "+1234567890")}
	amount := booking.NewMoney(30000, "USD")

	_, setupErr := svc.CreateReservation(ctx, "res-001", "guest-001", "room-101", dateRange, amount, guests)
	if setupErr != nil {
		t.Fatalf("setup failed: %v", setupErr)
	}

	// Act
	reservation, err := svc.GetReservation(ctx, "res-001")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "reservation must not be nil", reservation != nil, true)
	assert.That(t, "ID must match", string(reservation.ID), "res-001")
}

func Test_ReservationService_GetReservation_Not_Found_Should_Fail(t *testing.T) {
	// Arrange
	resRepo := newMockReservationRepository()
	availChecker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	svc := booking.NewReservationService(resRepo, availChecker, publisher)
	ctx := context.Background()

	// Act
	_, err := svc.GetReservation(ctx, "nonexistent")

	// Assert
	assert.That(t, "error must not be nil for nonexistent reservation", err != nil, true)
}

func Test_ReservationService_ListReservationsByGuest_Should_Return_Guest_Reservations(t *testing.T) {
	// Arrange
	resRepo := newMockReservationRepository()
	availChecker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	svc := booking.NewReservationService(resRepo, availChecker, publisher)
	ctx := context.Background()

	guests := []booking.GuestInfo{booking.NewGuestInfo("John Doe", "john@example.com", "+1234567890")}
	amount := booking.NewMoney(30000, "USD")

	checkIn := time.Now().AddDate(0, 0, 7)
	checkOut := time.Now().AddDate(0, 0, 10)
	dateRange := booking.NewDateRange(checkIn, checkOut)
	_, _ = svc.CreateReservation(ctx, "res-001", "guest-001", "room-101", dateRange, amount, guests)

	checkIn2 := time.Now().AddDate(0, 0, 14)
	checkOut2 := time.Now().AddDate(0, 0, 17)
	dateRange2 := booking.NewDateRange(checkIn2, checkOut2)
	_, _ = svc.CreateReservation(ctx, "res-002", "guest-001", "room-102", dateRange2, amount, guests)

	checkIn3 := time.Now().AddDate(0, 0, 21)
	checkOut3 := time.Now().AddDate(0, 0, 24)
	dateRange3 := booking.NewDateRange(checkIn3, checkOut3)
	_, _ = svc.CreateReservation(ctx, "res-003", "guest-002", "room-103", dateRange3, amount, guests)

	// Act
	reservations, err := svc.ListReservationsByGuest(ctx, "guest-001")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "must return 2 reservations for guest-001", len(reservations), 2)
}

func Test_ReservationService_ListReservationsByGuest_No_Reservations_Should_Return_Empty(t *testing.T) {
	// Arrange
	resRepo := newMockReservationRepository()
	availChecker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	svc := booking.NewReservationService(resRepo, availChecker, publisher)
	ctx := context.Background()

	// Act
	reservations, err := svc.ListReservationsByGuest(ctx, "guest-999")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "must return 0 reservations", len(reservations), 0)
}
