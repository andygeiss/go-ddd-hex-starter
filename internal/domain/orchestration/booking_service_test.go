package orchestration_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/event"
	"github.com/andygeiss/hotel-booking/internal/domain/orchestration"
	"github.com/andygeiss/hotel-booking/internal/domain/payment"
	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
	"github.com/andygeiss/hotel-booking/internal/domain/shared"
)

// ============================================================================
// Mock Implementations - Reservation
// ============================================================================

type mockReservationRepository struct {
	reservations map[reservation.ReservationID]reservation.Reservation
	createErr    error
	readErr      error
	updateErr    error
}

func newMockReservationRepository() *mockReservationRepository {
	return &mockReservationRepository{
		reservations: make(map[reservation.ReservationID]reservation.Reservation),
	}
}

func (m *mockReservationRepository) Create(ctx context.Context, id reservation.ReservationID, res reservation.Reservation) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.reservations[id] = res
	return nil
}

func (m *mockReservationRepository) Read(ctx context.Context, id reservation.ReservationID) (*reservation.Reservation, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	res, ok := m.reservations[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return &res, nil
}

func (m *mockReservationRepository) Update(ctx context.Context, id reservation.ReservationID, res reservation.Reservation) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.reservations[id] = res
	return nil
}

func (m *mockReservationRepository) Delete(ctx context.Context, id reservation.ReservationID) error {
	delete(m.reservations, id)
	return nil
}

func (m *mockReservationRepository) ReadAll(ctx context.Context) ([]reservation.Reservation, error) {
	result := make([]reservation.Reservation, 0, len(m.reservations))
	for _, res := range m.reservations {
		result = append(result, res)
	}
	return result, nil
}

type mockAvailabilityChecker struct {
	available bool
	err       error
}

func (m *mockAvailabilityChecker) IsRoomAvailable(ctx context.Context, roomID reservation.RoomID, dateRange reservation.DateRange) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.available, nil
}

func (m *mockAvailabilityChecker) GetOverlappingReservations(ctx context.Context, roomID reservation.RoomID, dateRange reservation.DateRange) ([]*reservation.Reservation, error) {
	return nil, nil
}

// ============================================================================
// Mock Implementations - Payment
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

// ============================================================================
// Mock Implementations - Common
// ============================================================================

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

type mockNotificationService struct {
	confirmationsSent int
	cancellationsSent int
	receiptsSent      int
	err               error
}

func (m *mockNotificationService) SendReservationConfirmation(ctx context.Context, r *reservation.Reservation) error {
	if m.err != nil {
		return m.err
	}
	m.confirmationsSent++
	return nil
}

func (m *mockNotificationService) SendCancellationNotice(ctx context.Context, r *reservation.Reservation, reason string) error {
	if m.err != nil {
		return m.err
	}
	m.cancellationsSent++
	return nil
}

func (m *mockNotificationService) SendPaymentReceipt(ctx context.Context, p *payment.Payment) error {
	if m.err != nil {
		return m.err
	}
	m.receiptsSent++
	return nil
}

// ============================================================================
// Test Helpers
// ============================================================================

type testServices struct {
	reservationRepo    *mockReservationRepository
	availabilityCheck  *mockAvailabilityChecker
	reservationPub     *mockEventPublisher
	reservationService *reservation.Service

	paymentRepo    *mockPaymentRepository
	paymentGateway *mockPaymentGateway
	paymentPub     *mockEventPublisher
	paymentService *payment.Service

	notificationService *mockNotificationService
	bookingService      *orchestration.BookingService
}

