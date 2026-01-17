package reservation_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/event"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
	"github.com/andygeiss/hotel-booking/internal/domain/shared"
)

// ============================================================================
// Mock Implementations
// ============================================================================

type mockReservationRepository struct {
	reservations map[reservation.ReservationID]reservation.Reservation
	createErr    error
	readErr      error
	updateErr    error
	deleteErr    error
}

func newMockReservationRepository() *mockReservationRepository {
	return &mockReservationRepository{
		reservations: make(map[reservation.ReservationID]reservation.Reservation),
	}
}

func (m *mockReservationRepository) Create(ctx context.Context, id reservation.ReservationID, res reservation.Reservation) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.reservations[id] = res
	return nil
}

func (m *mockReservationRepository) Read(ctx context.Context, id reservation.ReservationID) (*reservation.Reservation, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	res, ok := m.reservations[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return &res, nil
}

func (m *mockReservationRepository) Update(ctx context.Context, id reservation.ReservationID, res reservation.Reservation) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.reservations[id] = res
	return nil
}

func (m *mockReservationRepository) Delete(ctx context.Context, id reservation.ReservationID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.reservations, id)
	return nil
}

func (m *mockReservationRepository) ReadAll(ctx context.Context) ([]reservation.Reservation, error) {
	result := make([]reservation.Reservation, 0, len(m.reservations))
	for _, res := range m.reservations {
		result = append(result, res)
	}
	return result, nil
}

type mockAvailabilityChecker struct {
	available bool
	err       error
}

func (m *mockAvailabilityChecker) IsRoomAvailable(ctx context.Context, roomID reservation.RoomID, dateRange reservation.DateRange) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.available, nil
}

func (m *mockAvailabilityChecker) GetOverlappingReservations(ctx context.Context, roomID reservation.RoomID, dateRange reservation.DateRange) ([]*reservation.Reservation, error) {
	return nil, nil
}

type mockEventPublisher struct {
	published []event.Event
	err       error
}

func (m *mockEventPublisher) Publish(ctx context.Context, evt event.Event) error {
	if m.err != nil {
		return m.err
	}
	m.published = append(m.published, evt)
	return nil
}

// ============================================================================
// Service Test Helpers
// ============================================================================

func createTestService(repo *mockReservationRepository, checker *mockAvailabilityChecker, publisher *mockEventPublisher) *reservation.Service {
	return reservation.NewService(repo, checker, publisher)
}

func serviceValidDateRange() reservation.DateRange {
	checkIn := time.Now().Add(48 * time.Hour).Truncate(24 * time.Hour)
	checkOut := checkIn.Add(72 * time.Hour)
	return reservation.NewDateRange(checkIn, checkOut)
}

func serviceValidGuests() []reservation.GuestInfo {
	return []reservation.GuestInfo{
		reservation.NewGuestInfo("John Doe", "john@example.com", "+1234567890"),
	}
}

func serviceValidMoney() shared.Money {
	return shared.NewMoney(10000, "USD")
}

// ============================================================================
// CreateReservation Tests
// ============================================================================

func Test_Service_CreateReservation_Should_Succeed(t *testing.T) {
	// Arrange
	repo := newMockReservationRepository()
	checker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	service := createTestService(repo, checker, publisher)

	ctx := context.Background()
	id := reservation.ReservationID("res-001")
	guestID := reservation.GuestID("guest-001")
	roomID := reservation.RoomID("room-101")
	dateRange := serviceValidDateRange()
	amount := serviceValidMoney()
	guests := serviceValidGuests()

	// Act
	res, err := service.CreateReservation(ctx, id, guestID, roomID, dateRange, amount, guests)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "reservation must not be nil", res != nil, true)
	assert.That(t, "reservation ID must match", res.ID, id)
	assert.That(t, "reservation status must be pending", res.Status, reservation.StatusPending)
}

