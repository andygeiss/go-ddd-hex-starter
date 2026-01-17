package payment_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/hotel-booking/internal/domain/payment"
	"github.com/andygeiss/hotel-booking/internal/domain/shared"
)

// ============================================================================
// Test Helpers
// ============================================================================

func validMoney() shared.Money {
	return shared.NewMoney(10000, "USD")
}

func createValidPayment() *payment.Payment {
	return payment.NewPayment(
		"pay-001",
		"res-001",
		validMoney(),
		"credit_card",
	)
}

// ============================================================================
// NewPayment Tests
// ============================================================================

func Test_NewPayment_Should_Return_Valid_Payment(t *testing.T) {
	// Arrange
	id := payment.PaymentID("pay-001")
	reservationID := payment.ReservationID("res-001")
	amount := validMoney()
	method := "credit_card"

	// Act
	p := payment.NewPayment(id, reservationID, amount, method)

	// Assert
	assert.That(t, "payment must not be nil", p != nil, true)
	assert.That(t, "ID must match", p.ID, id)
	assert.That(t, "ReservationID must match", p.ReservationID, reservationID)
	assert.That(t, "Amount must match", p.Amount, amount)
	assert.That(t, "PaymentMethod must match", p.PaymentMethod, method)
	assert.That(t, "Status must be pending", p.Status, payment.StatusPending)
	assert.That(t, "Attempts must be empty", len(p.Attempts), 0)
}

func Test_NewPayment_Should_Initialize_Empty_TransactionID(t *testing.T) {
	// Arrange & Act
	p := createValidPayment()

	// Assert
	assert.That(t, "TransactionID must be empty", p.TransactionID, "")
}

// ============================================================================
// State Transition Tests - Authorize
// ============================================================================

func Test_Payment_Authorize_Should_Set_Status_Authorized(t *testing.T) {
	// Arrange
	p := createValidPayment()
	transactionID := "tx-12345"

	// Act
	err := p.Authorize(transactionID)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "status must be authorized", p.Status, payment.StatusAuthorized)
	assert.That(t, "TransactionID must be set", p.TransactionID, transactionID)
}

func Test_Payment_Authorize_Should_Add_Attempt(t *testing.T) {
	// Arrange
	p := createValidPayment()

	// Act
	_ = p.Authorize("tx-12345")

	// Assert
	assert.That(t, "must have 1 attempt", len(p.Attempts), 1)
	assert.That(t, "attempt status must be authorized", p.Attempts[0].Status, payment.StatusAuthorized)
}

func Test_Payment_Authorize_When_Already_Authorized_Should_Return_Error(t *testing.T) {
	// Arrange
	p := createValidPayment()
	_ = p.Authorize("tx-12345")

	// Act
	err := p.Authorize("tx-67890")

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "status must remain authorized", p.Status, payment.StatusAuthorized)
}

func Test_Payment_Authorize_From_Captured_Should_Return_Error(t *testing.T) {
	// Arrange
	p := createValidPayment()
	_ = p.Authorize("tx-12345")
	_ = p.Capture()

	// Act
	err := p.Authorize("tx-67890")

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "status must remain captured", p.Status, payment.StatusCaptured)
}

func Test_Payment_Authorize_From_Failed_Should_Succeed(t *testing.T) {
	// Arrange
	p := createValidPayment()
	_ = p.Fail("error", "error message")

	// Act
	err := p.Authorize("tx-12345")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "status must be authorized", p.Status, payment.StatusAuthorized)
}

// ============================================================================
// State Transition Tests - Capture
// ============================================================================

func Test_Payment_Capture_From_Authorized_Should_Succeed(t *testing.T) {
	// Arrange
	p := createValidPayment()
	_ = p.Authorize("tx-12345")

	// Act
	err := p.Capture()

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "status must be captured", p.Status, payment.StatusCaptured)
}

