package reservation

import (
	"context"
	"fmt"
	"time"

	"github.com/andygeiss/cloud-native-utils/event"
)

// Service handles reservation workflows.
type Service struct {
	reservationRepo     ReservationRepository
	availabilityChecker AvailabilityChecker
	publisher           event.EventPublisher
}

// NewService creates a new reservation Service with dependencies.
func NewService(
	repo ReservationRepository,
	checker AvailabilityChecker,
	pub event.EventPublisher,
) *Service {
	return &Service{
		reservationRepo:     repo,
		availabilityChecker: checker,
		publisher:           pub,
	}
}

// CreateReservation creates a new pending reservation after checking availability.
func (s *Service) CreateReservation(
	ctx context.Context,
	id ReservationID,
	guestID GuestID,
	roomID RoomID,
	dateRange DateRange,
	amount Money,
	guests []GuestInfo,
) (*Reservation, error) {
	// 1. Check room availability
	available, err := s.availabilityChecker.IsRoomAvailable(ctx, roomID, dateRange)
	if err != nil {
		return nil, fmt.Errorf("failed to check availability: %w", err)
	}
	if !available {
		return nil, fmt.Errorf("room %s is not available for the selected dates", roomID)
	}

	// 2. Create reservation aggregate
	reservation, err := NewReservation(id, guestID, roomID, dateRange, amount, guests)
	if err != nil {
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}

	// 3. Persist to repository
	if err := s.reservationRepo.Create(ctx, id, *reservation); err != nil {
		return nil, fmt.Errorf("failed to persist reservation: %w", err)
	}

	// 4. Publish domain event
	evt := NewEventCreated().
		WithReservationID(id).
		WithGuestID(guestID).
		WithRoomID(roomID).
		WithCheckIn(dateRange.CheckIn).
		WithCheckOut(dateRange.CheckOut).
		WithTotalAmount(amount)

	if err := s.publisher.Publish(ctx, evt); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	return reservation, nil
}

// ConfirmReservation transitions a reservation to confirmed status.
func (s *Service) ConfirmReservation(ctx context.Context, id ReservationID) error {
	// 1. Load reservation from repository
	reservation, err := s.reservationRepo.Read(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to read reservation: %w", err)
	}

	// 2. Confirm reservation (aggregate business logic)
	if err := reservation.Confirm(); err != nil {
		return fmt.Errorf("failed to confirm reservation: %w", err)
	}

	// 3. Update repository
	if err := s.reservationRepo.Update(ctx, id, *reservation); err != nil {
		return fmt.Errorf("failed to update reservation: %w", err)
	}

	// 4. Publish domain event
	evt := NewEventConfirmed().
		WithReservationID(id).
		WithGuestID(reservation.GuestID)

	if err := s.publisher.Publish(ctx, evt); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// CancelReservation cancels a reservation with business rule validation.
func (s *Service) CancelReservation(ctx context.Context, id ReservationID, reason string) error {
	// 1. Load reservation from repository
	reservation, err := s.reservationRepo.Read(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to read reservation: %w", err)
	}

	guestID := reservation.GuestID

	// 2. Cancel reservation (aggregate business logic validates rules)
	if err := reservation.Cancel(reason); err != nil {
		return fmt.Errorf("failed to cancel reservation: %w", err)
	}

	// 3. Update repository
	if err := s.reservationRepo.Update(ctx, id, *reservation); err != nil {
		return fmt.Errorf("failed to update reservation: %w", err)
	}

	// 4. Publish domain event
	evt := NewEventCancelled().
		WithReservationID(id).
		WithGuestID(guestID).
		WithReason(reason)

	if err := s.publisher.Publish(ctx, evt); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// ActivateReservation transitions a reservation to active status (check-in).
func (s *Service) ActivateReservation(ctx context.Context, id ReservationID) error {
	reservation, err := s.reservationRepo.Read(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to read reservation: %w", err)
	}

	if err := reservation.Activate(); err != nil {
		return fmt.Errorf("failed to activate reservation: %w", err)
	}

	if err := s.reservationRepo.Update(ctx, id, *reservation); err != nil {
		return fmt.Errorf("failed to update reservation: %w", err)
	}

	evt := NewEventActivated().WithReservationID(id)
	if err := s.publisher.Publish(ctx, evt); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// CompleteReservation transitions a reservation to completed status (check-out).
func (s *Service) CompleteReservation(ctx context.Context, id ReservationID) error {
	reservation, err := s.reservationRepo.Read(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to read reservation: %w", err)
	}

	if err := reservation.Complete(); err != nil {
		return fmt.Errorf("failed to complete reservation: %w", err)
	}

	if err := s.reservationRepo.Update(ctx, id, *reservation); err != nil {
		return fmt.Errorf("failed to update reservation: %w", err)
	}

	evt := NewEventCompleted().WithReservationID(id)
	if err := s.publisher.Publish(ctx, evt); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// GetReservation retrieves a reservation by ID.
func (s *Service) GetReservation(ctx context.Context, id ReservationID) (*Reservation, error) {
	reservation, err := s.reservationRepo.Read(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to read reservation: %w", err)
	}
	return reservation, nil
}

// ListReservationsByGuest retrieves all reservations for a guest.
func (s *Service) ListReservationsByGuest(ctx context.Context, guestID GuestID) ([]*Reservation, error) {
	allReservations, err := s.reservationRepo.ReadAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list reservations: %w", err)
	}

	var guestReservations []*Reservation
	for i := range allReservations {
		if allReservations[i].GuestID == guestID {
			guestReservations = append(guestReservations, &allReservations[i])
		}
	}

	return guestReservations, nil
}

// ConfirmReservationOnPaymentCaptured handles the payment.captured event.
// This is called by the event handler when a payment is successfully captured.
func (s *Service) ConfirmReservationOnPaymentCaptured(ctx context.Context, reservationID ReservationID) error {
	return s.ConfirmReservation(ctx, reservationID)
}

// CancelReservationOnPaymentFailed handles the payment.failed event.
// This is called by the event handler when a payment fails.
func (s *Service) CancelReservationOnPaymentFailed(ctx context.Context, reservationID ReservationID, reason string) error {
	return s.CancelReservation(ctx, reservationID, reason)
}

// Event subscription helper for creating reservation events from messages.
func NewEventCreatedFromValues(reservationID ReservationID, guestID GuestID, roomID RoomID, checkIn, checkOut time.Time, amount Money) *EventCreated {
	return NewEventCreated().
		WithReservationID(reservationID).
		WithGuestID(guestID).
		WithRoomID(roomID).
		WithCheckIn(checkIn).
		WithCheckOut(checkOut).
		WithTotalAmount(amount)
}
