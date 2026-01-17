package payment

import (
	"context"

	"github.com/andygeiss/cloud-native-utils/event"
	"github.com/andygeiss/cloud-native-utils/resource"
)

// PaymentRepository provides CRUD operations for payments.
type PaymentRepository resource.Access[PaymentID, Payment]

// PaymentGateway handles payment processing with external providers.
type PaymentGateway interface {
	// Authorize holds funds without capturing them
	Authorize(ctx context.Context, payment *Payment) (transactionID string, err error)
	// Capture finalizes an authorized payment
	Capture(ctx context.Context, transactionID string, amount Money) error
	// Refund returns funds to the customer
	Refund(ctx context.Context, transactionID string, amount Money) error
}

// EventPublisher publishes domain events.
type EventPublisher event.EventPublisher
