package payment_test

import (
	"context"
	"errors"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/event"
	"github.com/andygeiss/hotel-booking/internal/domain/payment"
	"github.com/andygeiss/hotel-booking/internal/domain/shared"
)

// ============================================================================
// Mock Implementations
// ============================================================================

type mockPaymentRepository struct {
	payments  map[payment.PaymentID]payment.Payment
	createErr error
	readErr   error
	updateErr error
}

func newMockPaymentRepository() *mockPaymentRepository {
	return &mockPaymentRepository{
		payments: make(map[payment.PaymentID]payment.Payment),
	}
}

func (m *mockPaymentRepository) Create(ctx context.Context, id payment.PaymentID, p payment.Payment) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.payments[id] = p
	return nil
}

func (m *mockPaymentRepository) Read(ctx context.Context, id payment.PaymentID) (*payment.Payment, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	p, ok := m.payments[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return &p, nil
}

func (m *mockPaymentRepository) Update(ctx context.Context, id payment.PaymentID, p payment.Payment) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.payments[id] = p
	return nil
}

func (m *mockPaymentRepository) Delete(ctx context.Context, id payment.PaymentID) error {
	delete(m.payments, id)
	return nil
}

func (m *mockPaymentRepository) ReadAll(ctx context.Context) ([]payment.Payment, error) {
	result := make([]payment.Payment, 0, len(m.payments))
	for _, p := range m.payments {
		result = append(result, p)
	}
	return result, nil
}

type mockPaymentGateway struct {
	authorizeTransactionID string
	authorizeErr           error
	captureErr             error
	refundErr              error
}

func (m *mockPaymentGateway) Authorize(ctx context.Context, p *payment.Payment) (string, error) {
	if m.authorizeErr != nil {
		return "", m.authorizeErr
	}
	return m.authorizeTransactionID, nil
}

func (m *mockPaymentGateway) Capture(ctx context.Context, transactionID string, amount shared.Money) error {
	return m.captureErr
}

func (m *mockPaymentGateway) Refund(ctx context.Context, transactionID string, amount shared.Money) error {
	return m.refundErr
}

type mockEventPublisher struct {
	published []event.Event
	err       error
}

func (m *mockEventPublisher) Publish(ctx context.Context, evt event.Event) error {
	if m.err != nil {
		return m.err
	}
	m.published = append(m.published, evt)
	return nil
}

// ============================================================================
// Service Test Helpers
// ============================================================================

func createPaymentTestService(repo *mockPaymentRepository, gateway *mockPaymentGateway, publisher *mockEventPublisher) *payment.Service {
	return payment.NewService(repo, gateway, publisher)
}

func paymentTestMoney() shared.Money {
	return shared.NewMoney(10000, "USD")
}

// ============================================================================
// AuthorizePayment Tests
// ============================================================================

func Test_Service_AuthorizePayment_Should_Succeed(t *testing.T) {
	// Arrange
	repo := newMockPaymentRepository()
	gateway := &mockPaymentGateway{authorizeTransactionID: "tx-12345"}
	publisher := &mockEventPublisher{}
	service := createPaymentTestService(repo, gateway, publisher)

	ctx := context.Background()
	id := payment.PaymentID("pay-001")
	reservationID := payment.ReservationID("res-001")
	amount := paymentTestMoney()
	method := "credit_card"

	// Act
	p, err := service.AuthorizePayment(ctx, id, reservationID, amount, method)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "payment must not be nil", p != nil, true)
	assert.That(t, "status must be authorized", p.Status, payment.StatusAuthorized)
	assert.That(t, "TransactionID must be set", p.TransactionID, "tx-12345")
}

