package outbound

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/andygeiss/hotel-booking/internal/domain/payment"
)

// PostgresPaymentRepository implements the PaymentRepository port using PostgreSQL.
type PostgresPaymentRepository struct {
	db *sql.DB
}

// NewPostgresPaymentRepository creates a new PostgreSQL-based payment repository.
func NewPostgresPaymentRepository(db *sql.DB) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

// Create inserts a new payment into the database.
func (r *PostgresPaymentRepository) Create(ctx context.Context, id payment.PaymentID, value payment.Payment) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert payment
	_, err = tx.ExecContext(ctx, `
		INSERT INTO payments (
			id, reservation_id, amount_cents, amount_currency, status,
			payment_method, transaction_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		value.ID, value.ReservationID,
		value.Amount.Amount, value.Amount.Currency, value.Status,
		value.PaymentMethod, value.TransactionID,
		value.CreatedAt, value.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert payment: %w", err)
	}

	// Insert payment attempts
	for _, attempt := range value.Attempts {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO payment_attempts (payment_id, attempted_at, status, error_code, error_msg)
			VALUES ($1, $2, $3, $4, $5)
		`, value.ID, attempt.AttemptedAt, attempt.Status, attempt.ErrorCode, attempt.ErrorMsg)
		if err != nil {
			return fmt.Errorf("failed to insert payment attempt: %w", err)
		}
	}

	return tx.Commit()
}

// Read retrieves a payment by ID.
func (r *PostgresPaymentRepository) Read(ctx context.Context, id payment.PaymentID) (*payment.Payment, error) {
	var pay payment.Payment
	var paymentMethod, transactionID sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT id, reservation_id, amount_cents, amount_currency, status,
			payment_method, transaction_id, created_at, updated_at
		FROM payments WHERE id = $1
	`, id).Scan(
		&pay.ID, &pay.ReservationID,
		&pay.Amount.Amount, &pay.Amount.Currency, &pay.Status,
		&paymentMethod, &transactionID,
		&pay.CreatedAt, &pay.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read payment: %w", err)
	}

	if paymentMethod.Valid {
		pay.PaymentMethod = paymentMethod.String
	}
	if transactionID.Valid {
		pay.TransactionID = transactionID.String
	}

	// Load payment attempts
	attempts, err := r.loadAttempts(ctx, id)
	if err != nil {
		return nil, err
	}
	pay.Attempts = attempts

	return &pay, nil
}

// ReadAll retrieves all payments.
func (r *PostgresPaymentRepository) ReadAll(ctx context.Context) ([]payment.Payment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, reservation_id, amount_cents, amount_currency, status,
			payment_method, transaction_id, created_at, updated_at
		FROM payments
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query payments: %w", err)
	}
	defer rows.Close()

	var payments []payment.Payment
	for rows.Next() {
		var pay payment.Payment
		var paymentMethod, transactionID sql.NullString

		if err := rows.Scan(
			&pay.ID, &pay.ReservationID,
			&pay.Amount.Amount, &pay.Amount.Currency, &pay.Status,
			&paymentMethod, &transactionID,
			&pay.CreatedAt, &pay.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}

		if paymentMethod.Valid {
			pay.PaymentMethod = paymentMethod.String
		}
		if transactionID.Valid {
			pay.TransactionID = transactionID.String
		}

		payments = append(payments, pay)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payments: %w", err)
	}

	// Load attempts for each payment
	for i := range payments {
		attempts, err := r.loadAttempts(ctx, payments[i].ID)
		if err != nil {
			return nil, err
		}
		payments[i].Attempts = attempts
	}

	return payments, nil
}

// Update modifies an existing payment.
func (r *PostgresPaymentRepository) Update(ctx context.Context, id payment.PaymentID, value payment.Payment) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		UPDATE payments SET
			reservation_id = $2, amount_cents = $3, amount_currency = $4,
			status = $5, payment_method = $6, transaction_id = $7, updated_at = $8
		WHERE id = $1
	`,
		id, value.ReservationID,
		value.Amount.Amount, value.Amount.Currency, value.Status,
		value.PaymentMethod, value.TransactionID, value.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("payment not found: %s", id)
	}

	// Delete and re-insert attempts
	_, err = tx.ExecContext(ctx, `DELETE FROM payment_attempts WHERE payment_id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete payment attempts: %w", err)
	}

	for _, attempt := range value.Attempts {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO payment_attempts (payment_id, attempted_at, status, error_code, error_msg)
			VALUES ($1, $2, $3, $4, $5)
		`, id, attempt.AttemptedAt, attempt.Status, attempt.ErrorCode, attempt.ErrorMsg)
		if err != nil {
			return fmt.Errorf("failed to insert payment attempt: %w", err)
		}
	}

	return tx.Commit()
}

// Delete removes a payment.
func (r *PostgresPaymentRepository) Delete(ctx context.Context, id payment.PaymentID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM payments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete payment: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("payment not found: %s", id)
	}

	return nil
}

// loadAttempts retrieves all payment attempts for a payment.
func (r *PostgresPaymentRepository) loadAttempts(ctx context.Context, paymentID payment.PaymentID) ([]payment.PaymentAttempt, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT attempted_at, status, error_code, error_msg
		FROM payment_attempts WHERE payment_id = $1
		ORDER BY attempted_at ASC
	`, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query payment attempts: %w", err)
	}
	defer rows.Close()

	var attempts []payment.PaymentAttempt
	for rows.Next() {
		var attempt payment.PaymentAttempt
		var errorCode, errorMsg sql.NullString

		if err := rows.Scan(&attempt.AttemptedAt, &attempt.Status, &errorCode, &errorMsg); err != nil {
			return nil, fmt.Errorf("failed to scan payment attempt: %w", err)
		}

		if errorCode.Valid {
			attempt.ErrorCode = errorCode.String
		}
		if errorMsg.Valid {
			attempt.ErrorMsg = errorMsg.String
		}

		attempts = append(attempts, attempt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payment attempts: %w", err)
	}

	return attempts, nil
}
