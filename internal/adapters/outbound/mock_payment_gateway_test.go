package outbound_test

import (
	"context"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/outbound"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/booking"
)

// ============================================================================
// MockPaymentGateway Tests
// ============================================================================

func Test_MockPaymentGateway_Authorize_Should_Succeed(t *testing.T) {
	// Arrange
	gateway := outbound.NewMockPaymentGateway()
	ctx := context.Background()
	payment := booking.NewPayment("pay-001", "res-001", booking.NewMoney(10000, "USD"), "credit_card")

	// Act
	txnID, err := gateway.Authorize(ctx, payment)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "transaction ID must not be empty", txnID != "", true)
}

func Test_MockPaymentGateway_Authorize_With_ShouldFail_Should_Return_Error(t *testing.T) {
	// Arrange
	gateway := outbound.NewMockPaymentGateway()
	gateway.SetShouldFail(true)
	ctx := context.Background()
	payment := booking.NewPayment("pay-001", "res-001", booking.NewMoney(10000, "USD"), "credit_card")

	// Act
	_, err := gateway.Authorize(ctx, payment)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

func Test_MockPaymentGateway_Capture_Should_Succeed(t *testing.T) {
	// Arrange
	gateway := outbound.NewMockPaymentGateway()
	ctx := context.Background()
	payment := booking.NewPayment("pay-001", "res-001", booking.NewMoney(10000, "USD"), "credit_card")

	txnID, authErr := gateway.Authorize(ctx, payment)
	if authErr != nil {
		t.Fatalf("setup failed: %v", authErr)
	}

	// Act
	err := gateway.Capture(ctx, txnID, payment.Amount)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
}

func Test_MockPaymentGateway_Capture_Unknown_Transaction_Should_Return_Error(t *testing.T) {
	// Arrange
	gateway := outbound.NewMockPaymentGateway()
	ctx := context.Background()

	// Act
	err := gateway.Capture(ctx, "unknown-txn", booking.NewMoney(10000, "USD"))

	// Assert
	assert.That(t, "error must not be nil for unknown transaction", err != nil, true)
}

func Test_MockPaymentGateway_Capture_Amount_Mismatch_Should_Return_Error(t *testing.T) {
	// Arrange
	gateway := outbound.NewMockPaymentGateway()
	ctx := context.Background()
	payment := booking.NewPayment("pay-001", "res-001", booking.NewMoney(10000, "USD"), "credit_card")

	txnID, authErr := gateway.Authorize(ctx, payment)
	if authErr != nil {
		t.Fatalf("setup failed: %v", authErr)
	}

	// Act
	err := gateway.Capture(ctx, txnID, booking.NewMoney(5000, "USD"))

	// Assert
	assert.That(t, "error must not be nil for amount mismatch", err != nil, true)
}

func Test_MockPaymentGateway_Capture_With_ShouldFail_Should_Return_Error(t *testing.T) {
	// Arrange
	gateway := outbound.NewMockPaymentGateway()
	ctx := context.Background()
	payment := booking.NewPayment("pay-001", "res-001", booking.NewMoney(10000, "USD"), "credit_card")

	txnID, authErr := gateway.Authorize(ctx, payment)
	if authErr != nil {
		t.Fatalf("setup failed: %v", authErr)
	}

	gateway.SetShouldFail(true)

	// Act
	err := gateway.Capture(ctx, txnID, payment.Amount)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

func Test_MockPaymentGateway_Refund_Should_Succeed(t *testing.T) {
	// Arrange
	gateway := outbound.NewMockPaymentGateway()
	ctx := context.Background()
	payment := booking.NewPayment("pay-001", "res-001", booking.NewMoney(10000, "USD"), "credit_card")

	txnID, authErr := gateway.Authorize(ctx, payment)
	if authErr != nil {
		t.Fatalf("setup failed: %v", authErr)
	}

	// Act
	err := gateway.Refund(ctx, txnID, payment.Amount)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
}

func Test_MockPaymentGateway_Refund_Unknown_Transaction_Should_Return_Error(t *testing.T) {
	// Arrange
	gateway := outbound.NewMockPaymentGateway()
	ctx := context.Background()

	// Act
	err := gateway.Refund(ctx, "unknown-txn", booking.NewMoney(10000, "USD"))

	// Assert
	assert.That(t, "error must not be nil for unknown transaction", err != nil, true)
}

func Test_MockPaymentGateway_Refund_With_ShouldFail_Should_Return_Error(t *testing.T) {
	// Arrange
	gateway := outbound.NewMockPaymentGateway()
	ctx := context.Background()
	payment := booking.NewPayment("pay-001", "res-001", booking.NewMoney(10000, "USD"), "credit_card")

	txnID, authErr := gateway.Authorize(ctx, payment)
	if authErr != nil {
		t.Fatalf("setup failed: %v", authErr)
	}

	gateway.SetShouldFail(true)

	// Act
	err := gateway.Refund(ctx, txnID, payment.Amount)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

func Test_MockPaymentGateway_Reset_Should_Clear_Failure_State(t *testing.T) {
	// Arrange
	gateway := outbound.NewMockPaymentGateway()
	ctx := context.Background()

	gateway.SetShouldFail(true)
	gateway.SetFailureRate(0.5)

	// Act
	gateway.Reset()

	// Assert
	assert.That(t, "ShouldFail must be false after reset", gateway.ShouldFail, false)
	assert.That(t, "FailureRate must be 0.0 after reset", gateway.FailureRate, 0.0)

	payment := booking.NewPayment("pay-001", "res-001", booking.NewMoney(10000, "USD"), "credit_card")
	_, err := gateway.Authorize(ctx, payment)
	assert.That(t, "authorize must work after reset", err == nil, true)
}

func Test_MockPaymentGateway_SetFailureRate_Should_Affect_Behavior(t *testing.T) {
	// Arrange
	gateway := outbound.NewMockPaymentGateway()

	// Act & Assert
	gateway.SetFailureRate(0.0)
	assert.That(t, "failure rate must be 0.0", gateway.FailureRate, 0.0)

	gateway.SetFailureRate(0.5)
	assert.That(t, "failure rate must be 0.5", gateway.FailureRate, 0.5)
}
