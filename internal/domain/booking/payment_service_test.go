package booking_test

import (
	"context"
	"errors"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/booking"
)

// ============================================================================
// PaymentService Tests
// ============================================================================

func Test_PaymentService_AuthorizePayment_Should_Succeed(t *testing.T) {
	// Arrange
	payRepo := newMockPaymentRepository()
	payGateway := &mockPaymentGateway{txnID: "txn-test-123"}
	publisher := &mockEventPublisher{}
	svc := booking.NewPaymentService(payRepo, payGateway, publisher)
	ctx := context.Background()

	// Act
	payment, err := svc.AuthorizePayment(
		ctx,
		"pay-001",
		"res-001",
		booking.NewMoney(10000, "USD"),
		"credit_card",
	)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "payment must not be nil", payment != nil, true)
	assert.That(t, "status must be authorized", payment.Status, booking.PaymentAuthorized)
	assert.That(t, "transaction ID must match", payment.TransactionID, "txn-test-123")

	persisted, readErr := payRepo.Read(ctx, "pay-001")
	assert.That(t, "payment must be persisted", readErr == nil, true)
	assert.That(t, "persisted status must be authorized", persisted.Status, booking.PaymentAuthorized)
	assert.That(t, "one event must be published", len(publisher.events), 1)
}

func Test_PaymentService_AuthorizePayment_Gateway_Failure_Should_Fail_And_Persist(t *testing.T) {
	// Arrange
	payRepo := newMockPaymentRepository()
	payGateway := &mockPaymentGateway{authorizeErr: errors.New("gateway error")}
	publisher := &mockEventPublisher{}
	svc := booking.NewPaymentService(payRepo, payGateway, publisher)
	ctx := context.Background()

	// Act
	_, err := svc.AuthorizePayment(
		ctx,
		"pay-001",
		"res-001",
		booking.NewMoney(10000, "USD"),
		"credit_card",
	)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)

	persisted, readErr := payRepo.Read(ctx, "pay-001")
	assert.That(t, "failed payment must be persisted", readErr == nil, true)
	assert.That(t, "status must be failed", persisted.Status, booking.PaymentFailed)
}

func Test_PaymentService_CapturePayment_Should_Succeed(t *testing.T) {
	// Arrange
	payRepo := newMockPaymentRepository()
	payGateway := &mockPaymentGateway{txnID: "txn-test-123"}
	publisher := &mockEventPublisher{}
	svc := booking.NewPaymentService(payRepo, payGateway, publisher)
	ctx := context.Background()

	_, setupErr := svc.AuthorizePayment(
		ctx,
		"pay-001",
		"res-001",
		booking.NewMoney(10000, "USD"),
		"credit_card",
	)
	if setupErr != nil {
		t.Fatalf("setup failed: %v", setupErr)
	}

	// Act
	err := svc.CapturePayment(ctx, "pay-001")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)

	persisted, readErr := payRepo.Read(ctx, "pay-001")
	assert.That(t, "payment must exist", readErr == nil, true)
	assert.That(t, "status must be captured", persisted.Status, booking.PaymentCaptured)
}

func Test_PaymentService_CapturePayment_Not_Found_Should_Fail(t *testing.T) {
	// Arrange
	payRepo := newMockPaymentRepository()
	payGateway := &mockPaymentGateway{}
	publisher := &mockEventPublisher{}
	svc := booking.NewPaymentService(payRepo, payGateway, publisher)
	ctx := context.Background()

	// Act
	err := svc.CapturePayment(ctx, "nonexistent")

	// Assert
	assert.That(t, "error must not be nil for nonexistent payment", err != nil, true)
}

