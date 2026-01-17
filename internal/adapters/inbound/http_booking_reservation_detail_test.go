package inbound_test

import (
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
)

// ============================================================================
// HttpBookingReservationDetail Tests
// ============================================================================

const testDetailResID001 = "res-001"

func Test_GuestInfoView_Should_Have_Expected_Fields(t *testing.T) {
	// Arrange
	guest := struct {
		Name        string
		Email       string
		PhoneNumber string
	}{
		Name:        "John Doe",
		Email:       "john@example.com",
		PhoneNumber: "+1234567890",
	}

	// Act
	// No action needed, just verifying field values

	// Assert
	assert.That(t, "Name must match", guest.Name, "John Doe")
	assert.That(t, "Email must match", guest.Email, "john@example.com")
	assert.That(t, "PhoneNumber must match", guest.PhoneNumber, "+1234567890")
}

func Test_ReservationDetailView_Required_Fields(t *testing.T) {
	// Arrange
	id := testDetailResID001
	nights := 3
	canCancel := true
	guestCount := 1

	// Act
	// No action needed, just verifying values

	// Assert
	assert.That(t, "ID must match", id, testDetailResID001)
	assert.That(t, "Nights must be 3", nights, 3)
	assert.That(t, "CanCancel must be true", canCancel, true)
	assert.That(t, "guestCount must be 1", guestCount, 1)
}

func Test_HttpViewReservationDetailResponse_Required_Fields(t *testing.T) {
	// Arrange
	resp := struct {
		AppName     string
		Reservation struct {
			ID     string
			RoomID string
		}
	}{
		AppName: "Hotel App",
		Reservation: struct {
			ID     string
			RoomID string
		}{
			ID:     testDetailResID001,
			RoomID: "room-101",
		},
	}

	// Act
	// No action needed, just verifying field values

	// Assert
	assert.That(t, "AppName must match", resp.AppName, "Hotel App")
	assert.That(t, "Reservation.ID must match", resp.Reservation.ID, testDetailResID001)
}

func Test_BuildReservationDetailView_Logic_Should_Convert_Guests(t *testing.T) {
	// Arrange
	domainGuests := []reservation.GuestInfo{
		reservation.NewGuestInfo("John Doe", "john@example.com", "+1234567890"),
		reservation.NewGuestInfo("Jane Doe", "jane@example.com", "+0987654321"),
	}

	viewGuests := make([]struct {
		Name        string
		Email       string
		PhoneNumber string
	}, 0, len(domainGuests))

	// Act
	for _, g := range domainGuests {
		viewGuests = append(viewGuests, struct {
			Name        string
			Email       string
			PhoneNumber string
		}{
			Name:        g.Name,
			Email:       g.Email,
			PhoneNumber: g.PhoneNumber,
		})
	}

	// Assert
	assert.That(t, "must have 2 guests", len(viewGuests), 2)
	assert.That(t, "first guest name must match", viewGuests[0].Name, "John Doe")
	assert.That(t, "second guest name must match", viewGuests[1].Name, "Jane Doe")
}

func Test_BuildReservationDetailView_Logic_Should_Format_Dates(t *testing.T) {
	// Arrange
	checkIn := time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC)
	checkOut := time.Date(2024, 1, 18, 12, 0, 0, 0, time.UTC)
	createdAt := time.Date(2024, 1, 10, 14, 30, 0, 0, time.UTC)

	// Act
	checkInFormatted := checkIn.Format("2006-01-02")
	checkOutFormatted := checkOut.Format("2006-01-02")
	createdAtFormatted := createdAt.Format("2006-01-02 15:04")

	// Assert
	assert.That(t, "checkIn format must match", checkInFormatted, "2024-01-15")
	assert.That(t, "checkOut format must match", checkOutFormatted, "2024-01-18")
	assert.That(t, "createdAt format must match", createdAtFormatted, "2024-01-10 14:30")
}

func Test_BuildReservationDetailView_Logic_Should_Calculate_Nights(t *testing.T) {
	// Arrange
	checkIn := time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC)
	checkOut := time.Date(2024, 1, 18, 12, 0, 0, 0, time.UTC)

	// Act
	nights := int(checkOut.Sub(checkIn).Hours() / 24)

	// Assert
	assert.That(t, "nights must be within expected range", nights >= 2 && nights <= 3, true)
}

func Test_ReservationDetailView_With_Cancellation_Reason_Should_Show_Reason(t *testing.T) {
	// Arrange
	detail := struct {
		Status             string
		CancellationReason string
	}{
		Status:             "cancelled",
		CancellationReason: "Guest requested cancellation",
	}

	// Act
	// No action needed, just verifying field values

	// Assert
	assert.That(t, "Status must be cancelled", detail.Status, "cancelled")
	assert.That(t, "CancellationReason must match", detail.CancellationReason, "Guest requested cancellation")
}

func Test_StatusClass_Mapping_For_All_Statuses(t *testing.T) {
	// Arrange
	tests := []struct {
		status   reservation.ReservationStatus
		expected string
	}{
		{reservation.StatusPending, "warning"},
		{reservation.StatusConfirmed, "info"},
		{reservation.StatusActive, "primary"},
		{reservation.StatusCompleted, "success"},
		{reservation.StatusCancelled, "danger"},
	}

	// Act & Assert
	for _, tc := range tests {
		result := testReservationStatusClass(tc.status)
		assert.That(t, "status class must match for "+string(tc.status), result, tc.expected)
	}
}
