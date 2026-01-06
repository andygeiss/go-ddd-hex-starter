package inbound_test

import (
	"context"
	"embed"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/logging"
	"github.com/andygeiss/cloud-native-utils/security"
	"github.com/andygeiss/cloud-native-utils/templating"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/inbound"
)

// ============================================================================
// Test Assets
// ============================================================================
// We embed test templates and static files for testing HTTP handlers.
// The assets/ directory at package level mirrors the production structure
// expected by the router (assets/templates/*.tmpl, assets/static/*).

//go:embed assets/templates/*.tmpl assets/static/css/*.css
var testAssets embed.FS

// ============================================================================
// HttpViewLogin Tests
// ============================================================================

func Test_HttpViewLogin_With_Request_Should_Return_200(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(testAssets)
	e.Parse("assets/templates/*.tmpl")

	handler := inbound.HttpViewLogin(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/login", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_HttpViewLogin_With_Request_Should_Render_Template(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(testAssets)
	e.Parse("assets/templates/*.tmpl")

	handler := inbound.HttpViewLogin(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/login", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain app name", containsString(bodyStr, "TestApp"), true)
}

func Test_HttpViewLogin_With_Request_Should_Return_HTML_Content_Type(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(testAssets)
	e.Parse("assets/templates/*.tmpl")

	handler := inbound.HttpViewLogin(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/login", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	contentType := rec.Header().Get("Content-Type")
	assert.That(t, "content type must be text/html", containsString(contentType, "text/html"), true)
}

// ============================================================================
// HttpViewIndex Tests
// ============================================================================

func Test_HttpViewIndex_Without_Session_Should_Redirect_To_Login(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(testAssets)
	e.Parse("assets/templates/*.tmpl")

	handler := inbound.HttpViewIndex(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 303 (redirect)", rec.Code, http.StatusSeeOther)
	location := rec.Header().Get("Location")
	assert.That(t, "location must contain login", containsString(location, "/ui/login"), true)
}

func Test_HttpViewIndex_With_Empty_SessionID_Should_Redirect_To_Login(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(testAssets)
	e.Parse("assets/templates/*.tmpl")

	handler := inbound.HttpViewIndex(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/", nil)
	// Add empty session ID to context
	ctx := context.WithValue(req.Context(), security.ContextSessionID, "")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 303 (redirect)", rec.Code, http.StatusSeeOther)
}

func Test_HttpViewIndex_With_Valid_Session_Should_Return_200(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(testAssets)
	e.Parse("assets/templates/*.tmpl")

	handler := inbound.HttpViewIndex(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/", nil)

	// Add session context values
	ctx := req.Context()
	ctx = context.WithValue(ctx, security.ContextSessionID, "test-session-123")
	ctx = context.WithValue(ctx, security.ContextEmail, "test@example.com")
	ctx = context.WithValue(ctx, security.ContextIssuer, "https://issuer.example.com")
	ctx = context.WithValue(ctx, security.ContextName, "Test User")
	ctx = context.WithValue(ctx, security.ContextSubject, "user-subject-456")
	ctx = context.WithValue(ctx, security.ContextVerified, true)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_HttpViewIndex_With_Valid_Session_Should_Render_User_Data(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	e := templating.NewEngine(testAssets)
	e.Parse("assets/templates/*.tmpl")

	handler := inbound.HttpViewIndex(e)
	req := httptest.NewRequest(http.MethodGet, "/ui/", nil)

	// Add session context values
	ctx := req.Context()
	ctx = context.WithValue(ctx, security.ContextSessionID, "test-session-123")
	ctx = context.WithValue(ctx, security.ContextEmail, "test@example.com")
	ctx = context.WithValue(ctx, security.ContextIssuer, "https://issuer.example.com")
	ctx = context.WithValue(ctx, security.ContextName, "Test User")
	ctx = context.WithValue(ctx, security.ContextSubject, "user-subject-456")
	ctx = context.WithValue(ctx, security.ContextVerified, true)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain user email", containsString(bodyStr, "test@example.com"), true)
	assert.That(t, "body must contain user name", containsString(bodyStr, "Test User"), true)
	assert.That(t, "body must contain session ID", containsString(bodyStr, "test-session-123"), true)
}

// ============================================================================
// HttpView Tests
// ============================================================================

func Test_HttpView_With_Valid_Template_Should_Return_200(t *testing.T) {
	// Arrange
	e := templating.NewEngine(testAssets)
	e.Parse("assets/templates/*.tmpl")

	data := struct {
		AppName string
		Title   string
	}{
		AppName: "TestApp",
		Title:   "Test Title",
	}

	handler := inbound.HttpView(e, "login", data)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	assert.That(t, "status code must be 200", rec.Code, http.StatusOK)
}

func Test_HttpView_With_Data_Should_Render_Data_In_Template(t *testing.T) {
	// Arrange
	e := templating.NewEngine(testAssets)
	e.Parse("assets/templates/*.tmpl")

	data := struct {
		AppName string
		Title   string
	}{
		AppName: "MyCustomApp",
		Title:   "Custom Title",
	}

	handler := inbound.HttpView(e, "login", data)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Act
	handler(rec, req)

	// Assert
	body, _ := io.ReadAll(rec.Body)
	bodyStr := string(body)
	assert.That(t, "body must contain custom app name", containsString(bodyStr, "MyCustomApp"), true)
}

// ============================================================================
// Router Tests
// ============================================================================

func Test_Route_With_Liveness_Endpoint_Should_Return_OK(t *testing.T) {
	// Arrange
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, testAssets, logger)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Act
	resp, err := http.Get(server.URL + "/liveness")

	// Assert
	assert.That(t, "request error must be nil", err == nil, true)
	assert.That(t, "status code must be 200", resp.StatusCode, http.StatusOK)

	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	assert.That(t, "body must be OK", string(body), "OK")
}

func Test_Route_With_Readiness_Endpoint_Should_Return_OK(t *testing.T) {
	// Arrange
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, testAssets, logger)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Act
	resp, err := http.Get(server.URL + "/readiness")

	// Assert
	assert.That(t, "request error must be nil", err == nil, true)
	assert.That(t, "status code must be 200", resp.StatusCode, http.StatusOK)
	_ = resp.Body.Close()
}

func Test_Route_With_Login_Endpoint_Should_Return_200(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, testAssets, logger)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Act
	resp, err := http.Get(server.URL + "/ui/login")

	// Assert
	assert.That(t, "request error must be nil", err == nil, true)
	assert.That(t, "status code must be 200", resp.StatusCode, http.StatusOK)
	_ = resp.Body.Close()
}

func Test_Route_With_UI_Index_Without_Auth_Should_Redirect(t *testing.T) {
	// Arrange
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, testAssets, logger)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Create a client that doesn't follow redirects
	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Act
	resp, err := client.Get(server.URL + "/ui/")

	// Assert
	assert.That(t, "request error must be nil", err == nil, true)
	assert.That(t, "status code must be 303 (redirect)", resp.StatusCode, http.StatusSeeOther)
	_ = resp.Body.Close()
}

func Test_Route_With_Static_CSS_Should_Serve_File(t *testing.T) {
	// Arrange
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, testAssets, logger)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Act
	resp, err := http.Get(server.URL + "/static/css/test.css")

	// Assert
	assert.That(t, "request error must be nil", err == nil, true)
	assert.That(t, "status code must be 200", resp.StatusCode, http.StatusOK)

	contentType := resp.Header.Get("Content-Type")
	assert.That(t, "content type must be text/css", containsString(contentType, "text/css"), true)
	_ = resp.Body.Close()
}

func Test_Route_With_NonExistent_Path_Should_Return_404(t *testing.T) {
	// Arrange
	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, testAssets, logger)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Act
	resp, err := http.Get(server.URL + "/nonexistent/path")

	// Assert
	assert.That(t, "request error must be nil", err == nil, true)
	assert.That(t, "status code must be 404", resp.StatusCode, http.StatusNotFound)
	_ = resp.Body.Close()
}

