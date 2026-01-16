package booking_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/event"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/booking"
)

// ============================================================================
// Mock Implementations
// ============================================================================

type mockReservationRepository struct {
	reservations map[booking.ReservationID]booking.Reservation
	createErr    error
	readErr      error
	updateErr    error
}

func newMockReservationRepository() *mockReservationRepository {
	return &mockReservationRepository{
		reservations: make(map[booking.ReservationID]booking.Reservation),
	}
}

func (m *mockReservationRepository) Create(ctx context.Context, id booking.ReservationID, res booking.Reservation) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.reservations[id] = res
	return nil
}

func (m *mockReservationRepository) Read(ctx context.Context, id booking.ReservationID) (*booking.Reservation, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	res, ok := m.reservations[id]
	if !ok {
		return nil, errors.New("reservation not found")
	}
	return &res, nil
}

func (m *mockReservationRepository) Update(ctx context.Context, id booking.ReservationID, res booking.Reservation) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.reservations[id] = res
	return nil
}

func (m *mockReservationRepository) Delete(ctx context.Context, id booking.ReservationID) error {
	delete(m.reservations, id)
	return nil
}

func (m *mockReservationRepository) ReadAll(ctx context.Context) ([]booking.Reservation, error) {
	result := make([]booking.Reservation, 0, len(m.reservations))
	for _, res := range m.reservations {
		result = append(result, res)
	}
	return result, nil
}

type mockPaymentRepository struct {
	payments  map[booking.PaymentID]booking.Payment
	createErr error
	readErr   error
	updateErr error
}

func newMockPaymentRepository() *mockPaymentRepository {
	return &mockPaymentRepository{
		payments: make(map[booking.PaymentID]booking.Payment),
	}
}

func (m *mockPaymentRepository) Create(ctx context.Context, id booking.PaymentID, payment booking.Payment) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.payments[id] = payment
	return nil
}

func (m *mockPaymentRepository) Read(ctx context.Context, id booking.PaymentID) (*booking.Payment, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	payment, ok := m.payments[id]
	if !ok {
		return nil, errors.New("payment not found")
	}
	return &payment, nil
}

func (m *mockPaymentRepository) Update(ctx context.Context, id booking.PaymentID, payment booking.Payment) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.payments[id] = payment
	return nil
}

func (m *mockPaymentRepository) Delete(ctx context.Context, id booking.PaymentID) error {
	delete(m.payments, id)
	return nil
}

func (m *mockPaymentRepository) ReadAll(ctx context.Context) ([]booking.Payment, error) {
	result := make([]booking.Payment, 0, len(m.payments))
	for _, payment := range m.payments {
		result = append(result, payment)
	}
	return result, nil
}

type mockAvailabilityChecker struct {
	checkErr  error
	available bool
}

func (m *mockAvailabilityChecker) IsRoomAvailable(ctx context.Context, roomID booking.RoomID, dateRange booking.DateRange) (bool, error) {
	if m.checkErr != nil {
		return false, m.checkErr
	}
	return m.available, nil
}

func (m *mockAvailabilityChecker) GetOverlappingReservations(ctx context.Context, roomID booking.RoomID, dateRange booking.DateRange) ([]*booking.Reservation, error) {
	if m.checkErr != nil {
		return nil, m.checkErr
	}
	if m.available {
		return nil, nil
	}
	return []*booking.Reservation{{}}, nil
}

type mockPaymentGateway struct {
	authorizeErr error
	captureErr   error
	refundErr    error
	txnID        string
}

func (m *mockPaymentGateway) Authorize(ctx context.Context, payment *booking.Payment) (string, error) {
	if m.authorizeErr != nil {
		return "", m.authorizeErr
	}
	if m.txnID == "" {
		return "txn-123", nil
	}
	return m.txnID, nil
}

func (m *mockPaymentGateway) Capture(ctx context.Context, transactionID string, amount booking.Money) error {
	return m.captureErr
}

func (m *mockPaymentGateway) Refund(ctx context.Context, transactionID string, amount booking.Money) error {
	return m.refundErr
}

type mockEventPublisher struct {
	publishErr error
	events     []event.Event
}

func (m *mockEventPublisher) Publish(ctx context.Context, e event.Event) error {
	if m.publishErr != nil {
		return m.publishErr
	}
	m.events = append(m.events, e)
	return nil
}

type mockNotificationService struct {
	confirmationErr   error
	cancellationErr   error
	paymentReceiptErr error
}

func (m *mockNotificationService) SendReservationConfirmation(ctx context.Context, reservation *booking.Reservation) error {
	return m.confirmationErr
}

func (m *mockNotificationService) SendCancellationNotice(ctx context.Context, reservation *booking.Reservation, reason string) error {
	return m.cancellationErr
}

func (m *mockNotificationService) SendPaymentReceipt(ctx context.Context, payment *booking.Payment) error {
	return m.paymentReceiptErr
}

// ============================================================================
// Test Helpers
// ============================================================================

type testServices struct {
	orchSvc   *booking.BookingOrchestrationService
	resRepo   *mockReservationRepository
	payRepo   *mockPaymentRepository
	publisher *mockEventPublisher
}

func createTestOrchestrationService(t *testing.T) testServices {
	t.Helper()

	resRepo := newMockReservationRepository()
	payRepo := newMockPaymentRepository()
	availChecker := &mockAvailabilityChecker{available: true}
	payGateway := &mockPaymentGateway{}
	publisher := &mockEventPublisher{}
	notifSvc := &mockNotificationService{}

	resSvc := booking.NewReservationService(resRepo, availChecker, publisher)
	paySvc := booking.NewPaymentService(payRepo, payGateway, publisher)
	orchSvc := booking.NewBookingOrchestrationService(resSvc, paySvc, notifSvc)

	return testServices{
		orchSvc:   orchSvc,
		resRepo:   resRepo,
		payRepo:   payRepo,
		publisher: publisher,
	}
}

