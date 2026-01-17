// Package reservation contains the Reservation bounded context.
// It handles all reservation-related domain logic including creation,
// state transitions, and business rule validation.
package reservation

import (
	"errors"
	"fmt"
	"time"

	"github.com/andygeiss/hotel-booking/internal/domain/shared"
)

// Type aliases for shared types
type ReservationID = shared.ReservationID
type Money = shared.Money

// Local ID types for this bounded context
type GuestID string
type RoomID string

// ReservationStatus represents the state of a reservation.
type ReservationStatus string

const (
	StatusPending   ReservationStatus = "pending"
	StatusConfirmed ReservationStatus = "confirmed"
	StatusActive    ReservationStatus = "active"
	StatusCompleted ReservationStatus = "completed"
	StatusCancelled ReservationStatus = "cancelled"
)

// Reservation is the aggregate root for booking reservations.
type Reservation struct {
	ID                 ReservationID
	GuestID            GuestID
	RoomID             RoomID
	DateRange          DateRange
	Status             ReservationStatus
	TotalAmount        Money
	CancellationReason string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Guests             []GuestInfo
}

// Validation errors.
var (
	ErrInvalidDateRange        = errors.New("check-out must be after check-in")
	ErrCheckInPast             = errors.New("check-in date must be in the future")
	ErrMinimumStay             = errors.New("minimum stay is 1 night")
	ErrInvalidStateTransition  = errors.New("invalid state transition")
	ErrCannotCancelNearCheckIn = errors.New("cannot cancel within 24 hours of check-in")
	ErrCannotCancelActive      = errors.New("cannot cancel active reservation")
	ErrCannotCancelCompleted   = errors.New("cannot cancel completed reservation")
	ErrAlreadyCancelled        = errors.New("reservation already cancelled")
	ErrNoGuests                = errors.New("at least one guest required")
)

// NewReservation creates a new reservation with validation.
func NewReservation(id ReservationID, guestID GuestID, roomID RoomID, dateRange DateRange, amount Money, guests []GuestInfo) (*Reservation, error) {
	r := &Reservation{
		ID:          id,
		GuestID:     guestID,
		RoomID:      roomID,
		DateRange:   dateRange,
		Status:      StatusPending,
		TotalAmount: amount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Guests:      guests,
	}

	if err := r.validate(); err != nil {
		return nil, err
	}

	return r, nil
}

// Confirm transitions the reservation from pending to confirmed.
func (r *Reservation) Confirm() error {
	if r.Status != StatusPending {
		return fmt.Errorf("%w: cannot confirm from %s", ErrInvalidStateTransition, r.Status)
	}

	r.Status = StatusConfirmed
	r.UpdatedAt = time.Now()
	return nil
}

// Activate transitions the reservation to active (check-in).
func (r *Reservation) Activate() error {
	if r.Status != StatusConfirmed {
		return fmt.Errorf("%w: cannot activate from %s", ErrInvalidStateTransition, r.Status)
	}

	r.Status = StatusActive
	r.UpdatedAt = time.Now()
	return nil
}

// Complete transitions the reservation to completed (check-out).
func (r *Reservation) Complete() error {
	if r.Status != StatusActive {
		return fmt.Errorf("%w: cannot complete from %s", ErrInvalidStateTransition, r.Status)
	}

	r.Status = StatusCompleted
	r.UpdatedAt = time.Now()
	return nil
}

// Cancel cancels the reservation with business rule validation.
func (r *Reservation) Cancel(reason string) error {
	if r.Status == StatusCancelled {
		return ErrAlreadyCancelled
	}

	if r.Status == StatusCompleted {
		return ErrCannotCancelCompleted
	}

	if r.Status == StatusActive {
		return ErrCannotCancelActive
	}

	if !r.CanBeCancelled() {
		return ErrCannotCancelNearCheckIn
	}

	r.Status = StatusCancelled
	r.CancellationReason = reason
	r.UpdatedAt = time.Now()
	return nil
}

// CanBeCancelled checks if the reservation can be cancelled based on business rules.
func (r *Reservation) CanBeCancelled() bool {
	if r.Status == StatusCancelled || r.Status == StatusCompleted || r.Status == StatusActive {
		return false
	}

	now := time.Now()
	hoursUntilCheckIn := r.DateRange.CheckIn.Sub(now).Hours()
	return hoursUntilCheckIn >= 24
}

// IsOverlapping checks if this reservation overlaps with another for the same room.
func (r *Reservation) IsOverlapping(other *Reservation) bool {
	if r.RoomID != other.RoomID {
		return false
	}

	if r.Status == StatusCancelled || other.Status == StatusCancelled {
		return false
	}

	return r.DateRange.CheckIn.Before(other.DateRange.CheckOut) &&
		r.DateRange.CheckOut.After(other.DateRange.CheckIn)
}

// DaysUntilCheckIn returns the number of days until check-in.
func (r *Reservation) DaysUntilCheckIn() int {
	now := time.Now().Truncate(24 * time.Hour)
	checkIn := r.DateRange.CheckIn.Truncate(24 * time.Hour)
	days := checkIn.Sub(now).Hours() / 24
	return int(days)
}

// Nights returns the number of nights for this reservation.
func (r *Reservation) Nights() int {
	nights := r.DateRange.CheckOut.Sub(r.DateRange.CheckIn).Hours() / 24
	return int(nights)
}

func (r *Reservation) validate() error {
	if err := r.validateDateRange(); err != nil {
		return err
	}

	if len(r.Guests) == 0 {
		return ErrNoGuests
	}

	return nil
}

func (r *Reservation) validateDateRange() error {
	nights := r.DateRange.CheckOut.Sub(r.DateRange.CheckIn).Hours() / 24

	if nights < 1 {
		if r.DateRange.CheckOut.Equal(r.DateRange.CheckIn) {
			return ErrMinimumStay
		}
		return ErrInvalidDateRange
	}

	if !r.DateRange.CheckOut.After(r.DateRange.CheckIn) {
		return ErrInvalidDateRange
	}

	now := time.Now().Truncate(24 * time.Hour)
	checkIn := r.DateRange.CheckIn.Truncate(24 * time.Hour)
	if checkIn.Before(now) {
		return ErrCheckInPast
	}

	return nil
}
