package reservation

import (
	"context"

	"github.com/andygeiss/cloud-native-utils/event"
	"github.com/andygeiss/cloud-native-utils/resource"
)

// ReservationRepository provides CRUD operations for reservations.
type ReservationRepository resource.Access[ReservationID, Reservation]

// AvailabilityChecker validates room availability for reservations.
type AvailabilityChecker interface {
	// IsRoomAvailable checks if a room is available for the given date range
	IsRoomAvailable(ctx context.Context, roomID RoomID, dateRange DateRange) (bool, error)
	// GetOverlappingReservations returns all reservations that overlap with the given date range
	GetOverlappingReservations(ctx context.Context, roomID RoomID, dateRange DateRange) ([]*Reservation, error)
}

// EventPublisher publishes domain events.
type EventPublisher event.EventPublisher
