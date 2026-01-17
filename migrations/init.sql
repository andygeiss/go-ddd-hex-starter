-- ======================================
-- Booking Domain Schema
-- ======================================
-- Initial database schema for the booking system.
-- This script runs automatically on first PostgreSQL startup.

-- ======================================
-- Reservations Table
-- ======================================
-- Stores the Reservation aggregate root.
-- Status transitions: pending -> confirmed -> active -> completed (or cancelled)
CREATE TABLE IF NOT EXISTS reservations (
    id VARCHAR(255) PRIMARY KEY,
    guest_id VARCHAR(255) NOT NULL,
    room_id VARCHAR(255) NOT NULL,
    check_in TIMESTAMP WITH TIME ZONE NOT NULL,
    check_out TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    total_amount_cents BIGINT NOT NULL,
    total_amount_currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    cancellation_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_reservation_status CHECK (status IN ('pending', 'confirmed', 'active', 'completed', 'cancelled')),
    CONSTRAINT valid_date_range CHECK (check_out > check_in)
);

-- ======================================
-- Reservation Guests Table
-- ======================================
-- Stores GuestInfo entities within the Reservation aggregate.
-- One reservation can have multiple guests.
CREATE TABLE IF NOT EXISTS reservation_guests (
    id SERIAL PRIMARY KEY,
    reservation_id VARCHAR(255) NOT NULL REFERENCES reservations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone_number VARCHAR(50)
);

-- ======================================
-- Payments Table
-- ======================================
-- Stores the Payment aggregate root.
-- Status transitions: pending -> authorized -> captured (or failed/refunded)
CREATE TABLE IF NOT EXISTS payments (
    id VARCHAR(255) PRIMARY KEY,
    reservation_id VARCHAR(255) NOT NULL,
    amount_cents BIGINT NOT NULL,
    amount_currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    payment_method VARCHAR(100),
    transaction_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_payment_status CHECK (status IN ('pending', 'authorized', 'captured', 'failed', 'refunded'))
);

-- ======================================
-- Payment Attempts Table
-- ======================================
-- Stores PaymentAttempt entities within the Payment aggregate.
-- Tracks history of payment processing attempts.
CREATE TABLE IF NOT EXISTS payment_attempts (
    id SERIAL PRIMARY KEY,
    payment_id VARCHAR(255) NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    attempted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    status VARCHAR(50) NOT NULL,
    error_code VARCHAR(100),
    error_msg TEXT
);

-- ======================================
-- Indexes for Common Queries
-- ======================================

-- Reservation lookups by guest (for listing user's reservations)
CREATE INDEX IF NOT EXISTS idx_reservations_guest_id ON reservations(guest_id);

-- Reservation lookups by room (for availability checking)
CREATE INDEX IF NOT EXISTS idx_reservations_room_id ON reservations(room_id);

-- Filter reservations by status
CREATE INDEX IF NOT EXISTS idx_reservations_status ON reservations(status);

-- Date range queries for availability checking
CREATE INDEX IF NOT EXISTS idx_reservations_check_in ON reservations(check_in);
CREATE INDEX IF NOT EXISTS idx_reservations_check_out ON reservations(check_out);

-- Composite index for availability queries (room + date range)
CREATE INDEX IF NOT EXISTS idx_reservations_room_dates ON reservations(room_id, check_in, check_out);

-- Payment lookups by reservation
CREATE INDEX IF NOT EXISTS idx_payments_reservation_id ON payments(reservation_id);

-- Filter payments by status
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);

-- Guest lookups by reservation
CREATE INDEX IF NOT EXISTS idx_reservation_guests_reservation_id ON reservation_guests(reservation_id);

-- Payment attempts by payment
CREATE INDEX IF NOT EXISTS idx_payment_attempts_payment_id ON payment_attempts(payment_id);