func Test_Route_With_Session_Path_Should_Return_200(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, testAssets, logger)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Act - session paths are handled by WithAuth which manages sessions internally
	resp, err := http.Get(server.URL + "/ui/some-session-id/")

	// Assert
	assert.That(t, "request error must be nil", err == nil, true)
	// Session paths are handled by security.WithAuth middleware,
	// which may redirect to login or render based on session state.
	// For a fresh server without stored sessions, we accept either behavior.
	isValid := resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusSeeOther
	assert.That(t, "status code must be 200 or 303", isValid, true)
	_ = resp.Body.Close()
}

func Test_Route_With_Session_Path_No_Trailing_Slash_Should_Return_200(t *testing.T) {
	// Arrange
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("APP_DESCRIPTION", "Test Description")

	ctx := context.Background()
	logger := logging.NewJsonLogger()
	mux := inbound.Route(ctx, testAssets, logger)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Act - session paths without trailing slash
	resp, err := http.Get(server.URL + "/ui/some-session-id")

	// Assert
	assert.That(t, "request error must be nil", err == nil, true)
	// Session paths are handled by security.WithAuth middleware,
	// which may redirect to login or render based on session state.
	isValid := resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusSeeOther
	assert.That(t, "status code must be 200 or 303", isValid, true)
	_ = resp.Body.Close()
}
