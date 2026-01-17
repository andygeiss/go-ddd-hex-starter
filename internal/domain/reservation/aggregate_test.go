package reservation_test

import (
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
	"github.com/andygeiss/hotel-booking/internal/domain/shared"
)

// ============================================================================
// Test Helpers
// ============================================================================

func validDateRange() reservation.DateRange {
	checkIn := time.Now().Add(48 * time.Hour).Truncate(24 * time.Hour)
	checkOut := checkIn.Add(72 * time.Hour)
	return reservation.NewDateRange(checkIn, checkOut)
}

func validGuests() []reservation.GuestInfo {
	return []reservation.GuestInfo{
		reservation.NewGuestInfo("John Doe", "john@example.com", "+1234567890"),
	}
}

func validMoney() shared.Money {
	return shared.NewMoney(10000, "USD")
}

func createValidReservation(t *testing.T) *reservation.Reservation {
	t.Helper()
	res, err := reservation.NewReservation(
		"res-001",
		"guest-001",
		"room-101",
		validDateRange(),
		validMoney(),
		validGuests(),
	)
	if err != nil {
		t.Fatalf("failed to create valid reservation: %v", err)
	}
	return res
}

// ============================================================================
// NewReservation Tests
// ============================================================================

func Test_NewReservation_Should_Return_Valid_Reservation(t *testing.T) {
	// Arrange
	id := reservation.ReservationID("res-001")
	guestID := reservation.GuestID("guest-001")
	roomID := reservation.RoomID("room-101")
	dateRange := validDateRange()
	amount := validMoney()
	guests := validGuests()

	// Act
	res, err := reservation.NewReservation(id, guestID, roomID, dateRange, amount, guests)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "reservation must not be nil", res != nil, true)
	assert.That(t, "ID must match", res.ID, id)
	assert.That(t, "GuestID must match", res.GuestID, guestID)
	assert.That(t, "RoomID must match", res.RoomID, roomID)
	assert.That(t, "Status must be pending", res.Status, reservation.StatusPending)
	assert.That(t, "TotalAmount must match", res.TotalAmount, amount)
	assert.That(t, "Guests count must be 1", len(res.Guests), 1)
}

func Test_NewReservation_With_No_Guests_Should_Return_Error(t *testing.T) {
	// Arrange
	dateRange := validDateRange()
	amount := validMoney()

	// Act
	res, err := reservation.NewReservation(
		"res-001",
		"guest-001",
		"room-101",
		dateRange,
		amount,
		[]reservation.GuestInfo{}, // empty guests
	)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "reservation must be nil", res == nil, true)
}

func Test_NewReservation_With_CheckOut_Before_CheckIn_Should_Return_Error(t *testing.T) {
	// Arrange
	checkIn := time.Now().Add(48 * time.Hour)
	checkOut := checkIn.Add(-24 * time.Hour) // before check-in
	dateRange := reservation.NewDateRange(checkIn, checkOut)

	// Act
	res, err := reservation.NewReservation(
		"res-001",
		"guest-001",
		"room-101",
		dateRange,
		validMoney(),
		validGuests(),
	)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "reservation must be nil", res == nil, true)
}

func Test_NewReservation_With_Same_CheckIn_And_CheckOut_Should_Return_Error(t *testing.T) {
	// Arrange
	checkIn := time.Now().Add(48 * time.Hour).Truncate(24 * time.Hour)
	dateRange := reservation.NewDateRange(checkIn, checkIn) // same date

	// Act
	res, err := reservation.NewReservation(
		"res-001",
		"guest-001",
		"room-101",
		dateRange,
		validMoney(),
		validGuests(),
	)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "reservation must be nil", res == nil, true)
}

func Test_NewReservation_With_CheckIn_In_Past_Should_Return_Error(t *testing.T) {
	// Arrange
	checkIn := time.Now().Add(-48 * time.Hour) // in the past
	checkOut := time.Now().Add(24 * time.Hour)
	dateRange := reservation.NewDateRange(checkIn, checkOut)

	// Act
	res, err := reservation.NewReservation(
		"res-001",
		"guest-001",
		"room-101",
		dateRange,
		validMoney(),
		validGuests(),
	)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "reservation must be nil", res == nil, true)
}

// ============================================================================
// State Transition Tests - Confirm
// ============================================================================

