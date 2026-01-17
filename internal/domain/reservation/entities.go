package reservation

import "time"

// DateRange represents a time period for a reservation.
type DateRange struct {
	CheckIn  time.Time
	CheckOut time.Time
}

// NewDateRange creates a DateRange value object.
func NewDateRange(checkIn, checkOut time.Time) DateRange {
	return DateRange{
		CheckIn:  checkIn,
		CheckOut: checkOut,
	}
}

// GuestInfo represents information about a guest (entity within Reservation aggregate).
type GuestInfo struct {
	Name        string
	Email       string
	PhoneNumber string
}

// NewGuestInfo creates a GuestInfo entity.
func NewGuestInfo(name, email, phoneNumber string) GuestInfo {
	return GuestInfo{
		Name:        name,
		Email:       email,
		PhoneNumber: phoneNumber,
	}
}
