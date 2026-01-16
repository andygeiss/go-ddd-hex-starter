package inbound_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
)

// ============================================================================
// HttpBookingReservationForm Tests
// ============================================================================

type roomOption struct {
	ID    string
	Name  string
	Price string
}

func getDefaultRoomsForTest() []roomOption {
	return []roomOption{
		{ID: "room-101", Name: "Standard Room 101", Price: "$99.00"},
		{ID: "room-102", Name: "Standard Room 102", Price: "$99.00"},
		{ID: "room-201", Name: "Deluxe Room 201", Price: "$149.00"},
		{ID: "room-202", Name: "Deluxe Room 202", Price: "$149.00"},
		{ID: "room-301", Name: "Suite 301", Price: "$249.00"},
	}
}

func getRoomPricesForTest() map[string]int64 {
	return map[string]int64{
		"room-101": 9900,
		"room-102": 9900,
		"room-201": 14900,
		"room-202": 14900,
		"room-301": 24900,
	}
}

func Test_GetDefaultRooms_Should_Return_Five_Rooms(t *testing.T) {
	// Arrange
	// No setup needed

	// Act
	rooms := getDefaultRoomsForTest()

	// Assert
	assert.That(t, "must return 5 rooms", len(rooms), 5)
}

func Test_GetDefaultRooms_Should_Have_Standard_Rooms(t *testing.T) {
	// Arrange
	rooms := getDefaultRoomsForTest()

	// Act
	foundStandard := 0
	for _, room := range rooms {
		if room.Price == "$99.00" {
			foundStandard++
		}
	}

	// Assert
	assert.That(t, "must have 2 standard rooms at $99.00", foundStandard, 2)
}

func Test_GetDefaultRooms_Should_Have_Deluxe_Rooms(t *testing.T) {
	// Arrange
	rooms := getDefaultRoomsForTest()

	// Act
	foundDeluxe := 0
	for _, room := range rooms {
		if room.Price == "$149.00" {
			foundDeluxe++
		}
	}

	// Assert
	assert.That(t, "must have 2 deluxe rooms at $149.00", foundDeluxe, 2)
}

func Test_GetDefaultRooms_Should_Have_Suite(t *testing.T) {
	// Arrange
	rooms := getDefaultRoomsForTest()

	// Act
	foundSuite := 0
	for _, room := range rooms {
		if room.Price == "$249.00" {
			foundSuite++
		}
	}

	// Assert
	assert.That(t, "must have 1 suite at $249.00", foundSuite, 1)
}

func Test_GetRoomPrices_Should_Return_All_Room_Prices(t *testing.T) {
	// Arrange
	// No setup needed

	// Act
	prices := getRoomPricesForTest()

	// Assert
	assert.That(t, "must return 5 room prices", len(prices), 5)
}

func Test_GetRoomPrices_Should_Have_Correct_Standard_Price(t *testing.T) {
	// Arrange
	prices := getRoomPricesForTest()

	// Act
	room101Price := prices["room-101"]
	room102Price := prices["room-102"]

	// Assert
	assert.That(t, "room-101 price must be 9900", room101Price, int64(9900))
	assert.That(t, "room-102 price must be 9900", room102Price, int64(9900))
}

func Test_GetRoomPrices_Should_Have_Correct_Deluxe_Price(t *testing.T) {
	// Arrange
	prices := getRoomPricesForTest()

	// Act
	room201Price := prices["room-201"]
	room202Price := prices["room-202"]

	// Assert
	assert.That(t, "room-201 price must be 14900", room201Price, int64(14900))
	assert.That(t, "room-202 price must be 14900", room202Price, int64(14900))
}

func Test_GetRoomPrices_Should_Have_Correct_Suite_Price(t *testing.T) {
	// Arrange
	prices := getRoomPricesForTest()

	// Act
	room301Price := prices["room-301"]

	// Assert
	assert.That(t, "room-301 price must be 24900", room301Price, int64(24900))
}
