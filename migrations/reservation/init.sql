-- ======================================
-- Reservation Domain Schema
-- ======================================
-- Schema for the Reservation bounded context.
-- Uses key/value storage pattern matching PostgresAccess from cloud-native-utils.
-- This script runs automatically on first PostgreSQL startup.

CREATE TABLE IF NOT EXISTS kv_store (
    key TEXT PRIMARY KEY,
    value TEXT
);

CREATE INDEX IF NOT EXISTS idx_kv_store_key ON kv_store (key);