func Test_Service_CreateReservation_When_Room_Unavailable_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newMockReservationRepository()
	checker := &mockAvailabilityChecker{available: false}
	publisher := &mockEventPublisher{}
	service := createTestService(repo, checker, publisher)

	ctx := context.Background()
	id := reservation.ReservationID("res-001")
	guestID := reservation.GuestID("guest-001")
	roomID := reservation.RoomID("room-101")
	dateRange := serviceValidDateRange()
	amount := serviceValidMoney()
	guests := serviceValidGuests()

	// Act
	res, err := service.CreateReservation(ctx, id, guestID, roomID, dateRange, amount, guests)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "reservation must be nil", res == nil, true)
}

func Test_Service_CreateReservation_Should_Publish_Event(t *testing.T) {
	// Arrange
	repo := newMockReservationRepository()
	checker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	service := createTestService(repo, checker, publisher)

	ctx := context.Background()
	id := reservation.ReservationID("res-001")
	guestID := reservation.GuestID("guest-001")
	roomID := reservation.RoomID("room-101")
	dateRange := serviceValidDateRange()
	amount := serviceValidMoney()
	guests := serviceValidGuests()

	// Act
	_, err := service.CreateReservation(ctx, id, guestID, roomID, dateRange, amount, guests)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "one event must be published", len(publisher.published), 1)
}

func Test_Service_CreateReservation_When_Repository_Fails_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newMockReservationRepository()
	repo.createErr = errors.New("database error")
	checker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	service := createTestService(repo, checker, publisher)

	ctx := context.Background()
	id := reservation.ReservationID("res-001")
	guestID := reservation.GuestID("guest-001")
	roomID := reservation.RoomID("room-101")
	dateRange := serviceValidDateRange()
	amount := serviceValidMoney()
	guests := serviceValidGuests()

	// Act
	res, err := service.CreateReservation(ctx, id, guestID, roomID, dateRange, amount, guests)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "reservation must be nil", res == nil, true)
}

func Test_Service_CreateReservation_When_Availability_Check_Fails_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newMockReservationRepository()
	checker := &mockAvailabilityChecker{err: errors.New("availability service unavailable")}
	publisher := &mockEventPublisher{}
	service := createTestService(repo, checker, publisher)

	ctx := context.Background()
	id := reservation.ReservationID("res-001")
	guestID := reservation.GuestID("guest-001")
	roomID := reservation.RoomID("room-101")
	dateRange := serviceValidDateRange()
	amount := serviceValidMoney()
	guests := serviceValidGuests()

	// Act
	res, err := service.CreateReservation(ctx, id, guestID, roomID, dateRange, amount, guests)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "reservation must be nil", res == nil, true)
}

func Test_Service_CreateReservation_When_Publisher_Fails_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newMockReservationRepository()
	checker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{err: errors.New("publish failed")}
	service := createTestService(repo, checker, publisher)

	ctx := context.Background()
	id := reservation.ReservationID("res-001")
	guestID := reservation.GuestID("guest-001")
	roomID := reservation.RoomID("room-101")
	dateRange := serviceValidDateRange()
	amount := serviceValidMoney()
	guests := serviceValidGuests()

	// Act
	res, err := service.CreateReservation(ctx, id, guestID, roomID, dateRange, amount, guests)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "reservation must be nil", res == nil, true)
}

// ============================================================================
// ConfirmReservation Tests
// ============================================================================

func Test_Service_ConfirmReservation_Should_Update_Status(t *testing.T) {
	// Arrange
	repo := newMockReservationRepository()
	checker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	service := createTestService(repo, checker, publisher)

	ctx := context.Background()
	id := reservation.ReservationID("res-001")

	// Create a reservation first
	_, err := service.CreateReservation(ctx, id, "guest-001", "room-101", serviceValidDateRange(), serviceValidMoney(), serviceValidGuests())
	assert.That(t, "create error must be nil", err == nil, true)

	// Act
	err = service.ConfirmReservation(ctx, id)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	res, _ := repo.Read(ctx, id)
	assert.That(t, "status must be confirmed", res.Status, reservation.StatusConfirmed)
}