func Test_Service_AuthorizePayment_When_Gateway_Fails_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newMockPaymentRepository()
	gateway := &mockPaymentGateway{authorizeErr: errors.New("gateway unavailable")}
	publisher := &mockEventPublisher{}
	service := createPaymentTestService(repo, gateway, publisher)

	ctx := context.Background()
	id := payment.PaymentID("pay-001")
	reservationID := payment.ReservationID("res-001")
	amount := paymentTestMoney()
	method := "credit_card"

	// Act
	p, err := service.AuthorizePayment(ctx, id, reservationID, amount, method)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "payment must be nil", p == nil, true)
}

func Test_Service_AuthorizePayment_When_Gateway_Fails_Should_Persist_Failed_Payment(t *testing.T) {
	// Arrange
	repo := newMockPaymentRepository()
	gateway := &mockPaymentGateway{authorizeErr: errors.New("declined")}
	publisher := &mockEventPublisher{}
	service := createPaymentTestService(repo, gateway, publisher)

	ctx := context.Background()
	id := payment.PaymentID("pay-001")

	// Act
	_, _ = service.AuthorizePayment(ctx, id, "res-001", paymentTestMoney(), "credit_card")

	// Assert
	storedPayment, err := repo.Read(ctx, id)
	assert.That(t, "stored payment must exist", err == nil, true)
	assert.That(t, "stored payment status must be failed", storedPayment.Status, payment.StatusFailed)
}

func Test_Service_AuthorizePayment_Should_Publish_Event(t *testing.T) {
	// Arrange
	repo := newMockPaymentRepository()
	gateway := &mockPaymentGateway{authorizeTransactionID: "tx-12345"}
	publisher := &mockEventPublisher{}
	service := createPaymentTestService(repo, gateway, publisher)

	ctx := context.Background()
	id := payment.PaymentID("pay-001")

	// Act
	_, err := service.AuthorizePayment(ctx, id, "res-001", paymentTestMoney(), "credit_card")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "one event must be published", len(publisher.published), 1)
}

func Test_Service_AuthorizePayment_When_Gateway_Fails_Should_Publish_Failure_Event(t *testing.T) {
	// Arrange
	repo := newMockPaymentRepository()
	gateway := &mockPaymentGateway{authorizeErr: errors.New("declined")}
	publisher := &mockEventPublisher{}
	service := createPaymentTestService(repo, gateway, publisher)

	ctx := context.Background()

	// Act
	_, _ = service.AuthorizePayment(ctx, "pay-001", "res-001", paymentTestMoney(), "credit_card")

	// Assert
	assert.That(t, "one event must be published", len(publisher.published), 1)
}

func Test_Service_AuthorizePayment_When_Repository_Fails_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newMockPaymentRepository()
	repo.createErr = errors.New("database error")
	gateway := &mockPaymentGateway{authorizeTransactionID: "tx-12345"}
	publisher := &mockEventPublisher{}
	service := createPaymentTestService(repo, gateway, publisher)

	ctx := context.Background()

	// Act
	p, err := service.AuthorizePayment(ctx, "pay-001", "res-001", paymentTestMoney(), "credit_card")

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "payment must be nil", p == nil, true)
}

// ============================================================================
// CapturePayment Tests
// ============================================================================

func Test_Service_CapturePayment_Should_Succeed(t *testing.T) {
	// Arrange
	repo := newMockPaymentRepository()
	gateway := &mockPaymentGateway{authorizeTransactionID: "tx-12345"}
	publisher := &mockEventPublisher{}
	service := createPaymentTestService(repo, gateway, publisher)

	ctx := context.Background()
	id := payment.PaymentID("pay-001")

	// Create and authorize first
	_, _ = service.AuthorizePayment(ctx, id, "res-001", paymentTestMoney(), "credit_card")
	publisher.published = nil // reset

	// Act
	err := service.CapturePayment(ctx, id)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	storedPayment, _ := repo.Read(ctx, id)
	assert.That(t, "status must be captured", storedPayment.Status, payment.StatusCaptured)
}

func Test_Service_CapturePayment_When_Not_Found_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newMockPaymentRepository()
	gateway := &mockPaymentGateway{}
	publisher := &mockEventPublisher{}
	service := createPaymentTestService(repo, gateway, publisher)

	ctx := context.Background()

	// Act
	err := service.CapturePayment(ctx, "non-existent")

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

