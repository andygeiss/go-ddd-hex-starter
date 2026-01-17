package outbound

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// NewPostgresConnection creates a new PostgreSQL database connection pool.
// Configuration is read from environment variables:
//   - POSTGRES_HOST: Database host (default: localhost)
//   - POSTGRES_PORT: Database port (default: 5432)
//   - POSTGRES_USER: Database user (default: booking)
//   - POSTGRES_PASSWORD: Database password (default: booking_secret)
//   - POSTGRES_DB: Database name (default: booking_db)
//   - POSTGRES_SSLMODE: SSL mode (default: disable)
//   - POSTGRES_MAX_OPEN_CONNS: Max open connections (default: 25)
//   - POSTGRES_MAX_IDLE_CONNS: Max idle connections (default: 5)
//   - POSTGRES_CONN_MAX_LIFETIME: Connection max lifetime (default: 5m)
func NewPostgresConnection() (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("POSTGRES_HOST", "localhost"),
		getEnv("POSTGRES_PORT", "5432"),
		getEnv("POSTGRES_USER", "booking"),
		getEnv("POSTGRES_PASSWORD", "booking_secret"),
		getEnv("POSTGRES_DB", "booking_db"),
		getEnv("POSTGRES_SSLMODE", "disable"),
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	maxOpenConns, _ := strconv.Atoi(getEnv("POSTGRES_MAX_OPEN_CONNS", "25"))
	maxIdleConns, _ := strconv.Atoi(getEnv("POSTGRES_MAX_IDLE_CONNS", "5"))
	connMaxLifetime, _ := time.ParseDuration(getEnv("POSTGRES_CONN_MAX_LIFETIME", "5m"))

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
