package outbound_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/outbound"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/booking"
)

// ============================================================================
// RepositoryAvailabilityChecker Tests
// ============================================================================

const testResID001 = "res-001"

type mockReservationRepo struct {
	reservations map[booking.ReservationID]booking.Reservation
	readAllErr   error
}

func newMockReservationRepo() *mockReservationRepo {
	return &mockReservationRepo{
		reservations: make(map[booking.ReservationID]booking.Reservation),
	}
}

func (m *mockReservationRepo) Create(ctx context.Context, id booking.ReservationID, res booking.Reservation) error {
	m.reservations[id] = res
	return nil
}

func (m *mockReservationRepo) Read(ctx context.Context, id booking.ReservationID) (*booking.Reservation, error) {
	res, ok := m.reservations[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return &res, nil
}

func (m *mockReservationRepo) Update(ctx context.Context, id booking.ReservationID, res booking.Reservation) error {
	m.reservations[id] = res
	return nil
}

func (m *mockReservationRepo) Delete(ctx context.Context, id booking.ReservationID) error {
	delete(m.reservations, id)
	return nil
}

func (m *mockReservationRepo) ReadAll(ctx context.Context) ([]booking.Reservation, error) {
	if m.readAllErr != nil {
		return nil, m.readAllErr
	}
	result := make([]booking.Reservation, 0, len(m.reservations))
	for _, res := range m.reservations {
		result = append(result, res)
	}
	return result, nil
}

func createTestReservationInRepo(repo *mockReservationRepo, id string, roomID string, checkInDays, checkOutDays int) {
	checkIn := time.Now().AddDate(0, 0, checkInDays)
	checkOut := time.Now().AddDate(0, 0, checkOutDays)

	res := booking.Reservation{
		ID:        booking.ReservationID(id),
		GuestID:   "guest-001",
		RoomID:    booking.RoomID(roomID),
		DateRange: booking.NewDateRange(checkIn, checkOut),
		Status:    booking.StatusPending,
		TotalAmount: booking.NewMoney(30000, "USD"),
		Guests: []booking.GuestInfo{
			booking.NewGuestInfo("John Doe", "john@example.com", "+1234567890"),
		},
	}
	repo.reservations[booking.ReservationID(id)] = res
}

func Test_RepositoryAvailabilityChecker_IsRoomAvailable_No_Reservations_Should_Return_True(t *testing.T) {
	// Arrange
	repo := newMockReservationRepo()
	checker := outbound.NewRepositoryAvailabilityChecker(repo)
	ctx := context.Background()

	checkIn := time.Now().AddDate(0, 0, 7)
	checkOut := time.Now().AddDate(0, 0, 10)
	dateRange := booking.NewDateRange(checkIn, checkOut)

	// Act
	available, err := checker.IsRoomAvailable(ctx, "room-101", dateRange)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "room must be available", available, true)
}

func Test_RepositoryAvailabilityChecker_IsRoomAvailable_With_Overlapping_Reservation_Should_Return_False(t *testing.T) {
	// Arrange
	repo := newMockReservationRepo()
	checker := outbound.NewRepositoryAvailabilityChecker(repo)
	ctx := context.Background()

	createTestReservationInRepo(repo, testResID001, "room-101", 7, 10)

	checkIn := time.Now().AddDate(0, 0, 8)
	checkOut := time.Now().AddDate(0, 0, 12)
	dateRange := booking.NewDateRange(checkIn, checkOut)

	// Act
	available, err := checker.IsRoomAvailable(ctx, "room-101", dateRange)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "room must be unavailable due to overlap", available, false)
}

func Test_RepositoryAvailabilityChecker_IsRoomAvailable_Different_Room_Should_Return_True(t *testing.T) {
	// Arrange
	repo := newMockReservationRepo()
	checker := outbound.NewRepositoryAvailabilityChecker(repo)
	ctx := context.Background()

	createTestReservationInRepo(repo, testResID001, "room-101", 7, 10)

	checkIn := time.Now().AddDate(0, 0, 7)
	checkOut := time.Now().AddDate(0, 0, 10)
	dateRange := booking.NewDateRange(checkIn, checkOut)

	// Act
	available, err := checker.IsRoomAvailable(ctx, "room-102", dateRange)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "different room must be available", available, true)
}