func Test_Payment_Capture_Should_Add_Attempt(t *testing.T) {
	// Arrange
	p := createValidPayment()
	_ = p.Authorize("tx-12345")

	// Act
	_ = p.Capture()

	// Assert
	assert.That(t, "must have 2 attempts", len(p.Attempts), 2)
	assert.That(t, "second attempt status must be captured", p.Attempts[1].Status, payment.StatusCaptured)
}

func Test_Payment_Capture_From_Pending_Should_Return_Error(t *testing.T) {
	// Arrange
	p := createValidPayment()

	// Act
	err := p.Capture()

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "status must remain pending", p.Status, payment.StatusPending)
}

func Test_Payment_Capture_When_Already_Captured_Should_Return_Error(t *testing.T) {
	// Arrange
	p := createValidPayment()
	_ = p.Authorize("tx-12345")
	_ = p.Capture()

	// Act
	err := p.Capture()

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "status must remain captured", p.Status, payment.StatusCaptured)
}

// ============================================================================
// State Transition Tests - Fail
// ============================================================================

func Test_Payment_Fail_Should_Set_Status_Failed(t *testing.T) {
	// Arrange
	p := createValidPayment()
	errorCode := "insufficient_funds"
	errorMsg := "Card declined"

	// Act
	err := p.Fail(errorCode, errorMsg)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "status must be failed", p.Status, payment.StatusFailed)
}

func Test_Payment_Fail_Should_Add_Attempt_With_Error_Details(t *testing.T) {
	// Arrange
	p := createValidPayment()
	errorCode := "insufficient_funds"
	errorMsg := "Card declined"

	// Act
	_ = p.Fail(errorCode, errorMsg)

	// Assert
	assert.That(t, "must have 1 attempt", len(p.Attempts), 1)
	assert.That(t, "attempt status must be failed", p.Attempts[0].Status, payment.StatusFailed)
	assert.That(t, "error code must match", p.Attempts[0].ErrorCode, errorCode)
	assert.That(t, "error message must match", p.Attempts[0].ErrorMsg, errorMsg)
}

func Test_Payment_Fail_From_Authorized_Should_Succeed(t *testing.T) {
	// Arrange
	p := createValidPayment()
	_ = p.Authorize("tx-12345")

	// Act
	err := p.Fail("timeout", "Gateway timeout")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "status must be failed", p.Status, payment.StatusFailed)
}

func Test_Payment_Fail_From_Captured_Should_Return_Error(t *testing.T) {
	// Arrange
	p := createValidPayment()
	_ = p.Authorize("tx-12345")
	_ = p.Capture()

	// Act
	err := p.Fail("error", "error message")

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "status must remain captured", p.Status, payment.StatusCaptured)
}

func Test_Payment_Fail_From_Refunded_Should_Return_Error(t *testing.T) {
	// Arrange
	p := createValidPayment()
	_ = p.Authorize("tx-12345")
	_ = p.Capture()
	_ = p.Refund()

	// Act
	err := p.Fail("error", "error message")

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "status must remain refunded", p.Status, payment.StatusRefunded)
}

// ============================================================================
// State Transition Tests - Refund
// ============================================================================

func Test_Payment_Refund_From_Captured_Should_Succeed(t *testing.T) {
	// Arrange
	p := createValidPayment()
	_ = p.Authorize("tx-12345")
	_ = p.Capture()

	// Act
	err := p.Refund()

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "status must be refunded", p.Status, payment.StatusRefunded)
}

func Test_Payment_Refund_Should_Add_Attempt(t *testing.T) {
	// Arrange
	p := createValidPayment()
	_ = p.Authorize("tx-12345")
	_ = p.Capture()

	// Act
	_ = p.Refund()

	// Assert
	assert.That(t, "must have 3 attempts", len(p.Attempts), 3)
	assert.That(t, "third attempt status must be refunded", p.Attempts[2].Status, payment.StatusRefunded)
}

func Test_Payment_Refund_From_Pending_Should_Return_Error(t *testing.T) {
	// Arrange
	p := createValidPayment()

	// Act
	err := p.Refund()

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "status must remain pending", p.Status, payment.StatusPending)
}

