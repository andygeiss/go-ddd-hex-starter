package outbound

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/andygeiss/hotel-booking/internal/domain/reservation"
)

// PostgresReservationRepository implements the ReservationRepository port using PostgreSQL.
type PostgresReservationRepository struct {
	db *sql.DB
}

// NewPostgresReservationRepository creates a new PostgreSQL-based reservation repository.
func NewPostgresReservationRepository(db *sql.DB) *PostgresReservationRepository {
	return &PostgresReservationRepository{db: db}
}

// Create inserts a new reservation into the database.
func (r *PostgresReservationRepository) Create(ctx context.Context, id reservation.ReservationID, value reservation.Reservation) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert reservation
	_, err = tx.ExecContext(ctx, `
		INSERT INTO reservations (
			id, guest_id, room_id, check_in, check_out, status,
			total_amount_cents, total_amount_currency, cancellation_reason,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		value.ID, value.GuestID, value.RoomID,
		value.DateRange.CheckIn, value.DateRange.CheckOut, value.Status,
		value.TotalAmount.Amount, value.TotalAmount.Currency, value.CancellationReason,
		value.CreatedAt, value.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert reservation: %w", err)
	}

	// Insert guests
	for _, guest := range value.Guests {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO reservation_guests (reservation_id, name, email, phone_number)
			VALUES ($1, $2, $3, $4)
		`, value.ID, guest.Name, guest.Email, guest.PhoneNumber)
		if err != nil {
			return fmt.Errorf("failed to insert guest: %w", err)
		}
	}

	return tx.Commit()
}

// Read retrieves a reservation by ID.
func (r *PostgresReservationRepository) Read(ctx context.Context, id reservation.ReservationID) (*reservation.Reservation, error) {
	var res reservation.Reservation
	var cancellationReason sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT id, guest_id, room_id, check_in, check_out, status,
			total_amount_cents, total_amount_currency, cancellation_reason,
			created_at, updated_at
		FROM reservations WHERE id = $1
	`, id).Scan(
		&res.ID, &res.GuestID, &res.RoomID,
		&res.DateRange.CheckIn, &res.DateRange.CheckOut, &res.Status,
		&res.TotalAmount.Amount, &res.TotalAmount.Currency, &cancellationReason,
		&res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("reservation not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read reservation: %w", err)
	}

	if cancellationReason.Valid {
		res.CancellationReason = cancellationReason.String
	}

	// Load guests
	guests, err := r.loadGuests(ctx, id)
	if err != nil {
		return nil, err
	}
	res.Guests = guests

	return &res, nil
}

// ReadAll retrieves all reservations.
func (r *PostgresReservationRepository) ReadAll(ctx context.Context) ([]reservation.Reservation, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, guest_id, room_id, check_in, check_out, status,
			total_amount_cents, total_amount_currency, cancellation_reason,
			created_at, updated_at
		FROM reservations
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query reservations: %w", err)
	}
	defer rows.Close()

	var reservations []reservation.Reservation
	for rows.Next() {
		var res reservation.Reservation
		var cancellationReason sql.NullString

		if err := rows.Scan(
			&res.ID, &res.GuestID, &res.RoomID,
			&res.DateRange.CheckIn, &res.DateRange.CheckOut, &res.Status,
			&res.TotalAmount.Amount, &res.TotalAmount.Currency, &cancellationReason,
			&res.CreatedAt, &res.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}

		if cancellationReason.Valid {
			res.CancellationReason = cancellationReason.String
		}

		reservations = append(reservations, res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reservations: %w", err)
	}

	// Load guests for each reservation
	for i := range reservations {
		guests, err := r.loadGuests(ctx, reservations[i].ID)
		if err != nil {
			return nil, err
		}
		reservations[i].Guests = guests
	}

	return reservations, nil
}

// Update modifies an existing reservation.
func (r *PostgresReservationRepository) Update(ctx context.Context, id reservation.ReservationID, value reservation.Reservation) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		UPDATE reservations SET
			guest_id = $2, room_id = $3, check_in = $4, check_out = $5,
			status = $6, total_amount_cents = $7, total_amount_currency = $8,
			cancellation_reason = $9, updated_at = $10
		WHERE id = $1
	`,
		id, value.GuestID, value.RoomID,
		value.DateRange.CheckIn, value.DateRange.CheckOut, value.Status,
		value.TotalAmount.Amount, value.TotalAmount.Currency, value.CancellationReason,
		value.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update reservation: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("reservation not found: %s", id)
	}

	// Delete and re-insert guests
	_, err = tx.ExecContext(ctx, `DELETE FROM reservation_guests WHERE reservation_id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete guests: %w", err)
	}

	for _, guest := range value.Guests {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO reservation_guests (reservation_id, name, email, phone_number)
			VALUES ($1, $2, $3, $4)
		`, id, guest.Name, guest.Email, guest.PhoneNumber)
		if err != nil {
			return fmt.Errorf("failed to insert guest: %w", err)
		}
	}

	return tx.Commit()
}

// Delete removes a reservation.
func (r *PostgresReservationRepository) Delete(ctx context.Context, id reservation.ReservationID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM reservations WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete reservation: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("reservation not found: %s", id)
	}

	return nil
}

// loadGuests retrieves all guests for a reservation.
func (r *PostgresReservationRepository) loadGuests(ctx context.Context, reservationID reservation.ReservationID) ([]reservation.GuestInfo, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT name, email, phone_number
		FROM reservation_guests WHERE reservation_id = $1
	`, reservationID)
	if err != nil {
		return nil, fmt.Errorf("failed to query guests: %w", err)
	}
	defer rows.Close()

	var guests []reservation.GuestInfo
	for rows.Next() {
		var guest reservation.GuestInfo
		var phoneNumber sql.NullString

		if err := rows.Scan(&guest.Name, &guest.Email, &phoneNumber); err != nil {
			return nil, fmt.Errorf("failed to scan guest: %w", err)
		}

		if phoneNumber.Valid {
			guest.PhoneNumber = phoneNumber.String
		}

		guests = append(guests, guest)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating guests: %w", err)
	}

	return guests, nil
}