func createTestServices() *testServices {
	// Reservation context
	reservationRepo := newMockReservationRepository()
	availabilityChecker := &mockAvailabilityChecker{available: true}
	reservationPub := &mockEventPublisher{}
	reservationService := reservation.NewService(reservationRepo, availabilityChecker, reservationPub)

	// Payment context
	paymentRepo := newMockPaymentRepository()
	paymentGateway := &mockPaymentGateway{authorizeTransactionID: "tx-12345"}
	paymentPub := &mockEventPublisher{}
	paymentService := payment.NewService(paymentRepo, paymentGateway, paymentPub)

	// Orchestration
	notificationService := &mockNotificationService{}
	bookingService := orchestration.NewBookingService(reservationService, paymentService, notificationService)

	return &testServices{
		reservationRepo:     reservationRepo,
		availabilityCheck:   availabilityChecker,
		reservationPub:      reservationPub,
		reservationService:  reservationService,
		paymentRepo:         paymentRepo,
		paymentGateway:      paymentGateway,
		paymentPub:          paymentPub,
		paymentService:      paymentService,
		notificationService: notificationService,
		bookingService:      bookingService,
	}
}

func validBookingDateRange() reservation.DateRange {
	checkIn := time.Now().Add(48 * time.Hour).Truncate(24 * time.Hour)
	checkOut := checkIn.Add(72 * time.Hour)
	return reservation.NewDateRange(checkIn, checkOut)
}

func validBookingGuests() []reservation.GuestInfo {
	return []reservation.GuestInfo{
		reservation.NewGuestInfo("John Doe", "john@example.com", "+1234567890"),
	}
}

func validBookingMoney() shared.Money {
	return shared.NewMoney(10000, "USD")
}

// ============================================================================
// InitiateBooking Tests
// ============================================================================

func Test_BookingService_InitiateBooking_Should_Create_Reservation(t *testing.T) {
	// Arrange
	svc := createTestServices()
	ctx := context.Background()
	reservationID := shared.ReservationID("res-001")

	// Act
	res, err := svc.bookingService.InitiateBooking(
		ctx,
		reservationID,
		"guest-001",
		"room-101",
		validBookingDateRange(),
		validBookingMoney(),
		validBookingGuests(),
	)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "reservation must not be nil", res != nil, true)
	assert.That(t, "reservation ID must match", string(res.ID), string(reservationID))
	assert.That(t, "status must be pending", res.Status, reservation.StatusPending)
}

func Test_BookingService_InitiateBooking_Should_Publish_Event(t *testing.T) {
	// Arrange
	svc := createTestServices()
	ctx := context.Background()

	// Act
	_, err := svc.bookingService.InitiateBooking(
		ctx,
		"res-001",
		"guest-001",
		"room-101",
		validBookingDateRange(),
		validBookingMoney(),
		validBookingGuests(),
	)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "reservation event must be published", len(svc.reservationPub.published), 1)
}

func Test_BookingService_InitiateBooking_When_Reservation_Fails_Should_Return_Error(t *testing.T) {
	// Arrange
	svc := createTestServices()
	svc.availabilityCheck.available = false
	ctx := context.Background()

	// Act
	res, err := svc.bookingService.InitiateBooking(
		ctx,
		"res-001",
		"guest-001",
		"room-101",
		validBookingDateRange(),
		validBookingMoney(),
		validBookingGuests(),
	)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "reservation must be nil", res == nil, true)
}

// ============================================================================
// CompleteBooking Tests
// ============================================================================

func Test_BookingService_CompleteBooking_Should_Complete_Full_Saga(t *testing.T) {
	// Arrange
	svc := createTestServices()
	ctx := context.Background()
	reservationID := shared.ReservationID("res-001")
	paymentID := payment.PaymentID("pay-001")

	// Act
	res, err := svc.bookingService.CompleteBooking(
		ctx,
		reservationID,
		paymentID,
		"guest-001",
		"room-101",
		validBookingDateRange(),
		validBookingMoney(),
		validBookingGuests(),
		"credit_card",
	)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "reservation must not be nil", res != nil, true)
	assert.That(t, "status must be confirmed", res.Status, reservation.StatusConfirmed)
}

func Test_BookingService_CompleteBooking_Should_Send_Confirmation_Notification(t *testing.T) {
	// Arrange
	svc := createTestServices()
	ctx := context.Background()

	// Act
	_, err := svc.bookingService.CompleteBooking(
		ctx,
		"res-001",
		"pay-001",
		"guest-001",
		"room-101",
		validBookingDateRange(),
		validBookingMoney(),
		validBookingGuests(),
		"credit_card",
	)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "confirmation must be sent", svc.notificationService.confirmationsSent, 1)
}

