package inbound_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/inbound"

	"github.com/andygeiss/cloud-native-utils/assert"
)

// Every test should follow the Test_<struct>_<method>_With_<condition>_Should_<result> pattern.
// This is important because we want the tests to be easy to read and understand.
// We use the Arrange-Act-Assert pattern for better readability.
// We use the assert package from the cloud-native-utils library for better readability.

func Test_WithLogging_With_Request_Should_Log_Request(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	handlerCalled := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}
	wrappedHandler := inbound.WithLogging(logger, handler)
	req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
	rec := httptest.NewRecorder()

	// Act
	wrappedHandler(rec, req)

	// Assert
	assert.That(t, "handler must be called", handlerCalled, true)
	logOutput := buf.String()
	assert.That(t, "log must contain method", containsString(logOutput, "GET"), true)
	assert.That(t, "log must contain path", containsString(logOutput, "/test/path"), true)
	assert.That(t, "log must contain duration", containsString(logOutput, "duration"), true)
}

func Test_WithLogging_With_POST_Request_Should_Log_Method(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}
	wrappedHandler := inbound.WithLogging(logger, handler)
	req := httptest.NewRequest(http.MethodPost, "/api/resource", nil)
	rec := httptest.NewRecorder()

	// Act
	wrappedHandler(rec, req)

	// Assert
	logOutput := buf.String()
	assert.That(t, "log must contain POST method", containsString(logOutput, "POST"), true)
}

func Test_WithSecurityHeaders_With_Request_Should_Add_Headers(t *testing.T) {
	// Arrange
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	wrappedHandler := inbound.WithSecurityHeaders(handler)
	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	rec := httptest.NewRecorder()

	// Act
	wrappedHandler(rec, req)

	// Assert
	assert.That(t, "Referrer-Policy must be set", rec.Header().Get("Referrer-Policy"), "strict-origin-when-cross-origin")
	assert.That(t, "Cache-Control must be set", rec.Header().Get("Cache-Control"), "no-store")
	assert.That(t, "X-Content-Type-Options must be set", rec.Header().Get("X-Content-Type-Options"), "nosniff")
	assert.That(t, "X-Frame-Options must be set", rec.Header().Get("X-Frame-Options"), "DENY")
}

func Test_WithSecurityHeaders_With_Request_Should_Call_Next_Handler(t *testing.T) {
	// Arrange
	handlerCalled := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}
	wrappedHandler := inbound.WithSecurityHeaders(handler)
	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	rec := httptest.NewRecorder()

	// Act
	wrappedHandler(rec, req)

	// Assert
	assert.That(t, "handler must be called", handlerCalled, true)
}

func Test_WithSecurityHeaders_With_Request_Should_Preserve_Status_Code(t *testing.T) {
	// Arrange
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}
	wrappedHandler := inbound.WithSecurityHeaders(handler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Act
	wrappedHandler(rec, req)

	// Assert
	assert.That(t, "status code must be preserved", rec.Code, http.StatusAccepted)
}

// containsString is a helper function to check if a string contains a substring.
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
