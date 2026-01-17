package reservation

import "time"

// Event topics for Kafka.
const (
	EventTopicCreated   = "reservation.created"
	EventTopicConfirmed = "reservation.confirmed"
	EventTopicActivated = "reservation.activated"
	EventTopicCompleted = "reservation.completed"
	EventTopicCancelled = "reservation.cancelled"
)

// EventCreated is published when a new reservation is created.
type EventCreated struct {
	ReservationID ReservationID `json:"reservation_id"`
	GuestID       GuestID       `json:"guest_id"`
	RoomID        RoomID        `json:"room_id"`
	CheckIn       time.Time     `json:"check_in"`
	CheckOut      time.Time     `json:"check_out"`
	TotalAmount   Money         `json:"total_amount"`
}

func NewEventCreated() *EventCreated {
	return &EventCreated{}
}

func (e *EventCreated) Topic() string { return EventTopicCreated }

func (e *EventCreated) WithReservationID(id ReservationID) *EventCreated {
	e.ReservationID = id
	return e
}

func (e *EventCreated) WithGuestID(id GuestID) *EventCreated {
	e.GuestID = id
	return e
}

func (e *EventCreated) WithRoomID(id RoomID) *EventCreated {
	e.RoomID = id
	return e
}

func (e *EventCreated) WithCheckIn(t time.Time) *EventCreated {
	e.CheckIn = t
	return e
}

func (e *EventCreated) WithCheckOut(t time.Time) *EventCreated {
	e.CheckOut = t
	return e
}

func (e *EventCreated) WithTotalAmount(m Money) *EventCreated {
	e.TotalAmount = m
	return e
}

// EventConfirmed is published when a reservation is confirmed.
type EventConfirmed struct {
	ReservationID ReservationID `json:"reservation_id"`
	GuestID       GuestID       `json:"guest_id"`
}

func NewEventConfirmed() *EventConfirmed {
	return &EventConfirmed{}
}

func (e *EventConfirmed) Topic() string { return EventTopicConfirmed }

func (e *EventConfirmed) WithReservationID(id ReservationID) *EventConfirmed {
	e.ReservationID = id
	return e
}

func (e *EventConfirmed) WithGuestID(id GuestID) *EventConfirmed {
	e.GuestID = id
	return e
}

// EventActivated is published when a guest checks in.
type EventActivated struct {
	ReservationID ReservationID `json:"reservation_id"`
}

func NewEventActivated() *EventActivated {
	return &EventActivated{}
}

func (e *EventActivated) Topic() string { return EventTopicActivated }

func (e *EventActivated) WithReservationID(id ReservationID) *EventActivated {
	e.ReservationID = id
	return e
}

// EventCompleted is published when a guest checks out.
type EventCompleted struct {
	ReservationID ReservationID `json:"reservation_id"`
}

func NewEventCompleted() *EventCompleted {
	return &EventCompleted{}
}

func (e *EventCompleted) Topic() string { return EventTopicCompleted }

func (e *EventCompleted) WithReservationID(id ReservationID) *EventCompleted {
	e.ReservationID = id
	return e
}

// EventCancelled is published when a reservation is cancelled.
type EventCancelled struct {
	ReservationID ReservationID `json:"reservation_id"`
	GuestID       GuestID       `json:"guest_id"`
	Reason        string        `json:"reason"`
}

func NewEventCancelled() *EventCancelled {
	return &EventCancelled{}
}

func (e *EventCancelled) Topic() string { return EventTopicCancelled }

func (e *EventCancelled) WithReservationID(id ReservationID) *EventCancelled {
	e.ReservationID = id
	return e
}

func (e *EventCancelled) WithGuestID(id GuestID) *EventCancelled {
	e.GuestID = id
	return e
}

func (e *EventCancelled) WithReason(reason string) *EventCancelled {
	e.Reason = reason
	return e
}