func Test_BookingService_CompleteBooking_When_Payment_Authorization_Fails_Should_Cancel_Reservation(t *testing.T) {
	// Arrange
	svc := createTestServices()
	svc.paymentGateway.authorizeErr = errors.New("gateway error")
	ctx := context.Background()

	// Act
	res, err := svc.bookingService.CompleteBooking(
		ctx,
		"res-001",
		"pay-001",
		"guest-001",
		"room-101",
		validBookingDateRange(),
		validBookingMoney(),
		validBookingGuests(),
		"credit_card",
	)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "reservation must be nil", res == nil, true)

	// Check compensation occurred
	storedRes, _ := svc.reservationRepo.Read(ctx, "res-001")
	assert.That(t, "reservation must be cancelled", storedRes.Status, reservation.StatusCancelled)
}

func Test_BookingService_CompleteBooking_When_Capture_Fails_Should_Cancel_Reservation(t *testing.T) {
	// Arrange
	svc := createTestServices()
	svc.paymentGateway.captureErr = errors.New("capture failed")
	ctx := context.Background()

	// Act
	res, err := svc.bookingService.CompleteBooking(
		ctx,
		"res-001",
		"pay-001",
		"guest-001",
		"room-101",
		validBookingDateRange(),
		validBookingMoney(),
		validBookingGuests(),
		"credit_card",
	)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "reservation must be nil", res == nil, true)

	// Check compensation occurred
	storedRes, _ := svc.reservationRepo.Read(ctx, "res-001")
	assert.That(t, "reservation must be cancelled", storedRes.Status, reservation.StatusCancelled)
}

// ============================================================================
// CancelBookingWithRefund Tests
// ============================================================================

func Test_BookingService_CancelBookingWithRefund_Should_Cancel_Reservation(t *testing.T) {
	// Arrange
	svc := createTestServices()
	ctx := context.Background()
	reservationID := shared.ReservationID("res-001")

	// First create a reservation
	_, _ = svc.bookingService.InitiateBooking(
		ctx,
		reservationID,
		"guest-001",
		"room-101",
		validBookingDateRange(),
		validBookingMoney(),
		validBookingGuests(),
	)

	// Act
	err := svc.bookingService.CancelBookingWithRefund(ctx, reservationID, "guest requested")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)

	storedRes, _ := svc.reservationRepo.Read(ctx, reservationID)
	assert.That(t, "status must be cancelled", storedRes.Status, reservation.StatusCancelled)
}

func Test_BookingService_CancelBookingWithRefund_Should_Send_Cancellation_Notice(t *testing.T) {
	// Arrange
	svc := createTestServices()
	ctx := context.Background()
	reservationID := shared.ReservationID("res-001")

	_, _ = svc.bookingService.InitiateBooking(
		ctx,
		reservationID,
		"guest-001",
		"room-101",
		validBookingDateRange(),
		validBookingMoney(),
		validBookingGuests(),
	)

	// Act
	err := svc.bookingService.CancelBookingWithRefund(ctx, reservationID, "guest requested")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "cancellation notice must be sent", svc.notificationService.cancellationsSent, 1)
}

func Test_BookingService_CancelBookingWithRefund_When_Not_Found_Should_Return_Error(t *testing.T) {
	// Arrange
	svc := createTestServices()
	ctx := context.Background()

	// Act
	err := svc.bookingService.CancelBookingWithRefund(ctx, "non-existent", "test")

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

// ============================================================================
// OnPaymentAuthorized Tests
// ============================================================================

func Test_BookingService_OnPaymentAuthorized_Should_Capture_Payment(t *testing.T) {
	// Arrange
	svc := createTestServices()
	ctx := context.Background()
	reservationID := shared.ReservationID("res-001")
	paymentID := payment.PaymentID("pay-001")

	// Setup: create reservation and authorize payment
	_, _ = svc.bookingService.InitiateBooking(
		ctx, reservationID, "guest-001", "room-101",
		validBookingDateRange(), validBookingMoney(), validBookingGuests(),
	)
	_, _ = svc.paymentService.AuthorizePayment(ctx, paymentID, reservationID, validBookingMoney(), "credit_card")

	// Act
	err := svc.bookingService.OnPaymentAuthorized(ctx, paymentID, reservationID)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)

	storedPayment, _ := svc.paymentRepo.Read(ctx, paymentID)
	assert.That(t, "payment must be captured", storedPayment.Status, payment.StatusCaptured)
}