func createTestDateRange() booking.DateRange {
	checkIn := time.Now().AddDate(0, 0, 7)
	checkOut := time.Now().AddDate(0, 0, 10)
	return booking.NewDateRange(checkIn, checkOut)
}

func createTestGuests() []booking.GuestInfo {
	return []booking.GuestInfo{
		booking.NewGuestInfo("John Doe", "john@example.com", "+1234567890"),
	}
}

// ============================================================================
// BookingOrchestrationService Tests
// ============================================================================

func Test_BookingOrchestrationService_CompleteBooking_Should_Succeed(t *testing.T) {
	// Arrange
	svc := createTestOrchestrationService(t)
	ctx := context.Background()
	dateRange := createTestDateRange()
	guests := createTestGuests()
	amount := booking.NewMoney(30000, "USD")

	// Act
	reservation, err := svc.orchSvc.CompleteBooking(
		ctx,
		"res-001",
		"pay-001",
		"guest-001",
		"room-101",
		dateRange,
		amount,
		guests,
		"credit_card",
	)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "reservation must not be nil", reservation != nil, true)
	assert.That(t, "status must be confirmed", reservation.Status, booking.StatusConfirmed)

	_, resErr := svc.resRepo.Read(ctx, "res-001")
	assert.That(t, "reservation must be persisted", resErr == nil, true)

	_, payErr := svc.payRepo.Read(ctx, "pay-001")
	assert.That(t, "payment must be persisted", payErr == nil, true)

	assert.That(t, "events must be published", len(svc.publisher.events) > 0, true)
}

func Test_BookingOrchestrationService_CompleteBooking_Room_Unavailable_Should_Fail(t *testing.T) {
	// Arrange
	resRepo := newMockReservationRepository()
	payRepo := newMockPaymentRepository()
	availChecker := &mockAvailabilityChecker{available: false}
	payGateway := &mockPaymentGateway{}
	publisher := &mockEventPublisher{}
	notifSvc := &mockNotificationService{}

	resSvc := booking.NewReservationService(resRepo, availChecker, publisher)
	paySvc := booking.NewPaymentService(payRepo, payGateway, publisher)
	orchSvc := booking.NewBookingOrchestrationService(resSvc, paySvc, notifSvc)

	ctx := context.Background()
	dateRange := createTestDateRange()
	guests := createTestGuests()
	amount := booking.NewMoney(30000, "USD")

	// Act
	_, err := orchSvc.CompleteBooking(
		ctx,
		"res-001",
		"pay-001",
		"guest-001",
		"room-101",
		dateRange,
		amount,
		guests,
		"credit_card",
	)

	// Assert
	assert.That(t, "error must not be nil for unavailable room", err != nil, true)
}

func Test_BookingOrchestrationService_CompleteBooking_Payment_Auth_Failure_Should_Cancel_Reservation(t *testing.T) {
	// Arrange
	resRepo := newMockReservationRepository()
	payRepo := newMockPaymentRepository()
	availChecker := &mockAvailabilityChecker{available: true}
	payGateway := &mockPaymentGateway{authorizeErr: errors.New("payment declined")}
	publisher := &mockEventPublisher{}
	notifSvc := &mockNotificationService{}

	resSvc := booking.NewReservationService(resRepo, availChecker, publisher)
	paySvc := booking.NewPaymentService(payRepo, payGateway, publisher)
	orchSvc := booking.NewBookingOrchestrationService(resSvc, paySvc, notifSvc)

	ctx := context.Background()
	dateRange := createTestDateRange()
	guests := createTestGuests()
	amount := booking.NewMoney(30000, "USD")

	// Act
	_, err := orchSvc.CompleteBooking(
		ctx,
		"res-001",
		"pay-001",
		"guest-001",
		"room-101",
		dateRange,
		amount,
		guests,
		"credit_card",
	)

	// Assert
	assert.That(t, "error must not be nil for payment failure", err != nil, true)

	res, readErr := resRepo.Read(ctx, "res-001")
	assert.That(t, "reservation must exist", readErr == nil, true)
	assert.That(t, "reservation must be cancelled", res.Status, booking.StatusCancelled)
}

func Test_BookingOrchestrationService_CancelBookingWithRefund_Should_Succeed(t *testing.T) {
	// Arrange
	svc := createTestOrchestrationService(t)
	ctx := context.Background()
	dateRange := createTestDateRange()
	guests := createTestGuests()
	amount := booking.NewMoney(30000, "USD")

	_, setupErr := svc.orchSvc.CompleteBooking(
		ctx,
		"res-001",
		"pay-001",
		"guest-001",
		"room-101",
		dateRange,
		amount,
		guests,
		"credit_card",
	)
	if setupErr != nil {
		t.Fatalf("setup failed: %v", setupErr)
	}

	// Act
	err := svc.orchSvc.CancelBookingWithRefund(ctx, "res-001", "guest requested cancellation")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)

	res, readErr := svc.resRepo.Read(ctx, "res-001")
	assert.That(t, "reservation must exist", readErr == nil, true)
	assert.That(t, "reservation must be cancelled", res.Status, booking.StatusCancelled)
}

func Test_BookingOrchestrationService_CancelBookingWithRefund_NotFound_Should_Fail(t *testing.T) {
	// Arrange
	svc := createTestOrchestrationService(t)
	ctx := context.Background()

	// Act
	err := svc.orchSvc.CancelBookingWithRefund(ctx, "nonexistent", "test")

	// Assert
	assert.That(t, "error must not be nil for nonexistent reservation", err != nil, true)
}