func Test_Service_ConfirmReservation_When_Not_Found_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newMockReservationRepository()
	checker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	service := createTestService(repo, checker, publisher)

	ctx := context.Background()
	id := reservation.ReservationID("non-existent")

	// Act
	err := service.ConfirmReservation(ctx, id)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

func Test_Service_ConfirmReservation_Should_Publish_Event(t *testing.T) {
	// Arrange
	repo := newMockReservationRepository()
	checker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	service := createTestService(repo, checker, publisher)

	ctx := context.Background()
	id := reservation.ReservationID("res-001")

	_, _ = service.CreateReservation(ctx, id, "guest-001", "room-101", serviceValidDateRange(), serviceValidMoney(), serviceValidGuests())
	publisher.published = nil // reset

	// Act
	err := service.ConfirmReservation(ctx, id)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "one event must be published", len(publisher.published), 1)
}

// ============================================================================
// CancelReservation Tests
// ============================================================================

func Test_Service_CancelReservation_Should_Update_Status(t *testing.T) {
	// Arrange
	repo := newMockReservationRepository()
	checker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	service := createTestService(repo, checker, publisher)

	ctx := context.Background()
	id := reservation.ReservationID("res-001")
	reason := "Guest requested"

	_, err := service.CreateReservation(ctx, id, "guest-001", "room-101", serviceValidDateRange(), serviceValidMoney(), serviceValidGuests())
	assert.That(t, "create error must be nil", err == nil, true)

	// Act
	err = service.CancelReservation(ctx, id, reason)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	res, _ := repo.Read(ctx, id)
	assert.That(t, "status must be cancelled", res.Status, reservation.StatusCancelled)
	assert.That(t, "cancellation reason must match", res.CancellationReason, reason)
}

func Test_Service_CancelReservation_Should_Publish_Event(t *testing.T) {
	// Arrange
	repo := newMockReservationRepository()
	checker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	service := createTestService(repo, checker, publisher)

	ctx := context.Background()
	id := reservation.ReservationID("res-001")

	_, _ = service.CreateReservation(ctx, id, "guest-001", "room-101", serviceValidDateRange(), serviceValidMoney(), serviceValidGuests())
	publisher.published = nil // reset

	// Act
	err := service.CancelReservation(ctx, id, "test reason")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "one event must be published", len(publisher.published), 1)
}

func Test_Service_CancelReservation_When_Not_Found_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newMockReservationRepository()
	checker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	service := createTestService(repo, checker, publisher)

	ctx := context.Background()

	// Act
	err := service.CancelReservation(ctx, "non-existent", "reason")

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

// ============================================================================
// ActivateReservation Tests
// ============================================================================

func Test_Service_ActivateReservation_Should_Update_Status(t *testing.T) {
	// Arrange
	repo := newMockReservationRepository()
	checker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	service := createTestService(repo, checker, publisher)

	ctx := context.Background()
	id := reservation.ReservationID("res-001")

	_, _ = service.CreateReservation(ctx, id, "guest-001", "room-101", serviceValidDateRange(), serviceValidMoney(), serviceValidGuests())
	_ = service.ConfirmReservation(ctx, id)

	// Act
	err := service.ActivateReservation(ctx, id)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	res, _ := repo.Read(ctx, id)
	assert.That(t, "status must be active", res.Status, reservation.StatusActive)
}

// ============================================================================
// CompleteReservation Tests
// ============================================================================

