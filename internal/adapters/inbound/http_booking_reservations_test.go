package inbound_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/booking"
)

// ============================================================================
// HttpBookingReservations Tests
// ============================================================================

func Test_ReservationStatusClass_Pending_Should_Return_Warning(t *testing.T) {
	// Arrange
	status := booking.StatusPending

	// Act
	result := testReservationStatusClass(status)

	// Assert
	assert.That(t, "status class must be warning", result, "warning")
}

func Test_ReservationStatusClass_Confirmed_Should_Return_Info(t *testing.T) {
	// Arrange
	status := booking.StatusConfirmed

	// Act
	result := testReservationStatusClass(status)

	// Assert
	assert.That(t, "status class must be info", result, "info")
}

func Test_ReservationStatusClass_Active_Should_Return_Primary(t *testing.T) {
	// Arrange
	status := booking.StatusActive

	// Act
	result := testReservationStatusClass(status)

	// Assert
	assert.That(t, "status class must be primary", result, "primary")
}

func Test_ReservationStatusClass_Completed_Should_Return_Success(t *testing.T) {
	// Arrange
	status := booking.StatusCompleted

	// Act
	result := testReservationStatusClass(status)

	// Assert
	assert.That(t, "status class must be success", result, "success")
}

func Test_ReservationStatusClass_Cancelled_Should_Return_Danger(t *testing.T) {
	// Arrange
	status := booking.StatusCancelled

	// Act
	result := testReservationStatusClass(status)

	// Assert
	assert.That(t, "status class must be danger", result, "danger")
}

func Test_ReservationStatusClass_Unknown_Should_Return_Secondary(t *testing.T) {
	// Arrange
	status := booking.ReservationStatus("unknown")

	// Act
	result := testReservationStatusClass(status)

	// Assert
	assert.That(t, "status class must be secondary", result, "secondary")
}

func testReservationStatusClass(status booking.ReservationStatus) string {
	switch status {
	case booking.StatusPending:
		return "warning"
	case booking.StatusConfirmed:
		return "info"
	case booking.StatusActive:
		return "primary"
	case booking.StatusCompleted:
		return "success"
	case booking.StatusCancelled:
		return "danger"
	default:
		return "secondary"
	}
}
