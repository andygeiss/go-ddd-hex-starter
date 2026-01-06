# VENDOR.md

## Overview

This document describes the external vendor libraries used in this project and provides guidance on when and how to use them. The project primarily depends on `cloud-native-utils` as its core utility library, with transitive dependencies for OIDC authentication and Kafka messaging.

**Philosophy:** Prefer reusing vendor functionality over re-implementing similar utilities. This keeps the codebase lean and benefits from battle-tested, maintained libraries.

---

## Approved Vendor Libraries

### cloud-native-utils

- **Purpose:** Core utility library providing modular, cloud-native building blocks for Go applications. This is the primary vendor dependency and should be the first choice for cross-cutting concerns.
- **Repository:** [github.com/andygeiss/cloud-native-utils](https://github.com/andygeiss/cloud-native-utils)
- **Version:** v0.4.10
- **Documentation:** [pkg.go.dev](https://pkg.go.dev/github.com/andygeiss/cloud-native-utils)

#### Key Packages

| Package | Purpose | Used In This Project |
|---------|---------|---------------------|
| `assert` | Test assertions (`assert.That`) | `internal/domain/*_test.go`, `internal/adapters/*_test.go`, `cmd/*_test.go` |
| `logging` | Structured JSON logging | `cmd/server/main.go` via `logging.NewJsonLogger()` |
| `messaging` | Pub/sub dispatcher (in-memory or Kafka) | `cmd/cli/main.go`, `internal/adapters/outbound/event_publisher.go`, tests |
| `redirecting` | HTMX-compatible redirects | `internal/adapters/inbound/http_index.go` |
| `resource` | Generic CRUD interface with backends | `internal/adapters/outbound/file_index_repository.go`, tests |
| `security` | OIDC, encryption, HTTP server | `cmd/server/main.go`, `internal/adapters/inbound/router.go` |
| `service` | Context helpers, lifecycle hooks | `cmd/server/main.go`, `cmd/cli/main.go` |
| `templating` | HTML template engine with `fs.FS` | `internal/adapters/inbound/router.go`, `http_view.go` |

#### When to Use

| Concern | Package | Pattern |
|---------|---------|---------|
| **Testing assertions** | `assert` | `assert.That(t, "description", actual, expected)` |
| **Structured logging** | `logging` | `logging.NewJsonLogger()` at startup, pass to handlers |
| **HTTP request logging** | `logging` | `logging.WithLogging(logger, handler)` middleware |
| **Event publishing/subscribing** | `messaging` | `messaging.NewExternalDispatcher()` for Kafka |
| **In-memory messaging (tests)** | `messaging` | `messaging.NewInternalDispatcher()` for unit tests |
| **HTTP redirects** | `redirecting` | `redirecting.Redirect(w, r, "/path")` |
| **CRUD repositories** | `resource` | `resource.NewJsonFileAccess[K, V](filename)` |
| **Mock repositories** | `resource` | `resource.NewMockAccess[K, V]()` for tests |
| **HTTP server setup** | `security` | `security.NewServer(mux)` with env-based config |
| **OIDC authentication** | `security` | `security.NewServeMux(ctx, efs)` with session management (accepts `fs.FS`) |
| **Auth middleware** | `security` | `security.WithAuth(sessions, handler)` |
| **Context with signals** | `service` | `service.Context()` for graceful shutdown |
| **Shutdown hooks** | `service` | `service.RegisterOnContextDone(ctx, fn)` |
| **Function wrapping** | `service` | `service.Wrap(fn)` for context-aware functions |
| **HTML templating** | `templating` | `templating.NewEngine(efs)` with `fs.FS` (use `fs.Sub()` for path remapping) |

#### Integration Patterns

**Logging Setup:**
```go
import "github.com/andygeiss/cloud-native-utils/logging"

logger := logging.NewJsonLogger()
mux.HandleFunc("GET /path", logging.WithLogging(logger, handler))
```

**Messaging Setup:**
```go
import "github.com/andygeiss/cloud-native-utils/messaging"

// For Kafka (requires KAFKA_BROKERS env var)
dispatcher := messaging.NewExternalDispatcher()

// For in-memory (testing/development)
dispatcher := messaging.NewInternalDispatcher()
```

**Repository Setup:**
```go
import "github.com/andygeiss/cloud-native-utils/resource"

// JSON file persistence
repo := resource.NewJsonFileAccess[KeyType, ValueType](filename)

// In-memory (testing)
repo := resource.NewInMemoryAccess[KeyType, ValueType]()
```

**HTTP Server Setup:**
```go
import "github.com/andygeiss/cloud-native-utils/security"

// efs can be embed.FS or any fs.FS implementation
mux, sessions := security.NewServeMux(ctx, efs)
srv := security.NewServer(mux)
```

#### Cautions

- **Environment variables:** Many packages read configuration from environment variables at initialization. Ensure variables are set before creating instances.
- **OIDC configuration:** Requires `OIDC_ISSUER`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, `OIDC_REDIRECT_URL` environment variables.
- **Kafka configuration:** Requires `KAFKA_BROKERS` and optionally `KAFKA_CONSUMER_GROUP_ID`.
- **Server timeouts:** Configured via `SERVER_*_TIMEOUT` environment variables.
- **fs.FS flexibility:** As of v0.4.10, `security.NewServeMux()` and `templating.NewEngine()` accept `fs.FS` instead of `embed.FS`. This enables using `fs.Sub()` to remap paths in tests (e.g., `fs.Sub(testAssets, "testdata")` to align embedded test assets with production paths).

---

### coreos/go-oidc (Transitive)

- **Purpose:** OpenID Connect client library for Go
- **Repository:** [github.com/coreos/go-oidc/v3](https://github.com/coreos/go-oidc)
- **Version:** v3.17.0
- **Status:** Transitive dependency via `cloud-native-utils/security`

#### When to Use

**Do not use directly.** Use `cloud-native-utils/security` which wraps this library with:
- Automatic provider discovery
- Session management integration
- Environment-based configuration

The only case to use directly is if you need advanced OIDC features not exposed by cloud-native-utils.

---

### segmentio/kafka-go (Transitive)

- **Purpose:** Kafka client library for Go
- **Repository:** [github.com/segmentio/kafka-go](https://github.com/segmentio/kafka-go)
- **Version:** v0.4.49
- **Status:** Transitive dependency via `cloud-native-utils/messaging`

#### When to Use

**Do not use directly.** Use `cloud-native-utils/messaging` which provides:
- Unified `Dispatcher` interface
- In-memory and Kafka backends with the same API
- Simplified publish/subscribe patterns

The only case to use directly is for advanced Kafka features (consumer groups, partitioning, transactions) not exposed by the dispatcher abstraction.

---

### golang.org/x/oauth2 (Transitive)

- **Purpose:** OAuth 2.0 client library
- **Repository:** [golang.org/x/oauth2](https://pkg.go.dev/golang.org/x/oauth2)
- **Version:** v0.34.0
- **Status:** Transitive dependency via `cloud-native-utils/security`

#### When to Use

**Do not use directly.** Use `cloud-native-utils/security` which handles:
- PKCE code generation
- Token exchange
- Session management

---

### LM Studio (OpenAI-compatible API)

- **Purpose:** Local LLM inference server with OpenAI-compatible REST API. Enables running large language models locally for AI agent functionality.
- **Website:** [lmstudio.ai](https://lmstudio.ai/)
- **API Compatibility:** OpenAI Chat Completions API (`/v1/chat/completions`)
- **Status:** External service (not a Go dependency)

#### Integration in This Project

The `agent` bounded context uses LM Studio through `internal/adapters/outbound/lmstudio_client.go`:

| Component | Purpose |
|-----------|---------|
| `LMStudioClient` | Implements `agent.LLMClient` interface |
| `Run(ctx, messages)` | Sends conversation to LM Studio, returns LLM response |
| `convertToAPIMessages` | Translates domain `agent.Message` to OpenAI format |
| `convertToResponse` | Translates OpenAI response to domain `agent.LLMResponse` |

#### Configuration

| Environment Variable | Purpose | Default |
|---------------------|---------|---------|
| `LM_STUDIO_URL` | Base URL for LM Studio API | `http://localhost:1234` |
| `LM_STUDIO_MODEL` | Model identifier to use | `default` (uses loaded model) |

#### When to Use

| Concern | Pattern |
|---------|---------|
| **AI agent tasks** | Use `agent.TaskService` with `LMStudioClient` |
| **Custom LLM calls** | Implement `agent.LLMClient` interface |
| **Testing** | Create mock `LLMClient` for unit tests |
| **Integration testing** | Use `//go:build integration` tag with real LM Studio |

#### Integration Patterns

**Creating an LLM Client:**
```go
import "github.com/andygeiss/go-ddd-hex-starter/internal/adapters/outbound"

client := outbound.NewLMStudioClient(
    os.Getenv("LM_STUDIO_URL"),
    os.Getenv("LM_STUDIO_MODEL"),
)
```

**Using with TaskService:**
```go
import "github.com/andygeiss/go-ddd-hex-starter/internal/domain/agent"

service := agent.NewTaskService(client, toolExecutor, publisher)
result, err := service.RunTask(ctx, agentInstance, task)
```

**Custom HTTP Client (for timeouts):**
```go
httpClient := &http.Client{Timeout: 60 * time.Second}
client := outbound.NewLMStudioClient(url, model).WithHTTPClient(httpClient)
```

#### Cautions

- **Local service required:** LM Studio must be running locally with the API server enabled
- **Model loading:** Ensure a model is loaded in LM Studio before making requests
- **Timeouts:** LLM responses can be slow; use appropriate HTTP client timeouts (60s+ recommended)
- **Memory usage:** Large models require significant RAM/VRAM
- **No authentication:** LM Studio local API has no authentication; only use on trusted networks
- **OpenAI compatibility:** Not all OpenAI features are supported (e.g., function calling may vary by model)

#### Alternative Implementations

The `agent.LLMClient` interface is designed to be implementation-agnostic:

```go
type LLMClient interface {
    Run(ctx context.Context, messages []Message) (LLMResponse, error)
}
```

To use a different LLM provider (OpenAI, Anthropic, Ollama, etc.):
1. Create a new adapter in `internal/adapters/outbound/`
2. Implement the `agent.LLMClient` interface
3. Inject the new client into `TaskService`

---

## Python Standard Library (Tooling)

Python scripts in `tools/` use only the standard library to avoid external dependencies.

### Modules Used

| Module | Purpose | Used In |
|--------|---------|--------|
| `unittest` | Test framework | `tools/*_test.py` |
| `subprocess` | Run shell commands | `tools/create_pgo.py` |
| `glob` | File pattern matching | `tools/create_pgo.py` |
| `secrets` | Cryptographic random | `tools/change_me_local_secret.py` |
| `pathlib` | Path manipulation | `tools/*.py` |
| `shutil` | File operations | `tools/create_pgo.py` |
| `tempfile` | Temporary files | `tools/*_test.py` |
| `unittest.mock` | Test mocking | `tools/*_test.py` |

### When to Use

| Concern | Module | Pattern |
|---------|--------|---------|
| **Unit testing** | `unittest` | `class TestX(unittest.TestCase)` |
| **Assertions** | `unittest` | `self.assertEqual(actual, expected)` |
| **Mocking** | `unittest.mock` | `@patch('module.function')` |
| **Secure random** | `secrets` | `secrets.choice(alphabet)` |
| **Run commands** | `subprocess` | `subprocess.run(cmd, check=True)` |
| **File patterns** | `glob` | `glob.glob('*.pprof')` |

### Why No External Dependencies

- **Simplicity:** No `pip install` required for development
- **Portability:** Works on any system with Python 3
- **Stability:** Standard library is stable and well-documented
- **CI/CD:** No additional setup steps in pipelines

---

## Cross-cutting Concerns and Recommended Patterns

### Testing

| Concern | Vendor | Pattern |
|---------|--------|---------|
| Assertions | `cloud-native-utils/assert` | `assert.That(t, msg, actual, expected)` |
| Mock repositories | `cloud-native-utils/resource` | `resource.NewMockAccess[K, V]().WithCreateFn(...)` |
| In-memory messaging | `cloud-native-utils/messaging` | `messaging.NewInternalDispatcher()` |
| Mock functions | Standard library | Create interface + mock struct |

### Logging

| Concern | Vendor | Pattern |
|---------|--------|---------|
| Structured logging | `cloud-native-utils/logging` | `logging.NewJsonLogger()` |
| Request logging | `cloud-native-utils/logging` | `logging.WithLogging(logger, handler)` |

### HTTP

| Concern | Vendor | Pattern |
|---------|--------|---------|
| Server creation | `cloud-native-utils/security` | `security.NewServer(mux)` |
| Routing + sessions | `cloud-native-utils/security` | `security.NewServeMux(ctx, efs)` (accepts `fs.FS`) |
| Authentication | `cloud-native-utils/security` | `security.WithAuth(sessions, handler)` |
| Templating | `cloud-native-utils/templating` | `templating.NewEngine(efs)` (accepts `fs.FS`) |
| Redirects | `cloud-native-utils/redirecting` | `redirecting.Redirect(w, r, path)` |

### Messaging

| Concern | Vendor | Pattern |
|---------|--------|---------|
| Event dispatcher | `cloud-native-utils/messaging` | `messaging.NewExternalDispatcher()` |
| Publishing | `messaging.Dispatcher` | `dispatcher.Publish(ctx, msg)` |
| Subscribing | `messaging.Dispatcher` | `dispatcher.Subscribe(ctx, topic, handler)` |

### Persistence

| Concern | Vendor | Pattern |
|---------|--------|---------|
| CRUD interface | `cloud-native-utils/resource` | `resource.Access[K, V]` interface |
| JSON file storage | `cloud-native-utils/resource` | `resource.NewJsonFileAccess[K, V](file)` |
| In-memory storage | `cloud-native-utils/resource` | `resource.NewInMemoryAccess[K, V]()` |

### AI / LLM Integration

| Concern | Vendor | Pattern |
|---------|--------|---------|
| LLM client interface | Domain port | `agent.LLMClient` in `internal/domain/agent/ports_outbound.go` |
| LM Studio adapter | Project adapter | `outbound.NewLMStudioClient(url, model)` |
| Tool execution | Domain port | `agent.ToolExecutor` interface |
| Agent loop | Domain service | `agent.NewTaskService(client, executor, publisher)` |

### Resilience

| Concern | Vendor | Pattern |
|---------|--------|---------|
| Circuit breaker | `cloud-native-utils/stability` | `stability.Breaker(fn, threshold)` |
| Retry | `cloud-native-utils/stability` | `stability.Retry(fn, attempts, delay)` |
| Throttle | `cloud-native-utils/stability` | `stability.Throttle(fn, limit)` |
| Timeout | `cloud-native-utils/stability` | `stability.Timeout(fn, duration)` |

---

## Vendors to Avoid

### Testing Frameworks

**Avoid:** testify, gomega, ginkgo, goconvey

**Reason:** `cloud-native-utils/assert` provides sufficient assertion capabilities with a minimal API. The standard `testing` package plus `assert.That` covers all testing needs without additional complexity.

### Logging Libraries

**Avoid:** logrus, zap, zerolog

**Reason:** `cloud-native-utils/logging` wraps the standard `log/slog` package with JSON formatting. Using a different logger would create inconsistency and duplicate functionality.

### HTTP Routers

**Avoid:** gorilla/mux, chi, gin, echo

**Reason:** Go 1.22+ `http.ServeMux` supports pattern matching (e.g., `GET /path/{id}`). Combined with `cloud-native-utils/security.NewServeMux`, there's no need for third-party routers.

### Configuration Libraries

**Avoid:** viper, envconfig, godotenv

**Reason:** This project uses environment variables directly via `os.Getenv`. The `.env` file is loaded by Docker Compose or shell, not by the application. Adding configuration libraries adds complexity without benefit.

### Dependency Injection Frameworks

**Avoid:** wire, dig, fx

**Reason:** This project uses constructor-based dependency injection. Functions like `NewIndexingService(reader, repo, publisher)` provide explicit, traceable dependencies without magic.

### LLM Client Libraries

**Avoid:** langchaingo, go-openai (as direct dependencies)

**Reason:** The `agent.LLMClient` interface provides a minimal abstraction over LLM APIs. The `LMStudioClient` adapter implements OpenAI-compatible HTTP calls directly (~100 lines). This keeps the codebase lean and avoids pulling in large framework dependencies. If you need a different LLM provider, implement the interface directly rather than adding a framework.

---

## Adding New Vendor Libraries

Before adding a new vendor library:

1. **Check cloud-native-utils first:** It may already have the functionality you need.
2. **Evaluate necessity:** Can the functionality be implemented in 50-100 lines of Go?
3. **Consider maintenance:** Is the library actively maintained? Does it have a stable API?
4. **Check compatibility:** Does it work with Go 1.25+ and the existing stack?

If adding a new library:

1. Add to `go.mod` with `go get`
2. Document in this file with:
   - Purpose and repository link
   - Key packages/functions used
   - Integration patterns
   - Cautions and constraints
3. Update `CONTEXT.md` if it affects architecture or conventions