func Test_Service_CapturePayment_Should_Publish_Event(t *testing.T) {
	// Arrange
	repo := newMockPaymentRepository()
	gateway := &mockPaymentGateway{authorizeTransactionID: "tx-12345"}
	publisher := &mockEventPublisher{}
	service := createPaymentTestService(repo, gateway, publisher)

	ctx := context.Background()
	id := payment.PaymentID("pay-001")

	_, _ = service.AuthorizePayment(ctx, id, "res-001", paymentTestMoney(), "credit_card")
	publisher.published = nil // reset

	// Act
	err := service.CapturePayment(ctx, id)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "one event must be published", len(publisher.published), 1)
}

func Test_Service_CapturePayment_When_Gateway_Fails_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newMockPaymentRepository()
	gateway := &mockPaymentGateway{
		authorizeTransactionID: "tx-12345",
		captureErr:             errors.New("capture failed"),
	}
	publisher := &mockEventPublisher{}
	service := createPaymentTestService(repo, gateway, publisher)

	ctx := context.Background()
	id := payment.PaymentID("pay-001")

	_, _ = service.AuthorizePayment(ctx, id, "res-001", paymentTestMoney(), "credit_card")

	// Act
	err := service.CapturePayment(ctx, id)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

func Test_Service_CapturePayment_When_Gateway_Fails_Should_Mark_Payment_Failed(t *testing.T) {
	// Arrange
	repo := newMockPaymentRepository()
	gateway := &mockPaymentGateway{
		authorizeTransactionID: "tx-12345",
		captureErr:             errors.New("capture failed"),
	}
	publisher := &mockEventPublisher{}
	service := createPaymentTestService(repo, gateway, publisher)

	ctx := context.Background()
	id := payment.PaymentID("pay-001")

	_, _ = service.AuthorizePayment(ctx, id, "res-001", paymentTestMoney(), "credit_card")

	// Act
	_ = service.CapturePayment(ctx, id)

	// Assert
	storedPayment, _ := repo.Read(ctx, id)
	assert.That(t, "status must be failed", storedPayment.Status, payment.StatusFailed)
}

// ============================================================================
// RefundPayment Tests
// ============================================================================

func Test_Service_RefundPayment_Should_Succeed(t *testing.T) {
	// Arrange
	repo := newMockPaymentRepository()
	gateway := &mockPaymentGateway{authorizeTransactionID: "tx-12345"}
	publisher := &mockEventPublisher{}
	service := createPaymentTestService(repo, gateway, publisher)

	ctx := context.Background()
	id := payment.PaymentID("pay-001")

	// Create, authorize, and capture first
	_, _ = service.AuthorizePayment(ctx, id, "res-001", paymentTestMoney(), "credit_card")
	_ = service.CapturePayment(ctx, id)
	publisher.published = nil // reset

	// Act
	err := service.RefundPayment(ctx, id)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	storedPayment, _ := repo.Read(ctx, id)
	assert.That(t, "status must be refunded", storedPayment.Status, payment.StatusRefunded)
}

func Test_Service_RefundPayment_When_Not_Captured_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newMockPaymentRepository()
	gateway := &mockPaymentGateway{authorizeTransactionID: "tx-12345"}
	publisher := &mockEventPublisher{}
	service := createPaymentTestService(repo, gateway, publisher)

	ctx := context.Background()
	id := payment.PaymentID("pay-001")

	// Only authorize, don't capture
	_, _ = service.AuthorizePayment(ctx, id, "res-001", paymentTestMoney(), "credit_card")

	// Act
	err := service.RefundPayment(ctx, id)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

