package payment

// Event topics for Kafka.
const (
	EventTopicAuthorized = "payment.authorized"
	EventTopicCaptured   = "payment.captured"
	EventTopicFailed     = "payment.failed"
	EventTopicRefunded   = "payment.refunded"
)

// EventAuthorized is published when a payment is authorized.
type EventAuthorized struct {
	PaymentID     PaymentID     `json:"payment_id"`
	ReservationID ReservationID `json:"reservation_id"`
	TransactionID string        `json:"transaction_id"`
	Amount        Money         `json:"amount"`
}

func NewEventAuthorized() *EventAuthorized {
	return &EventAuthorized{}
}

func (e *EventAuthorized) Topic() string { return EventTopicAuthorized }

func (e *EventAuthorized) WithPaymentID(id PaymentID) *EventAuthorized {
	e.PaymentID = id
	return e
}

func (e *EventAuthorized) WithReservationID(id ReservationID) *EventAuthorized {
	e.ReservationID = id
	return e
}

func (e *EventAuthorized) WithTransactionID(id string) *EventAuthorized {
	e.TransactionID = id
	return e
}

func (e *EventAuthorized) WithAmount(m Money) *EventAuthorized {
	e.Amount = m
	return e
}

// EventCaptured is published when a payment is captured.
type EventCaptured struct {
	PaymentID     PaymentID     `json:"payment_id"`
	ReservationID ReservationID `json:"reservation_id"`
	Amount        Money         `json:"amount"`
}

func NewEventCaptured() *EventCaptured {
	return &EventCaptured{}
}

func (e *EventCaptured) Topic() string { return EventTopicCaptured }

func (e *EventCaptured) WithPaymentID(id PaymentID) *EventCaptured {
	e.PaymentID = id
	return e
}

func (e *EventCaptured) WithReservationID(id ReservationID) *EventCaptured {
	e.ReservationID = id
	return e
}

func (e *EventCaptured) WithAmount(m Money) *EventCaptured {
	e.Amount = m
	return e
}

// EventFailed is published when a payment fails.
type EventFailed struct {
	PaymentID     PaymentID     `json:"payment_id"`
	ReservationID ReservationID `json:"reservation_id"`
	ErrorCode     string        `json:"error_code"`
	ErrorMsg      string        `json:"error_msg"`
}

func NewEventFailed() *EventFailed {
	return &EventFailed{}
}

func (e *EventFailed) Topic() string { return EventTopicFailed }

func (e *EventFailed) WithPaymentID(id PaymentID) *EventFailed {
	e.PaymentID = id
	return e
}

func (e *EventFailed) WithReservationID(id ReservationID) *EventFailed {
	e.ReservationID = id
	return e
}

func (e *EventFailed) WithErrorCode(code string) *EventFailed {
	e.ErrorCode = code
	return e
}

func (e *EventFailed) WithErrorMsg(msg string) *EventFailed {
	e.ErrorMsg = msg
	return e
}

// EventRefunded is published when a payment is refunded.
type EventRefunded struct {
	PaymentID     PaymentID     `json:"payment_id"`
	ReservationID ReservationID `json:"reservation_id"`
	Amount        Money         `json:"amount"`
}

func NewEventRefunded() *EventRefunded {
	return &EventRefunded{}
}

func (e *EventRefunded) Topic() string { return EventTopicRefunded }

func (e *EventRefunded) WithPaymentID(id PaymentID) *EventRefunded {
	e.PaymentID = id
	return e
}

func (e *EventRefunded) WithReservationID(id ReservationID) *EventRefunded {
	e.ReservationID = id
	return e
}

func (e *EventRefunded) WithAmount(m Money) *EventRefunded {
	e.Amount = m
	return e
}
