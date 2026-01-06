package inbound_test

import (
	"embed"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/templating"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/inbound"
)

// ============================================================================
// Test Assets
// ============================================================================

//go:embed testdata/assets/templates/*.tmpl testdata/assets/static/css/*.css
var serviceWorkerTestAssets embed.FS

// ============================================================================
// HttpViewServiceWorker Tests
// ============================================================================

func Test_HttpViewServiceWorker_With_Request_Should_Return_200(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "test-app")
	t.Setenv("APP_VERSION", "1.0.0")
	e := templating.NewEngine(serviceWorkerTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewServiceWorker(e)
	req := httptest.NewRequest(http.MethodGet, "/sw.js", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_HttpViewServiceWorker_With_Request_Should_Return_JavaScript_Content_Type(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "test-app")
	t.Setenv("APP_VERSION", "1.0.0")
	e := templating.NewEngine(serviceWorkerTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewServiceWorker(e)
	req := httptest.NewRequest(http.MethodGet, "/sw.js", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	contentType := rec.Header().Get("Content-Type")
	assert.That(t, "content type must be application/javascript", contentType, "application/javascript")
}

func Test_HttpViewServiceWorker_With_Request_Should_Have_No_Cache_Header(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "test-app")
	t.Setenv("APP_VERSION", "1.0.0")
	e := templating.NewEngine(serviceWorkerTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewServiceWorker(e)
	req := httptest.NewRequest(http.MethodGet, "/sw.js", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	cacheControl := rec.Header().Get("Cache-Control")
	assert.That(t, "cache control must prevent caching", cacheControl, "no-cache, no-store, must-revalidate")
}

func Test_HttpViewServiceWorker_With_Request_Should_Contain_Cache_Name(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "my-pwa")
	t.Setenv("APP_VERSION", "2.0.0")
	e := templating.NewEngine(serviceWorkerTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewServiceWorker(e)
	req := httptest.NewRequest(http.MethodGet, "/sw.js", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain cache name", containsString(bodyStr, "my-pwa-v2.0.0"), true)
}

func Test_HttpViewServiceWorker_With_Request_Should_Contain_Install_Handler(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "test-app")
	t.Setenv("APP_VERSION", "1.0.0")
	e := templating.NewEngine(serviceWorkerTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewServiceWorker(e)
	req := httptest.NewRequest(http.MethodGet, "/sw.js", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain install event listener", containsString(bodyStr, "addEventListener('install'"), true)
}

func Test_HttpViewServiceWorker_With_Request_Should_Contain_Fetch_Handler(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "test-app")
	t.Setenv("APP_VERSION", "1.0.0")
	e := templating.NewEngine(serviceWorkerTestAssets)
	e.Parse("testdata/assets/templates/*.tmpl")

	handler := inbound.HttpViewServiceWorker(e)
	req := httptest.NewRequest(http.MethodGet, "/sw.js", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain fetch event listener", containsString(bodyStr, "addEventListener('fetch'"), true)
}