func Test_Reservation_Confirm_From_Pending_Should_Succeed(t *testing.T) {
	// Arrange
	res := createValidReservation(t)
	assert.That(t, "initial status must be pending", res.Status, reservation.StatusPending)

	// Act
	err := res.Confirm()

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "status must be confirmed", res.Status, reservation.StatusConfirmed)
}

func Test_Reservation_Confirm_From_Confirmed_Should_Return_Error(t *testing.T) {
	// Arrange
	res := createValidReservation(t)
	_ = res.Confirm() // move to confirmed

	// Act
	err := res.Confirm()

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "status must remain confirmed", res.Status, reservation.StatusConfirmed)
}

func Test_Reservation_Confirm_From_Cancelled_Should_Return_Error(t *testing.T) {
	// Arrange
	res := createValidReservation(t)
	_ = res.Cancel("test reason")

	// Act
	err := res.Confirm()

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "status must remain cancelled", res.Status, reservation.StatusCancelled)
}

// ============================================================================
// State Transition Tests - Activate
// ============================================================================

func Test_Reservation_Activate_From_Confirmed_Should_Succeed(t *testing.T) {
	// Arrange
	res := createValidReservation(t)
	_ = res.Confirm()
	assert.That(t, "status must be confirmed", res.Status, reservation.StatusConfirmed)

	// Act
	err := res.Activate()

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "status must be active", res.Status, reservation.StatusActive)
}

func Test_Reservation_Activate_From_Pending_Should_Return_Error(t *testing.T) {
	// Arrange
	res := createValidReservation(t)

	// Act
	err := res.Activate()

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "status must remain pending", res.Status, reservation.StatusPending)
}

func Test_Reservation_Activate_From_Active_Should_Return_Error(t *testing.T) {
	// Arrange
	res := createValidReservation(t)
	_ = res.Confirm()
	_ = res.Activate()

	// Act
	err := res.Activate()

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "status must remain active", res.Status, reservation.StatusActive)
}

// ============================================================================
// State Transition Tests - Complete
// ============================================================================

func Test_Reservation_Complete_From_Active_Should_Succeed(t *testing.T) {
	// Arrange
	res := createValidReservation(t)
	_ = res.Confirm()
	_ = res.Activate()
	assert.That(t, "status must be active", res.Status, reservation.StatusActive)

	// Act
	err := res.Complete()

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "status must be completed", res.Status, reservation.StatusCompleted)
}

func Test_Reservation_Complete_From_Pending_Should_Return_Error(t *testing.T) {
	// Arrange
	res := createValidReservation(t)

	// Act
	err := res.Complete()

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "status must remain pending", res.Status, reservation.StatusPending)
}

func Test_Reservation_Complete_From_Confirmed_Should_Return_Error(t *testing.T) {
	// Arrange
	res := createValidReservation(t)
	_ = res.Confirm()

	// Act
	err := res.Complete()

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "status must remain confirmed", res.Status, reservation.StatusConfirmed)
}

// ============================================================================
// State Transition Tests - Cancel
// ============================================================================

func Test_Reservation_Cancel_From_Pending_Should_Succeed(t *testing.T) {
	// Arrange
	res := createValidReservation(t)
	reason := "Guest requested cancellation"

	// Act
	err := res.Cancel(reason)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "status must be cancelled", res.Status, reservation.StatusCancelled)
	assert.That(t, "cancellation reason must match", res.CancellationReason, reason)
}

func Test_Reservation_Cancel_From_Confirmed_Should_Succeed(t *testing.T) {
	// Arrange
	res := createValidReservation(t)
	_ = res.Confirm()
	reason := "Guest requested cancellation"

	// Act
	err := res.Cancel(reason)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "status must be cancelled", res.Status, reservation.StatusCancelled)
}

func Test_Reservation_Cancel_From_Active_Should_Return_Error(t *testing.T) {
	// Arrange
	res := createValidReservation(t)
	_ = res.Confirm()
	_ = res.Activate()

	// Act
	err := res.Cancel("test")

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "status must remain active", res.Status, reservation.StatusActive)
}

func Test_Reservation_Cancel_From_Completed_Should_Return_Error(t *testing.T) {
	// Arrange
	res := createValidReservation(t)
	_ = res.Confirm()
	_ = res.Activate()
	_ = res.Complete()

	// Act
	err := res.Cancel("test")

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "status must remain completed", res.Status, reservation.StatusCompleted)
}

func Test_Reservation_Cancel_Already_Cancelled_Should_Return_Error(t *testing.T) {
	// Arrange
	res := createValidReservation(t)
	_ = res.Cancel("first cancellation")

	// Act
	err := res.Cancel("second cancellation")

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "cancellation reason must be first", res.CancellationReason, "first cancellation")
}

