package outbound_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/outbound"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/booking"
)

// ============================================================================
// FilePaymentRepository Tests
// ============================================================================

const testPaymentResID001 = "res-001"

func createPaymentTempFile(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	return filepath.Join(tmpDir, "payments.json")
}

func createSamplePayment(id string) booking.Payment {
	payment := booking.NewPayment(
		booking.PaymentID(id),
		testPaymentResID001,
		booking.NewMoney(30000, "USD"),
		"credit_card",
	)
	return *payment
}

func Test_FilePaymentRepository_Create_And_Read_Should_Succeed(t *testing.T) {
	// Arrange
	filename := createPaymentTempFile(t)
	repo := outbound.NewFilePaymentRepository(filename)
	ctx := context.Background()
	payment := createSamplePayment("pay-001")

	// Act
	err := repo.Create(ctx, "pay-001", payment)

	// Assert
	assert.That(t, "error on create must be nil", err == nil, true)

	read, readErr := repo.Read(ctx, "pay-001")
	assert.That(t, "error on read must be nil", readErr == nil, true)
	assert.That(t, "ID must match", string(read.ID), "pay-001")
	assert.That(t, "ReservationID must match", string(read.ReservationID), testPaymentResID001)
	assert.That(t, "Amount must match", read.Amount.Amount, int64(30000))
}

func Test_FilePaymentRepository_Update_Should_Modify_Payment(t *testing.T) {
	// Arrange
	filename := createPaymentTempFile(t)
	repo := outbound.NewFilePaymentRepository(filename)
	ctx := context.Background()
	payment := createSamplePayment("pay-001")

	setupErr := repo.Create(ctx, "pay-001", payment)
	if setupErr != nil {
		t.Fatalf("setup failed: %v", setupErr)
	}

	payment.Status = booking.PaymentAuthorized
	payment.TransactionID = "txn-123"

	// Act
	err := repo.Update(ctx, "pay-001", payment)

	// Assert
	assert.That(t, "error on update must be nil", err == nil, true)

	read, readErr := repo.Read(ctx, "pay-001")
	assert.That(t, "error on read must be nil", readErr == nil, true)
	assert.That(t, "status must be authorized", read.Status, booking.PaymentAuthorized)
	assert.That(t, "transaction ID must match", read.TransactionID, "txn-123")
}

func Test_FilePaymentRepository_Delete_Should_Remove_Payment(t *testing.T) {
	// Arrange
	filename := createPaymentTempFile(t)
	repo := outbound.NewFilePaymentRepository(filename)
	ctx := context.Background()
	payment := createSamplePayment("pay-001")

	setupErr := repo.Create(ctx, "pay-001", payment)
	if setupErr != nil {
		t.Fatalf("setup failed: %v", setupErr)
	}

	// Act
	err := repo.Delete(ctx, "pay-001")

	// Assert
	assert.That(t, "error on delete must be nil", err == nil, true)

	_, readErr := repo.Read(ctx, "pay-001")
	assert.That(t, "error must not be nil for deleted payment", readErr != nil, true)
}

func Test_FilePaymentRepository_ReadAll_Should_Return_All_Payments(t *testing.T) {
	// Arrange
	filename := createPaymentTempFile(t)
	repo := outbound.NewFilePaymentRepository(filename)
	ctx := context.Background()

	pay1 := createSamplePayment("pay-001")
	pay2 := createSamplePayment("pay-002")
	pay2.ReservationID = "res-002"

	_ = repo.Create(ctx, "pay-001", pay1)
	_ = repo.Create(ctx, "pay-002", pay2)

	// Act
	all, err := repo.ReadAll(ctx)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "must return 2 payments", len(all), 2)
}

func Test_FilePaymentRepository_Read_NonExistent_Should_Return_Error(t *testing.T) {
	// Arrange
	filename := createPaymentTempFile(t)
	repo := outbound.NewFilePaymentRepository(filename)
	ctx := context.Background()

	// Act
	_, err := repo.Read(ctx, "nonexistent")

	// Assert
	assert.That(t, "error must not be nil for nonexistent payment", err != nil, true)
}

func Test_FilePaymentRepository_Persistence_Across_Instances(t *testing.T) {
	// Arrange
	filename := createPaymentTempFile(t)
	ctx := context.Background()

	repo1 := outbound.NewFilePaymentRepository(filename)
	payment := createSamplePayment("pay-001")
	_ = repo1.Create(ctx, "pay-001", payment)

	// Act
	repo2 := outbound.NewFilePaymentRepository(filename)
	read, err := repo2.Read(ctx, "pay-001")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "ID must match", string(read.ID), "pay-001")
}

func Test_FilePaymentRepository_Create_Multiple_Should_Not_Overwrite_Others(t *testing.T) {
	// Arrange
	filename := createPaymentTempFile(t)
	repo := outbound.NewFilePaymentRepository(filename)
	ctx := context.Background()

	pay1 := createSamplePayment("pay-001")
	pay2 := createSamplePayment("pay-002")
	pay3 := createSamplePayment("pay-003")

	_ = repo.Create(ctx, "pay-001", pay1)
	_ = repo.Create(ctx, "pay-002", pay2)
	_ = repo.Create(ctx, "pay-003", pay3)

	// Act
	_, err1 := repo.Read(ctx, "pay-001")
	_, err2 := repo.Read(ctx, "pay-002")
	_, err3 := repo.Read(ctx, "pay-003")
	all, _ := repo.ReadAll(ctx)

	// Assert
	assert.That(t, "pay-001 must be readable", err1 == nil, true)
	assert.That(t, "pay-002 must be readable", err2 == nil, true)
	assert.That(t, "pay-003 must be readable", err3 == nil, true)
	assert.That(t, "must return 3 payments", len(all), 3)
}

func Test_FilePaymentRepository_Update_With_Payment_Attempts(t *testing.T) {
	// Arrange
	filename := createPaymentTempFile(t)
	repo := outbound.NewFilePaymentRepository(filename)
	ctx := context.Background()
	payment := createSamplePayment("pay-001")

	_ = repo.Create(ctx, "pay-001", payment)

	payment.Attempts = append(payment.Attempts, booking.NewPaymentAttempt(
		booking.PaymentAuthorized,
		"",
		"",
	))

	// Act
	_ = repo.Update(ctx, "pay-001", payment)
	read, err := repo.Read(ctx, "pay-001")

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "must have 1 attempt", len(read.Attempts), 1)
}
