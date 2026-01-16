package outbound_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/outbound"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/booking"
)

// ============================================================================
// FileReservationRepository Tests
// ============================================================================

const testGuestID002 = "guest-002"

func createTempFile(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	return filepath.Join(tmpDir, "reservations.json")
}

func createSampleReservation(id string) booking.Reservation {
	checkIn := time.Now().AddDate(0, 0, 7)
	checkOut := time.Now().AddDate(0, 0, 10)

	return booking.Reservation{
		ID:        booking.ReservationID(id),
		GuestID:   "guest-001",
		RoomID:    "room-101",
		DateRange: booking.NewDateRange(checkIn, checkOut),
		Status:    booking.StatusPending,
		TotalAmount: booking.NewMoney(30000, "USD"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Guests: []booking.GuestInfo{
			booking.NewGuestInfo("John Doe", "john@example.com", "+1234567890"),
		},
	}
}

func Test_FileReservationRepository_Create_And_Read_Should_Succeed(t *testing.T) {
	// Arrange
	filename := createTempFile(t)
	repo := outbound.NewFileReservationRepository(filename)
	ctx := context.Background()
	reservation := createSampleReservation("res-001")

	// Act
	err := repo.Create(ctx, "res-001", reservation)

	// Assert
	assert.That(t, "error on create must be nil", err == nil, true)

	read, readErr := repo.Read(ctx, "res-001")
	assert.That(t, "error on read must be nil", readErr == nil, true)
	assert.That(t, "ID must match", string(read.ID), "res-001")
	assert.That(t, "GuestID must match", string(read.GuestID), "guest-001")
}

func Test_FileReservationRepository_Update_Should_Modify_Reservation(t *testing.T) {
	// Arrange
	filename := createTempFile(t)
	repo := outbound.NewFileReservationRepository(filename)
	ctx := context.Background()
	reservation := createSampleReservation("res-001")

	setupErr := repo.Create(ctx, "res-001", reservation)
	if setupErr != nil {
		t.Fatalf("setup failed: %v", setupErr)
	}

	reservation.Status = booking.StatusConfirmed

	// Act
	err := repo.Update(ctx, "res-001", reservation)

	// Assert
	assert.That(t, "error on update must be nil", err == nil, true)

	read, readErr := repo.Read(ctx, "res-001")
	assert.That(t, "error on read must be nil", readErr == nil, true)
	assert.That(t, "status must be confirmed", read.Status, booking.StatusConfirmed)
}

func Test_FileReservationRepository_Delete_Should_Remove_Reservation(t *testing.T) {
	// Arrange
	filename := createTempFile(t)
	repo := outbound.NewFileReservationRepository(filename)
	ctx := context.Background()
	reservation := createSampleReservation("res-001")

	setupErr := repo.Create(ctx, "res-001", reservation)
	if setupErr != nil {
		t.Fatalf("setup failed: %v", setupErr)
	}

	// Act
	err := repo.Delete(ctx, "res-001")

	// Assert
	assert.That(t, "error on delete must be nil", err == nil, true)

	_, readErr := repo.Read(ctx, "res-001")
	assert.That(t, "error must not be nil for deleted reservation", readErr != nil, true)
}

func Test_FileReservationRepository_ReadAll_Should_Return_All_Reservations(t *testing.T) {
	// Arrange
	filename := createTempFile(t)
	repo := outbound.NewFileReservationRepository(filename)
	ctx := context.Background()

	res1 := createSampleReservation("res-001")
	res2 := createSampleReservation("res-002")
	res2.GuestID = testGuestID002

	_ = repo.Create(ctx, "res-001", res1)
	_ = repo.Create(ctx, "res-002", res2)

	// Act
	all, err := repo.ReadAll(ctx)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "must return 2 reservations", len(all), 2)
}

func Test_FileReservationRepository_Read_NonExistent_Should_Return_Error(t *testing.T) {
	// Arrange
	filename := createTempFile(t)
	repo := outbound.NewFileReservationRepository(filename)
	ctx := context.Background()

	// Act
	_, err := repo.Read(ctx, "nonexistent")

	// Assert
	assert.That(t, "error must not be nil for nonexistent reservation", err != nil, true)
}

func Test_FileReservationRepository_Persistence_Across_Instances(t *testing.T) {
	// Arrange
	filename := createTempFile(t)
	ctx := context.Background()

	repo1 := outbound.NewFileReservationRepository(filename)
	reservation := createSampleReservation("res-001")
	_ = repo1.Create(ctx, "res-001", reservation)

	// Act
	repo2 := outbound.NewFileReservationRepository(filename)
	read, err := repo2.Read(ctx, "res-001")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "ID must match", string(read.ID), "res-001")
}

func Test_FileReservationRepository_Empty_File_ReadAll_Should_Return_Empty(t *testing.T) {
	// Arrange
	filename := createTempFile(t)
	repo := outbound.NewFileReservationRepository(filename)
	ctx := context.Background()

	// Act
	all, err := repo.ReadAll(ctx)

	// Assert
	if err != nil {
		// This is acceptable for an empty/non-existent file.
		return
	}

	if all == nil {
		all = []booking.Reservation{}
	}

	assert.That(t, "must return 0 reservations for empty file", len(all), 0)
}

func Test_FileReservationRepository_Update_Should_Replace_Existing(t *testing.T) {
	// Arrange
	filename := createTempFile(t)
	repo := outbound.NewFileReservationRepository(filename)
	ctx := context.Background()

	res1 := createSampleReservation("res-001")
	res1.GuestID = "guest-001"
	_ = repo.Create(ctx, "res-001", res1)

	res2 := createSampleReservation("res-001")
	res2.GuestID = testGuestID002

	// Act
	_ = repo.Update(ctx, "res-001", res2)
	read, err := repo.Read(ctx, "res-001")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "GuestID must be updated", string(read.GuestID), testGuestID002)
}

func Test_FileReservationRepository_Cleanup(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "reservations.json")
	repo := outbound.NewFileReservationRepository(filename)
	ctx := context.Background()
	reservation := createSampleReservation("res-001")

	// Act
	_ = repo.Create(ctx, "res-001", reservation)

	// Assert
	_, err := os.Stat(filename)
	assert.That(t, "file must exist after create", !os.IsNotExist(err), true)
}