// ============================================================================
// Business Logic Tests
// ============================================================================

func Test_Reservation_Nights_Should_Return_Correct_Value(t *testing.T) {
	// Arrange
	res := createValidReservation(t)

	// Act
	nights := res.Nights()

	// Assert
	assert.That(t, "nights must be 3", nights, 3)
}

func Test_Reservation_CanBeCancelled_For_Pending_Far_Future_Should_Return_True(t *testing.T) {
	// Arrange
	res := createValidReservation(t)

	// Act
	canCancel := res.CanBeCancelled()

	// Assert
	assert.That(t, "should be cancellable", canCancel, true)
}

func Test_Reservation_CanBeCancelled_For_Active_Should_Return_False(t *testing.T) {
	// Arrange
	res := createValidReservation(t)
	_ = res.Confirm()
	_ = res.Activate()

	// Act
	canCancel := res.CanBeCancelled()

	// Assert
	assert.That(t, "should not be cancellable", canCancel, false)
}

func Test_Reservation_CanBeCancelled_For_Completed_Should_Return_False(t *testing.T) {
	// Arrange
	res := createValidReservation(t)
	_ = res.Confirm()
	_ = res.Activate()
	_ = res.Complete()

	// Act
	canCancel := res.CanBeCancelled()

	// Assert
	assert.That(t, "should not be cancellable", canCancel, false)
}

func Test_Reservation_IsOverlapping_Same_Room_Overlapping_Dates_Should_Return_True(t *testing.T) {
	// Arrange
	checkIn1 := time.Now().Add(48 * time.Hour).Truncate(24 * time.Hour)
	checkOut1 := checkIn1.Add(72 * time.Hour)
	dateRange1 := reservation.NewDateRange(checkIn1, checkOut1)

	checkIn2 := checkIn1.Add(24 * time.Hour) // overlaps
	checkOut2 := checkIn2.Add(72 * time.Hour)
	dateRange2 := reservation.NewDateRange(checkIn2, checkOut2)

	res1, _ := reservation.NewReservation("res-001", "guest-001", "room-101", dateRange1, validMoney(), validGuests())
	res2, _ := reservation.NewReservation("res-002", "guest-002", "room-101", dateRange2, validMoney(), validGuests())

	// Act
	overlapping := res1.IsOverlapping(res2)

	// Assert
	assert.That(t, "should be overlapping", overlapping, true)
}

func Test_Reservation_IsOverlapping_Different_Room_Should_Return_False(t *testing.T) {
	// Arrange
	checkIn := time.Now().Add(48 * time.Hour).Truncate(24 * time.Hour)
	checkOut := checkIn.Add(72 * time.Hour)
	dateRange := reservation.NewDateRange(checkIn, checkOut)

	res1, _ := reservation.NewReservation("res-001", "guest-001", "room-101", dateRange, validMoney(), validGuests())
	res2, _ := reservation.NewReservation("res-002", "guest-002", "room-102", dateRange, validMoney(), validGuests())

	// Act
	overlapping := res1.IsOverlapping(res2)

	// Assert
	assert.That(t, "should not be overlapping", overlapping, false)
}

func Test_Reservation_IsOverlapping_One_Cancelled_Should_Return_False(t *testing.T) {
	// Arrange
	checkIn := time.Now().Add(48 * time.Hour).Truncate(24 * time.Hour)
	checkOut := checkIn.Add(72 * time.Hour)
	dateRange := reservation.NewDateRange(checkIn, checkOut)

	res1, _ := reservation.NewReservation("res-001", "guest-001", "room-101", dateRange, validMoney(), validGuests())
	res2, _ := reservation.NewReservation("res-002", "guest-002", "room-101", dateRange, validMoney(), validGuests())
	_ = res2.Cancel("cancelled")

	// Act
	overlapping := res1.IsOverlapping(res2)

	// Assert
	assert.That(t, "should not be overlapping", overlapping, false)
}

// ============================================================================
// Value Object Tests - DateRange
// ============================================================================

func Test_NewDateRange_Should_Create_Valid_DateRange(t *testing.T) {
	// Arrange
	checkIn := time.Date(2024, 6, 1, 14, 0, 0, 0, time.UTC)
	checkOut := time.Date(2024, 6, 5, 12, 0, 0, 0, time.UTC)

	// Act
	dateRange := reservation.NewDateRange(checkIn, checkOut)

	// Assert
	assert.That(t, "CheckIn must match", dateRange.CheckIn, checkIn)
	assert.That(t, "CheckOut must match", dateRange.CheckOut, checkOut)
}