func Test_RepositoryAvailabilityChecker_IsRoomAvailable_SameDay_Checkout_Checkin_Should_Return_True(t *testing.T) {
	// Arrange
	repo := newMockReservationRepo()
	checker := outbound.NewRepositoryAvailabilityChecker(repo)
	ctx := context.Background()

	createTestReservationInRepo(repo, testResID001, "room-101", 7, 10)

	checkIn := time.Now().AddDate(0, 0, 10)
	checkOut := time.Now().AddDate(0, 0, 13)
	dateRange := booking.NewDateRange(checkIn, checkOut)

	// Act
	available, err := checker.IsRoomAvailable(ctx, "room-101", dateRange)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "room must be available for same-day checkout/check-in", available, true)
}

func Test_RepositoryAvailabilityChecker_IsRoomAvailable_Repository_Error_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newMockReservationRepo()
	repo.readAllErr = errors.New("database error")
	checker := outbound.NewRepositoryAvailabilityChecker(repo)
	ctx := context.Background()

	checkIn := time.Now().AddDate(0, 0, 7)
	checkOut := time.Now().AddDate(0, 0, 10)
	dateRange := booking.NewDateRange(checkIn, checkOut)

	// Act
	_, err := checker.IsRoomAvailable(ctx, "room-101", dateRange)

	// Assert
	assert.That(t, "error must not be nil for repository failure", err != nil, true)
}

func Test_RepositoryAvailabilityChecker_GetOverlappingReservations_Should_Return_Overlapping(t *testing.T) {
	// Arrange
	repo := newMockReservationRepo()
	checker := outbound.NewRepositoryAvailabilityChecker(repo)
	ctx := context.Background()

	createTestReservationInRepo(repo, testResID001, "room-101", 7, 10)
	createTestReservationInRepo(repo, "res-002", "room-101", 14, 17)
	createTestReservationInRepo(repo, "res-003", "room-102", 7, 10)

	checkIn := time.Now().AddDate(0, 0, 8)
	checkOut := time.Now().AddDate(0, 0, 12)
	dateRange := booking.NewDateRange(checkIn, checkOut)

	// Act
	overlapping, err := checker.GetOverlappingReservations(ctx, "room-101", dateRange)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "must return 1 overlapping reservation", len(overlapping), 1)

	if len(overlapping) > 0 {
		assert.That(t, "overlapping reservation ID must match", string(overlapping[0].ID), testResID001)
	}
}

func Test_RepositoryAvailabilityChecker_GetOverlappingReservations_No_Overlaps_Should_Return_Empty(t *testing.T) {
	// Arrange
	repo := newMockReservationRepo()
	checker := outbound.NewRepositoryAvailabilityChecker(repo)
	ctx := context.Background()

	createTestReservationInRepo(repo, testResID001, "room-101", 7, 10)

	checkIn := time.Now().AddDate(0, 0, 14)
	checkOut := time.Now().AddDate(0, 0, 17)
	dateRange := booking.NewDateRange(checkIn, checkOut)

	// Act
	overlapping, err := checker.GetOverlappingReservations(ctx, "room-101", dateRange)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "must return 0 overlapping reservations", len(overlapping), 0)
}

func Test_RepositoryAvailabilityChecker_IsRoomAvailable_Cancelled_Reservation_Should_Return_True(t *testing.T) {
	// Arrange
	repo := newMockReservationRepo()
	checker := outbound.NewRepositoryAvailabilityChecker(repo)
	ctx := context.Background()

	createTestReservationInRepo(repo, testResID001, "room-101", 7, 10)
	res := repo.reservations[testResID001]
	res.Status = booking.StatusCancelled
	repo.reservations[testResID001] = res

	checkIn := time.Now().AddDate(0, 0, 7)
	checkOut := time.Now().AddDate(0, 0, 10)
	dateRange := booking.NewDateRange(checkIn, checkOut)

	// Act
	available, err := checker.IsRoomAvailable(ctx, "room-101", dateRange)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "room must be available when existing reservation is cancelled", available, true)
}