func Test_BookingService_OnPaymentAuthorized_When_Capture_Fails_Should_Cancel_Reservation(t *testing.T) {
	// Arrange
	svc := createTestServices()
	svc.paymentGateway.captureErr = errors.New("capture failed")
	ctx := context.Background()
	reservationID := shared.ReservationID("res-001")
	paymentID := payment.PaymentID("pay-001")

	// Setup: create reservation and authorize payment
	_, _ = svc.bookingService.InitiateBooking(
		ctx, reservationID, "guest-001", "room-101",
		validBookingDateRange(), validBookingMoney(), validBookingGuests(),
	)
	_, _ = svc.paymentService.AuthorizePayment(ctx, paymentID, reservationID, validBookingMoney(), "credit_card")

	// Act
	err := svc.bookingService.OnPaymentAuthorized(ctx, paymentID, reservationID)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)

	storedRes, _ := svc.reservationRepo.Read(ctx, reservationID)
	assert.That(t, "reservation must be cancelled", storedRes.Status, reservation.StatusCancelled)
}

// ============================================================================
// OnPaymentCaptured Tests
// ============================================================================

func Test_BookingService_OnPaymentCaptured_Should_Confirm_Reservation(t *testing.T) {
	// Arrange
	svc := createTestServices()
	ctx := context.Background()
	reservationID := shared.ReservationID("res-001")

	// Setup: create reservation
	_, _ = svc.bookingService.InitiateBooking(
		ctx, reservationID, "guest-001", "room-101",
		validBookingDateRange(), validBookingMoney(), validBookingGuests(),
	)

	// Act
	err := svc.bookingService.OnPaymentCaptured(ctx, reservationID)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)

	storedRes, _ := svc.reservationRepo.Read(ctx, reservationID)
	assert.That(t, "reservation must be confirmed", storedRes.Status, reservation.StatusConfirmed)
}

func Test_BookingService_OnPaymentCaptured_Should_Send_Notification(t *testing.T) {
	// Arrange
	svc := createTestServices()
	ctx := context.Background()
	reservationID := shared.ReservationID("res-001")

	_, _ = svc.bookingService.InitiateBooking(
		ctx, reservationID, "guest-001", "room-101",
		validBookingDateRange(), validBookingMoney(), validBookingGuests(),
	)

	// Act
	err := svc.bookingService.OnPaymentCaptured(ctx, reservationID)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "confirmation must be sent", svc.notificationService.confirmationsSent, 1)
}

// ============================================================================
// OnPaymentFailed Tests
// ============================================================================

func Test_BookingService_OnPaymentFailed_Should_Cancel_Reservation(t *testing.T) {
	// Arrange
	svc := createTestServices()
	ctx := context.Background()
	reservationID := shared.ReservationID("res-001")

	// Setup: create reservation
	_, _ = svc.bookingService.InitiateBooking(
		ctx, reservationID, "guest-001", "room-101",
		validBookingDateRange(), validBookingMoney(), validBookingGuests(),
	)

	// Act
	err := svc.bookingService.OnPaymentFailed(ctx, reservationID, "payment_declined")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)

	storedRes, _ := svc.reservationRepo.Read(ctx, reservationID)
	assert.That(t, "reservation must be cancelled", storedRes.Status, reservation.StatusCancelled)
	assert.That(t, "cancellation reason must match", storedRes.CancellationReason, "payment_declined")
}

func Test_BookingService_OnPaymentFailed_When_Reservation_Not_Found_Should_Return_Error(t *testing.T) {
	// Arrange
	svc := createTestServices()
	ctx := context.Background()

	// Act
	err := svc.bookingService.OnPaymentFailed(ctx, "non-existent", "test")

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}