func Test_Service_RefundPayment_Should_Publish_Event(t *testing.T) {
	// Arrange
	repo := newMockPaymentRepository()
	gateway := &mockPaymentGateway{authorizeTransactionID: "tx-12345"}
	publisher := &mockEventPublisher{}
	service := createPaymentTestService(repo, gateway, publisher)

	ctx := context.Background()
	id := payment.PaymentID("pay-001")

	_, _ = service.AuthorizePayment(ctx, id, "res-001", paymentTestMoney(), "credit_card")
	_ = service.CapturePayment(ctx, id)
	publisher.published = nil // reset

	// Act
	err := service.RefundPayment(ctx, id)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "one event must be published", len(publisher.published), 1)
}

func Test_Service_RefundPayment_When_Gateway_Fails_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newMockPaymentRepository()
	gateway := &mockPaymentGateway{
		authorizeTransactionID: "tx-12345",
		refundErr:              errors.New("refund failed"),
	}
	publisher := &mockEventPublisher{}
	service := createPaymentTestService(repo, gateway, publisher)

	ctx := context.Background()
	id := payment.PaymentID("pay-001")

	_, _ = service.AuthorizePayment(ctx, id, "res-001", paymentTestMoney(), "credit_card")
	_ = service.CapturePayment(ctx, id)

	// Act
	err := service.RefundPayment(ctx, id)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

// ============================================================================
// GetPayment Tests
// ============================================================================

func Test_Service_GetPayment_Should_Return_Payment(t *testing.T) {
	// Arrange
	repo := newMockPaymentRepository()
	gateway := &mockPaymentGateway{authorizeTransactionID: "tx-12345"}
	publisher := &mockEventPublisher{}
	service := createPaymentTestService(repo, gateway, publisher)

	ctx := context.Background()
	id := payment.PaymentID("pay-001")

	_, _ = service.AuthorizePayment(ctx, id, "res-001", paymentTestMoney(), "credit_card")

	// Act
	p, err := service.GetPayment(ctx, id)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "payment must not be nil", p != nil, true)
	assert.That(t, "payment ID must match", p.ID, id)
}

func Test_Service_GetPayment_When_Not_Found_Should_Return_Error(t *testing.T) {
	// Arrange
	repo := newMockPaymentRepository()
	gateway := &mockPaymentGateway{}
	publisher := &mockEventPublisher{}
	service := createPaymentTestService(repo, gateway, publisher)

	ctx := context.Background()

	// Act
	p, err := service.GetPayment(ctx, "non-existent")

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "payment must be nil", p == nil, true)
}

// ============================================================================
// Event Handler Integration Tests
// ============================================================================

func Test_Service_AuthorizePaymentForReservation_Should_Authorize(t *testing.T) {
	// Arrange
	repo := newMockPaymentRepository()
	gateway := &mockPaymentGateway{authorizeTransactionID: "tx-12345"}
	publisher := &mockEventPublisher{}
	service := createPaymentTestService(repo, gateway, publisher)

	ctx := context.Background()
	id := payment.PaymentID("pay-001")
	reservationID := payment.ReservationID("res-001")

	// Act
	p, err := service.AuthorizePaymentForReservation(ctx, id, reservationID, paymentTestMoney(), "credit_card")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "payment must not be nil", p != nil, true)
	assert.That(t, "status must be authorized", p.Status, payment.StatusAuthorized)
}

func Test_Service_CapturePaymentOnAuthorization_Should_Capture(t *testing.T) {
	// Arrange
	repo := newMockPaymentRepository()
	gateway := &mockPaymentGateway{authorizeTransactionID: "tx-12345"}
	publisher := &mockEventPublisher{}
	service := createPaymentTestService(repo, gateway, publisher)

	ctx := context.Background()
	id := payment.PaymentID("pay-001")

	_, _ = service.AuthorizePayment(ctx, id, "res-001", paymentTestMoney(), "credit_card")

	// Act
	err := service.CapturePaymentOnAuthorization(ctx, id)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	storedPayment, _ := repo.Read(ctx, id)
	assert.That(t, "status must be captured", storedPayment.Status, payment.StatusCaptured)
}
