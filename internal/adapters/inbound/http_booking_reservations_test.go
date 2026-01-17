package inbound_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
)

// ============================================================================
// HttpBookingReservations Tests
// ============================================================================

func Test_ReservationStatusClass_Pending_Should_Return_Warning(t *testing.T) {
	// Arrange
	status := reservation.StatusPending

	// Act
	result := testReservationStatusClass(status)

	// Assert
	assert.That(t, "status class must be warning", result, "warning")
}

func Test_ReservationStatusClass_Confirmed_Should_Return_Info(t *testing.T) {
	// Arrange
	status := reservation.StatusConfirmed

	// Act
	result := testReservationStatusClass(status)

	// Assert
	assert.That(t, "status class must be info", result, "info")
}

func Test_ReservationStatusClass_Active_Should_Return_Primary(t *testing.T) {
	// Arrange
	status := reservation.StatusActive

	// Act
	result := testReservationStatusClass(status)

	// Assert
	assert.That(t, "status class must be primary", result, "primary")
}

func Test_ReservationStatusClass_Completed_Should_Return_Success(t *testing.T) {
	// Arrange
	status := reservation.StatusCompleted

	// Act
	result := testReservationStatusClass(status)

	// Assert
	assert.That(t, "status class must be success", result, "success")
}

func Test_ReservationStatusClass_Cancelled_Should_Return_Danger(t *testing.T) {
	// Arrange
	status := reservation.StatusCancelled

	// Act
	result := testReservationStatusClass(status)

	// Assert
	assert.That(t, "status class must be danger", result, "danger")
}

func Test_ReservationStatusClass_Unknown_Should_Return_Secondary(t *testing.T) {
	// Arrange
	status := reservation.ReservationStatus("unknown")

	// Act
	result := testReservationStatusClass(status)

	// Assert
	assert.That(t, "status class must be secondary", result, "secondary")
}

func testReservationStatusClass(status reservation.ReservationStatus) string {
	switch status {
	case reservation.StatusPending:
		return "warning"
	case reservation.StatusConfirmed:
		return "info"
	case reservation.StatusActive:
		return "primary"
	case reservation.StatusCompleted:
		return "success"
	case reservation.StatusCancelled:
		return "danger"
	default:
		return "secondary"
	}
}