func Test_Payment_Refund_From_Authorized_Should_Return_Error(t *testing.T) {
	// Arrange
	p := createValidPayment()
	_ = p.Authorize("tx-12345")

	// Act
	err := p.Refund()

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "status must remain authorized", p.Status, payment.StatusAuthorized)
}

func Test_Payment_Refund_When_Already_Refunded_Should_Return_Error(t *testing.T) {
	// Arrange
	p := createValidPayment()
	_ = p.Authorize("tx-12345")
	_ = p.Capture()
	_ = p.Refund()

	// Act
	err := p.Refund()

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "status must remain refunded", p.Status, payment.StatusRefunded)
}

// ============================================================================
// Business Logic Tests
// ============================================================================

func Test_Payment_IsSuccessful_When_Captured_Should_Return_True(t *testing.T) {
	// Arrange
	p := createValidPayment()
	_ = p.Authorize("tx-12345")
	_ = p.Capture()

	// Act
	result := p.IsSuccessful()

	// Assert
	assert.That(t, "should be successful", result, true)
}

func Test_Payment_IsSuccessful_When_Not_Captured_Should_Return_False(t *testing.T) {
	// Arrange
	p := createValidPayment()
	_ = p.Authorize("tx-12345")

	// Act
	result := p.IsSuccessful()

	// Assert
	assert.That(t, "should not be successful", result, false)
}

func Test_Payment_CanBeRetried_When_Pending_Should_Return_True(t *testing.T) {
	// Arrange
	p := createValidPayment()

	// Act
	result := p.CanBeRetried()

	// Assert
	assert.That(t, "should be retryable", result, true)
}

func Test_Payment_CanBeRetried_When_Failed_Under_Max_Attempts_Should_Return_True(t *testing.T) {
	// Arrange
	p := createValidPayment()
	_ = p.Fail("error1", "first failure")

	// Act
	result := p.CanBeRetried()

	// Assert
	assert.That(t, "should be retryable", result, true)
}

func Test_Payment_CanBeRetried_When_At_Max_Attempts_Should_Return_False(t *testing.T) {
	// Arrange
	p := createValidPayment()
	_ = p.Fail("error1", "first failure")
	// Reset to pending-like state to fail again
	p.Status = payment.StatusPending
	_ = p.Fail("error2", "second failure")
	p.Status = payment.StatusPending
	_ = p.Fail("error3", "third failure")

	// Act
	result := p.CanBeRetried()

	// Assert
	assert.That(t, "should not be retryable", result, false)
}

func Test_Payment_CanBeRetried_When_Authorized_Should_Return_False(t *testing.T) {
	// Arrange
	p := createValidPayment()
	_ = p.Authorize("tx-12345")

	// Act
	result := p.CanBeRetried()

	// Assert
	assert.That(t, "should not be retryable", result, false)
}

func Test_Payment_CanBeRetried_When_Captured_Should_Return_False(t *testing.T) {
	// Arrange
	p := createValidPayment()
	_ = p.Authorize("tx-12345")
	_ = p.Capture()

	// Act
	result := p.CanBeRetried()

	// Assert
	assert.That(t, "should not be retryable", result, false)
}

// ============================================================================
// Entity Tests - PaymentAttempt
// ============================================================================

func Test_NewPaymentAttempt_Should_Create_Valid_Attempt(t *testing.T) {
	// Arrange
	status := payment.StatusFailed
	errorCode := "timeout"
	errorMsg := "Gateway timeout"

	// Act
	attempt := payment.NewPaymentAttempt(status, errorCode, errorMsg)

	// Assert
	assert.That(t, "Status must match", attempt.Status, status)
	assert.That(t, "ErrorCode must match", attempt.ErrorCode, errorCode)
	assert.That(t, "ErrorMsg must match", attempt.ErrorMsg, errorMsg)
	assert.That(t, "AttemptedAt must be set", attempt.AttemptedAt.IsZero(), false)
}

// ============================================================================
// Money Tests (shared)
// ============================================================================