// ============================================================================
// Value Object Tests - GuestInfo
// ============================================================================

func Test_NewGuestInfo_Should_Create_Valid_GuestInfo(t *testing.T) {
	// Arrange
	name := "John Doe"
	email := "john@example.com"
	phone := "+1234567890"

	// Act
	guest := reservation.NewGuestInfo(name, email, phone)

	// Assert
	assert.That(t, "Name must match", guest.Name, name)
	assert.That(t, "Email must match", guest.Email, email)
	assert.That(t, "PhoneNumber must match", guest.PhoneNumber, phone)
}

// ============================================================================
// Value Object Tests - Money (shared)
// ============================================================================

func Test_NewMoney_Should_Create_Valid_Money(t *testing.T) {
	// Arrange
	amount := int64(10000)
	currency := "usd"

	// Act
	money := shared.NewMoney(amount, currency)

	// Assert
	assert.That(t, "Amount must match", money.Amount, amount)
	assert.That(t, "Currency must be uppercase", money.Currency, "USD")
}

func Test_Money_FormatAmount_Should_Return_Formatted_String(t *testing.T) {
	// Arrange
	money := shared.NewMoney(10050, "USD")

	// Act
	formatted := money.FormatAmount()

	// Assert
	assert.That(t, "formatted must be correct", formatted, "100.50 USD")
}

// ============================================================================
// Additional Coverage Tests
// ============================================================================

func Test_Reservation_DaysUntilCheckIn_Should_Return_Correct_Value(t *testing.T) {
	// Arrange
	res := createValidReservation(t)

	// Act
	days := res.DaysUntilCheckIn()

	// Assert
	assert.That(t, "days until check-in must be approximately 2", days >= 1 && days <= 3, true)
}

func Test_Reservation_Cancel_Near_CheckIn_Should_Return_Error(t *testing.T) {
	// Arrange - create reservation with check-in very soon (less than 24 hours)
	checkIn := time.Now().Add(12 * time.Hour).Truncate(24 * time.Hour)
	if checkIn.Before(time.Now()) {
		checkIn = time.Now().Add(6 * time.Hour)
	}
	checkOut := checkIn.Add(72 * time.Hour)
	dateRange := reservation.NewDateRange(checkIn, checkOut)

	// Create reservation directly to bypass date validation
	res := &reservation.Reservation{
		ID:        "res-near",
		GuestID:   "guest-001",
		RoomID:    "room-101",
		DateRange: dateRange,
		Status:    reservation.StatusPending,
		Guests:    validGuests(),
	}

	// Act
	err := res.Cancel("test")

	// Assert
	assert.That(t, "error must not be nil for near check-in", err != nil, true)
}

// ============================================================================
// Event Topic Tests - Reservation
// ============================================================================

func Test_EventCreated_Topic_Should_Return_Correct_Value(t *testing.T) {
	// Arrange
	evt := reservation.NewEventCreated()

	// Act
	topic := evt.Topic()

	// Assert
	assert.That(t, "topic must be reservation.created", topic, "reservation.created")
}

func Test_EventConfirmed_Topic_Should_Return_Correct_Value(t *testing.T) {
	// Arrange
	evt := reservation.NewEventConfirmed()

	// Act
	topic := evt.Topic()

	// Assert
	assert.That(t, "topic must be reservation.confirmed", topic, "reservation.confirmed")
}

func Test_EventActivated_Topic_Should_Return_Correct_Value(t *testing.T) {
	// Arrange
	evt := reservation.NewEventActivated()

	// Act
	topic := evt.Topic()

	// Assert
	assert.That(t, "topic must be reservation.activated", topic, "reservation.activated")
}

func Test_EventCompleted_Topic_Should_Return_Correct_Value(t *testing.T) {
	// Arrange
	evt := reservation.NewEventCompleted()

	// Act
	topic := evt.Topic()

	// Assert
	assert.That(t, "topic must be reservation.completed", topic, "reservation.completed")
}

func Test_EventCancelled_Topic_Should_Return_Correct_Value(t *testing.T) {
	// Arrange
	evt := reservation.NewEventCancelled()

	// Act
	topic := evt.Topic()

	// Assert
	assert.That(t, "topic must be reservation.cancelled", topic, "reservation.cancelled")
}
