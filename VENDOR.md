# Vendor Documentation

This document describes approved third-party libraries and their recommended usage patterns within this template. Agents and developers should consult this document before implementing custom utilities that might duplicate existing vendor functionality.

---

## Table of Contents

- [cloud-native-utils](#cloud-native-utils)
  - [Purpose](#purpose)
  - [Package Reference](#package-reference)
  - [Usage in This Template](#usage-in-this-template)
  - [Recommended Patterns](#recommended-patterns)
  - [When NOT to Use](#when-not-to-use)
- [htmx](#htmx)
  - [Purpose](#purpose-1)
  - [Core Attributes](#core-attributes)
  - [Usage in This Template](#usage-in-this-template-1)
  - [Recommended Patterns](#recommended-patterns-1)
  - [When NOT to Use](#when-not-to-use-1)

---

## cloud-native-utils

| Property | Value |
|----------|-------|
| Repository | [github.com/andygeiss/cloud-native-utils](https://github.com/andygeiss/cloud-native-utils) |
| Documentation | [pkg.go.dev/github.com/andygeiss/cloud-native-utils](https://pkg.go.dev/github.com/andygeiss/cloud-native-utils) |
| Version | v0.4.8 (as of 2026-01-05) |
| License | MIT |

### Purpose

A collection of modular, high-performance utilities for building cloud-native Go applications. Each package is designed to be used independently—no monolithic framework. Use this library first before implementing custom solutions for:

- Testing assertions
- Data persistence and CRUD operations
- Concurrent processing
- Security (encryption, auth, HTTP server)
- Messaging / pub-sub
- Stability patterns (retries, circuit breakers)
- HTML templating

---

### Package Reference

| Package | Purpose | Key Types/Functions |
|---------|---------|---------------------|
| `assert` | Minimal test assertions | `assert.That(t, desc, got, want)` |
| `consistency` | Transactional event log with JSON file persistence | `JsonFileLogger[K, V]` |
| `efficiency` | Channel helpers for concurrent processing | `Generate`, `Merge`, `Split`, `Process` |
| `extensibility` | Dynamic plugin loading | `LoadPlugin[T]` |
| `i18n` | Internationalization with embedded YAML | `Translations`, `FormatDateISO`, `FormatMoney` |
| `imaging` | QR code generation | `GenerateQRCodeDataURL` |
| `logging` | Structured JSON logging + HTTP middleware | `NewJsonLogger`, `WithLogging` |
| `messaging` | Publish-subscribe dispatcher (in-memory or Kafka) | `Dispatcher`, `NewInternalDispatcher`, `NewExternalDispatcher` |
| `redirecting` | HTMX-compatible PRG redirects | `WithPRG`, `Redirect`, `RedirectWithMessage` |
| `resource` | Generic CRUD access (memory, JSON, YAML, SQLite) | `Access[K, V]`, `JsonFileAccess`, `InMemoryAccess`, `SqliteAccess`, `IndexedAccess` |
| `scheduling` | Time/slot primitives for booking systems | `TimeOfDay`, `DayHours`, `MustTimeOfDay` |
| `security` | Encryption, hashing, OIDC, secure HTTP server | `Encrypt`, `Decrypt`, `Password`, `NewServer`, `NewServeMux`, `WithAuth`, `IdentityProvider` |
| `service` | Context-aware function wrapper, signal handling | `Context`, `Wrap`, `RegisterOnContextDone` |
| `slices` | Generic slice helpers | `Map`, `Filter`, `Unique`, `Contains` |
| `stability` | Resilience wrappers for `service.Function` | `Breaker`, `Retry`, `Throttle`, `Debounce`, `Timeout` |
| `templating` | HTML templating on embedded filesystems | `Engine`, `Parse`, `Render`, `View` |

---

### Usage in This Template

The following table shows where each package is used and the recommended integration layer:

| Package | Template Layer | Example Usage |
|---------|----------------|---------------|
| `assert` | Tests (`*_test.go`) | Test assertions in [file_reader_test.go](internal/adapters/inbound/file_reader_test.go) |
| `logging` | Inbound adapters | Request logging via `logging.WithLogging` in [router.go](internal/adapters/inbound/router.go) |
| `messaging` | Adapters (inbound & outbound) | Event pub/sub in [event_publisher.go](internal/adapters/outbound/event_publisher.go), [event_subscriber.go](internal/adapters/inbound/event_subscriber.go) |
| `redirecting` | Inbound adapters | PRG pattern in [http_index.go](internal/adapters/inbound/http_index.go) |
| `resource` | Outbound adapters | Repository implementation via `JsonFileAccess` in [file_index_repository.go](internal/adapters/outbound/file_index_repository.go) |
| `security` | Inbound adapters | Secure mux, auth middleware in [router.go](internal/adapters/inbound/router.go) |
| `service` | Adapters & cmd | Function wrapping for messaging handlers in [event_subscriber.go](internal/adapters/inbound/event_subscriber.go) |
| `templating` | Inbound adapters | HTML rendering in [http_view.go](internal/adapters/inbound/http_view.go), [http_login.go](internal/adapters/inbound/http_login.go), [router.go](internal/adapters/inbound/router.go) |

---

### Recommended Patterns

#### Testing with `assert`

Use `assert.That` for all test assertions. Follow the Arrange-Act-Assert pattern.

```go
import "github.com/andygeiss/cloud-native-utils/assert"

func Test_Example_Should_Pass(t *testing.T) {
    // Arrange
    expected := 42

    // Act
    result := ComputeValue()

    // Assert
    assert.That(t, "result should be 42", result, expected)
}
```

#### Repository with `resource`

Use `resource.JsonFileAccess` (or other backends) instead of custom repository implementations when a key-value shape fits.

```go
import "github.com/andygeiss/cloud-native-utils/resource"

// Embed the generic access type
type MyRepository struct {
    resource.JsonFileAccess[MyID, MyEntity]
}

func NewMyRepository(filename string) MyPort {
    return resource.NewJsonFileAccess[MyID, MyEntity](filename)
}
```

#### Secure HTTP Server with `security`

Use `security.NewServeMux` for the base mux with built-in liveness/readiness probes and session support. Use `security.WithAuth` for authentication middleware.

```go
import "github.com/andygeiss/cloud-native-utils/security"

mux, sessions := security.NewServeMux(ctx, efs)
mux.HandleFunc("GET /protected", security.WithAuth(sessions, handler))
server := security.NewServer(mux)
```

#### Event Messaging with `messaging`

Use `messaging.Dispatcher` for pub/sub patterns. Prefer `NewInternalDispatcher` for in-process communication; use `NewExternalDispatcher` with Kafka for distributed systems.

```go
import "github.com/andygeiss/cloud-native-utils/messaging"

dispatcher := messaging.NewInternalDispatcher()
_ = dispatcher.Subscribe(ctx, "topic", handlerFunc)
_ = dispatcher.Publish(ctx, messaging.NewMessage("topic", payload))
```

#### Stability Wrappers

Wrap external service calls with stability patterns instead of implementing custom retry/circuit-breaker logic.

```go
import "github.com/andygeiss/cloud-native-utils/stability"

// Circuit breaker: opens after 3 failures
fn := stability.Breaker(externalCall, 3)

// Retry: up to 5 attempts with 1s delay
fn := stability.Retry(externalCall, 5, time.Second)

// Combine patterns as needed
fn := stability.Timeout(stability.Retry(externalCall, 3, time.Second), 10*time.Second)
```

#### Templating with `templating`

Use the templating engine with embedded filesystems. Parse templates once at startup, reuse the engine for all views.

```go
import "github.com/andygeiss/cloud-native-utils/templating"

//go:embed assets/templates/*.tmpl
var templatesFS embed.FS

engine := templating.NewEngine(templatesFS)
engine.Parse("assets/templates/*.tmpl")

// Use as HTTP handler
mux.HandleFunc("GET /page", engine.View("page.tmpl", data))
```

#### Logging with `logging`

Use `logging.WithLogging` to wrap HTTP handlers for structured request logs.

```go
import "github.com/andygeiss/cloud-native-utils/logging"

logger := logging.NewJsonLogger()
mux.HandleFunc("GET /endpoint", logging.WithLogging(logger, handler))
```

---

### When NOT to Use

| Scenario | Guidance |
|----------|----------|
| Domain logic | Never import vendor utilities into `internal/domain/*`. Domain should be pure Go with no external dependencies. |
| Complex ORM needs | `resource` is for simple key-value CRUD. For complex relational queries, consider a dedicated ORM or query builder. |
| High-throughput streaming | `efficiency` helpers are suitable for moderate workloads. For extreme throughput, evaluate specialized libraries. |
| Platform-specific plugins | `extensibility.LoadPlugin` only works on supported OS/architectures. |

---

## htmx

| Property | Value |
|----------|-------|
| Website | [htmx.org](https://htmx.org) |
| Documentation | [htmx.org/docs](https://htmx.org/docs/) |
| Reference | [htmx.org/reference](https://htmx.org/reference/) |
| Repository | [github.com/bigskysoftware/htmx](https://github.com/bigskysoftware/htmx) |
| Version | 2.0.8 (embedded in static assets) |
| Size | ~16kb min+gzip |
| License | BSD 2-Clause |

### Purpose

htmx extends HTML with attributes that enable AJAX requests, CSS transitions, WebSockets, and Server-Sent Events directly in markup—no custom JavaScript required. It enables hypermedia-driven applications where the server returns HTML fragments instead of JSON, keeping UI logic server-side.

Use htmx instead of:
- Writing custom `fetch()` calls for dynamic updates
- Building SPA-style JavaScript frameworks
- Complex client-side state management

---

### Core Attributes

| Attribute | Purpose | Example |
|-----------|---------|--------|
| `hx-get` | Issue GET request | `<div hx-get="/items">` |
| `hx-post` | Issue POST request | `<button hx-post="/submit">` |
| `hx-put` | Issue PUT request | `<button hx-put="/update">` |
| `hx-patch` | Issue PATCH request | `<button hx-patch="/partial">` |
| `hx-delete` | Issue DELETE request | `<button hx-delete="/remove">` |
| `hx-trigger` | Event that triggers request | `hx-trigger="click"`, `hx-trigger="load"`, `hx-trigger="revealed"` |
| `hx-target` | Element to update with response | `hx-target="#result"`, `hx-target="closest tr"` |
| `hx-swap` | How to swap content | `innerHTML`, `outerHTML`, `beforeend`, `afterbegin`, `delete` |
| `hx-boost` | Progressive enhancement for links/forms | `<body hx-boost="true">` |
| `hx-indicator` | Show during request | `hx-indicator="#spinner"` |
| `hx-confirm` | Confirmation dialog | `hx-confirm="Are you sure?"` |
| `hx-vals` | Include extra values | `hx-vals='{"key": "value"}'` |
| `hx-headers` | Add request headers | `hx-headers='{"X-Custom": "value"}'` |

---

### Usage in This Template

| Location | Purpose |
|----------|--------|
| [cmd/server/assets/static/js/htmx.min.js](cmd/server/assets/static/js/htmx.min.js) | Embedded htmx library |
| [cmd/server/assets/templates/*.tmpl](cmd/server/assets/templates/) | Templates using htmx attributes |
| [internal/adapters/inbound/http_*.go](internal/adapters/inbound/) | Handlers returning HTML fragments |

**Integration with cloud-native-utils:**

- `templating.Engine` renders HTML partials that htmx swaps into the DOM
- `redirecting.WithPRG` handles Post-Redirect-Get compatible with htmx
- `security.WithAuth` protects htmx endpoints the same as regular routes

---

### Recommended Patterns

#### Basic AJAX Update

```html
<!-- Button that loads content into a target div -->
<button hx-get="/api/data" hx-target="#result" hx-swap="innerHTML">
    Load Data
</button>
<div id="result"></div>
```

#### Form Submission

```html
<!-- Form that submits via AJAX and replaces itself -->
<form hx-post="/submit" hx-swap="outerHTML">
    <input name="email" type="email" required>
    <button type="submit">Subscribe</button>
</form>
```

#### Progressive Enhancement

```html
<!-- Enable htmx for all links and forms in body -->
<body hx-boost="true">
    <a href="/page">This link uses AJAX</a>
</body>
```

#### Loading Indicator

```html
<button hx-get="/slow" hx-indicator="#spinner">Load</button>
<span id="spinner" class="htmx-indicator">Loading...</span>
```

#### Server Handler Pattern

Handlers return HTML fragments, not JSON:

```go
// Return an HTML partial for htmx to swap
func HandlePartial(e *templating.Engine) http.HandlerFunc {
    return e.View("partial.tmpl", data)
}
```

---

### When NOT to Use

| Scenario | Guidance |
|----------|----------|
| Complex client-side state | For rich interactive apps with complex state, consider a JS framework |
| Real-time collaboration | WebSocket-heavy apps may need more control than htmx provides |
| Offline-first apps | htmx requires server connectivity; use service workers/PWA instead |
| Third-party API integration | htmx is for your own server; external APIs need JavaScript |

---

## Adding New Vendors

When introducing a new vendor library:

1. Justify the addition—does it solve a problem not covered by existing vendors?
2. Add a section to this document following the same structure:
   - Purpose statement
   - Package reference table
   - Usage in template (layer mapping)
   - Recommended patterns
   - When not to use
3. Ensure the new vendor respects layering rules from `CONTEXT.md`.

---

## Deprecating Vendors

When replacing or removing a vendor:

1. Mark the section as **DEPRECATED** with a note and migration guidance.
2. Update all references in the codebase.
3. Remove the deprecated section after migration is complete.