func Test_PaymentService_CapturePayment_Gateway_Failure_Should_Fail(t *testing.T) {
	// Arrange
	payRepo := newMockPaymentRepository()
	payGateway := &mockPaymentGateway{txnID: "txn-test-123"}
	publisher := &mockEventPublisher{}
	svc := booking.NewPaymentService(payRepo, payGateway, publisher)
	ctx := context.Background()

	_, setupErr := svc.AuthorizePayment(
		ctx,
		"pay-001",
		"res-001",
		booking.NewMoney(10000, "USD"),
		"credit_card",
	)
	if setupErr != nil {
		t.Fatalf("setup failed: %v", setupErr)
	}

	payGateway.captureErr = errors.New("capture failed")

	// Act
	err := svc.CapturePayment(ctx, "pay-001")

	// Assert
	assert.That(t, "error must not be nil for capture failure", err != nil, true)

	persisted, readErr := payRepo.Read(ctx, "pay-001")
	assert.That(t, "payment must exist", readErr == nil, true)
	assert.That(t, "status must be failed", persisted.Status, booking.PaymentFailed)
}

func Test_PaymentService_RefundPayment_Should_Succeed(t *testing.T) {
	// Arrange
	payRepo := newMockPaymentRepository()
	payGateway := &mockPaymentGateway{txnID: "txn-test-123"}
	publisher := &mockEventPublisher{}
	svc := booking.NewPaymentService(payRepo, payGateway, publisher)
	ctx := context.Background()

	_, setupErr := svc.AuthorizePayment(
		ctx,
		"pay-001",
		"res-001",
		booking.NewMoney(10000, "USD"),
		"credit_card",
	)
	if setupErr != nil {
		t.Fatalf("setup failed: %v", setupErr)
	}

	captureErr := svc.CapturePayment(ctx, "pay-001")
	if captureErr != nil {
		t.Fatalf("setup failed: %v", captureErr)
	}

	// Act
	err := svc.RefundPayment(ctx, "pay-001")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)

	persisted, readErr := payRepo.Read(ctx, "pay-001")
	assert.That(t, "payment must exist", readErr == nil, true)
	assert.That(t, "status must be refunded", persisted.Status, booking.PaymentRefunded)
}

func Test_PaymentService_RefundPayment_Gateway_Failure_Should_Fail(t *testing.T) {
	// Arrange
	payRepo := newMockPaymentRepository()
	payGateway := &mockPaymentGateway{txnID: "txn-test-123"}
	publisher := &mockEventPublisher{}
	svc := booking.NewPaymentService(payRepo, payGateway, publisher)
	ctx := context.Background()

	_, setupErr := svc.AuthorizePayment(
		ctx,
		"pay-001",
		"res-001",
		booking.NewMoney(10000, "USD"),
		"credit_card",
	)
	if setupErr != nil {
		t.Fatalf("setup failed: %v", setupErr)
	}

	captureErr := svc.CapturePayment(ctx, "pay-001")
	if captureErr != nil {
		t.Fatalf("setup failed: %v", captureErr)
	}

	payGateway.refundErr = errors.New("refund failed")

	// Act
	err := svc.RefundPayment(ctx, "pay-001")

	// Assert
	assert.That(t, "error must not be nil for refund failure", err != nil, true)
}

func Test_PaymentService_GetPayment_Should_Return_Payment(t *testing.T) {
	// Arrange
	payRepo := newMockPaymentRepository()
	payGateway := &mockPaymentGateway{txnID: "txn-test-123"}
	publisher := &mockEventPublisher{}
	svc := booking.NewPaymentService(payRepo, payGateway, publisher)
	ctx := context.Background()

	_, setupErr := svc.AuthorizePayment(
		ctx,
		"pay-001",
		"res-001",
		booking.NewMoney(10000, "USD"),
		"credit_card",
	)
	if setupErr != nil {
		t.Fatalf("setup failed: %v", setupErr)
	}

	// Act
	payment, err := svc.GetPayment(ctx, "pay-001")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "payment must not be nil", payment != nil, true)
	assert.That(t, "ID must match", string(payment.ID), "pay-001")
}

func Test_PaymentService_GetPayment_Not_Found_Should_Fail(t *testing.T) {
	// Arrange
	payRepo := newMockPaymentRepository()
	payGateway := &mockPaymentGateway{}
	publisher := &mockEventPublisher{}
	svc := booking.NewPaymentService(payRepo, payGateway, publisher)
	ctx := context.Background()

	// Act
	_, err := svc.GetPayment(ctx, "nonexistent")

	// Assert
	assert.That(t, "error must not be nil for nonexistent payment", err != nil, true)
}