func Test_NewMoney_Convenience_Function_Should_Work(t *testing.T) {
	// Arrange & Act
	money := payment.NewMoney(5000, "EUR")

	// Assert
	assert.That(t, "Amount must match", money.Amount, int64(5000))
	assert.That(t, "Currency must be uppercase", money.Currency, "EUR")
}

// ============================================================================
// Event Topic Tests
// ============================================================================

func Test_EventAuthorized_Topic_Should_Return_Correct_Value(t *testing.T) {
	// Arrange
	evt := payment.NewEventAuthorized()

	// Act
	topic := evt.Topic()

	// Assert
	assert.That(t, "topic must be payment.authorized", topic, "payment.authorized")
}

func Test_EventCaptured_Topic_Should_Return_Correct_Value(t *testing.T) {
	// Arrange
	evt := payment.NewEventCaptured()

	// Act
	topic := evt.Topic()

	// Assert
	assert.That(t, "topic must be payment.captured", topic, "payment.captured")
}

func Test_EventFailed_Topic_Should_Return_Correct_Value(t *testing.T) {
	// Arrange
	evt := payment.NewEventFailed()

	// Act
	topic := evt.Topic()

	// Assert
	assert.That(t, "topic must be payment.failed", topic, "payment.failed")
}

func Test_EventRefunded_Topic_Should_Return_Correct_Value(t *testing.T) {
	// Arrange
	evt := payment.NewEventRefunded()

	// Act
	topic := evt.Topic()

	// Assert
	assert.That(t, "topic must be payment.refunded", topic, "payment.refunded")
}

// ============================================================================
// Event Builder Tests
// ============================================================================

func Test_EventAuthorized_Builder_Should_Set_All_Fields(t *testing.T) {
	// Arrange & Act
	evt := payment.NewEventAuthorized().
		WithPaymentID("pay-001").
		WithReservationID("res-001").
		WithTransactionID("tx-12345").
		WithAmount(validMoney())

	// Assert
	assert.That(t, "PaymentID must match", evt.PaymentID, payment.PaymentID("pay-001"))
	assert.That(t, "ReservationID must match", evt.ReservationID, payment.ReservationID("res-001"))
	assert.That(t, "TransactionID must match", evt.TransactionID, "tx-12345")
	assert.That(t, "Amount must match", evt.Amount, validMoney())
}

func Test_EventCaptured_Builder_Should_Set_All_Fields(t *testing.T) {
	// Arrange & Act
	evt := payment.NewEventCaptured().
		WithPaymentID("pay-001").
		WithReservationID("res-001").
		WithAmount(validMoney())

	// Assert
	assert.That(t, "PaymentID must match", evt.PaymentID, payment.PaymentID("pay-001"))
	assert.That(t, "ReservationID must match", evt.ReservationID, payment.ReservationID("res-001"))
	assert.That(t, "Amount must match", evt.Amount, validMoney())
}

func Test_EventFailed_Builder_Should_Set_All_Fields(t *testing.T) {
	// Arrange & Act
	evt := payment.NewEventFailed().
		WithPaymentID("pay-001").
		WithReservationID("res-001").
		WithErrorCode("declined").
		WithErrorMsg("Card declined")

	// Assert
	assert.That(t, "PaymentID must match", evt.PaymentID, payment.PaymentID("pay-001"))
	assert.That(t, "ReservationID must match", evt.ReservationID, payment.ReservationID("res-001"))
	assert.That(t, "ErrorCode must match", evt.ErrorCode, "declined")
	assert.That(t, "ErrorMsg must match", evt.ErrorMsg, "Card declined")
}

func Test_EventRefunded_Builder_Should_Set_All_Fields(t *testing.T) {
	// Arrange & Act
	evt := payment.NewEventRefunded().
		WithPaymentID("pay-001").
		WithReservationID("res-001").
		WithAmount(validMoney())

	// Assert
	assert.That(t, "PaymentID must match", evt.PaymentID, payment.PaymentID("pay-001"))
	assert.That(t, "ReservationID must match", evt.ReservationID, payment.ReservationID("res-001"))
	assert.That(t, "Amount must match", evt.Amount, validMoney())
}