func Test_Service_CompleteReservation_Should_Update_Status(t *testing.T) {
	// Arrange
	repo := newMockReservationRepository()
	checker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	service := createTestService(repo, checker, publisher)

	ctx := context.Background()
	id := reservation.ReservationID("res-001")

	_, _ = service.CreateReservation(ctx, id, "guest-001", "room-101", serviceValidDateRange(), serviceValidMoney(), serviceValidGuests())
	_ = service.ConfirmReservation(ctx, id)
	_ = service.ActivateReservation(ctx, id)

	// Act
	err := service.CompleteReservation(ctx, id)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	res, _ := repo.Read(ctx, id)
	assert.That(t, "status must be completed", res.Status, reservation.StatusCompleted)
}

// ============================================================================
// GetReservation Tests
// ============================================================================

func Test_Service_GetReservation_Should_Return_Reservation(t *testing.T) {
	// Arrange
	repo := newMockReservationRepository()
	checker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	service := createTestService(repo, checker, publisher)

	ctx := context.Background()
	id := reservation.ReservationID("res-001")

	_, _ = service.CreateReservation(ctx, id, "guest-001", "room-101", serviceValidDateRange(), serviceValidMoney(), serviceValidGuests())

	// Act
	res, err := service.GetReservation(ctx, id)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "reservation must not be nil", res != nil, true)
	assert.That(t, "reservation ID must match", res.ID, id)
}

func Test_Service_GetReservation_When_Not_Found_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newMockReservationRepository()
	checker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	service := createTestService(repo, checker, publisher)

	ctx := context.Background()

	// Act
	res, err := service.GetReservation(ctx, "non-existent")

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "reservation must be nil", res == nil, true)
}

// ============================================================================
// ListReservationsByGuest Tests
// ============================================================================

func Test_Service_ListReservationsByGuest_Should_Return_Guest_Reservations(t *testing.T) {
	// Arrange
	repo := newMockReservationRepository()
	checker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	service := createTestService(repo, checker, publisher)

	ctx := context.Background()
	guestID := reservation.GuestID("guest-001")

	_, _ = service.CreateReservation(ctx, "res-001", guestID, "room-101", serviceValidDateRange(), serviceValidMoney(), serviceValidGuests())
	_, _ = service.CreateReservation(ctx, "res-002", guestID, "room-102", serviceValidDateRange(), serviceValidMoney(), serviceValidGuests())
	_, _ = service.CreateReservation(ctx, "res-003", "guest-002", "room-103", serviceValidDateRange(), serviceValidMoney(), serviceValidGuests())

	// Act
	reservations, err := service.ListReservationsByGuest(ctx, guestID)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "must have 2 reservations", len(reservations), 2)
}

// ============================================================================
// Event Handler Integration Tests
// ============================================================================

func Test_Service_ConfirmReservationOnPaymentCaptured_Should_Confirm(t *testing.T) {
	// Arrange
	repo := newMockReservationRepository()
	checker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	service := createTestService(repo, checker, publisher)

	ctx := context.Background()
	id := reservation.ReservationID("res-001")

	_, _ = service.CreateReservation(ctx, id, "guest-001", "room-101", serviceValidDateRange(), serviceValidMoney(), serviceValidGuests())

	// Act
	err := service.ConfirmReservationOnPaymentCaptured(ctx, id)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	res, _ := repo.Read(ctx, id)
	assert.That(t, "status must be confirmed", res.Status, reservation.StatusConfirmed)
}

func Test_Service_CancelReservationOnPaymentFailed_Should_Cancel(t *testing.T) {
	// Arrange
	repo := newMockReservationRepository()
	checker := &mockAvailabilityChecker{available: true}
	publisher := &mockEventPublisher{}
	service := createTestService(repo, checker, publisher)

	ctx := context.Background()
	id := reservation.ReservationID("res-001")
	reason := "payment_failed"

	_, _ = service.CreateReservation(ctx, id, "guest-001", "room-101", serviceValidDateRange(), serviceValidMoney(), serviceValidGuests())

	// Act
	err := service.CancelReservationOnPaymentFailed(ctx, id, reason)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	res, _ := repo.Read(ctx, id)
	assert.That(t, "status must be cancelled", res.Status, reservation.StatusCancelled)
}
