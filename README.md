<p align="center">
  <img src="cmd/server/assets/static/img/icon-192.png" alt="Go DDD Hexagonal Starter logo" width="96" height="96">
</p>

# Go DDD Hexagonal Starter

[![Go Reference](https://pkg.go.dev/badge/go-ddd-hex-starter.svg)](https://pkg.go.dev/go-ddd-hex-starter)
[![Go Report Card](https://goreportcard.com/badge/github.com/andygeiss/go-ddd-hex-starter)](https://goreportcard.com/report/github.com/andygeiss/go-ddd-hex-starter)
[![License](https://img.shields.io/github/license/andygeiss/go-ddd-hex-starter.svg)](https://github.com/andygeiss/go-ddd-hex-starter)
[![Release](https://img.shields.io/github/v/release/andygeiss/go-ddd-hex-starter.svg)](https://github.com/andygeiss/go-ddd-hex-starter/releases)

> A production-ready Go template demonstrating Domain-Driven Design (DDD) and Hexagonal Architecture (Ports and Adapters) for building maintainable, testable, and scalable applications.

---

## Overview

**go-ddd-hex-starter** is a reusable blueprint for Go applications that need clean architecture, well-defined boundaries, and explicit dependency management. It provides a reference implementation of the Ports and Adapters pattern with working examples that you can adapt to your own projects.

### Motivation

Most Go projects begin with a simple structure that grows organically. As features accumulate, boundaries blur, business logic leaks into HTTP handlers, and tests become difficult to write. This template solves those problems by:

- **Isolating business logic** in a pure domain layer free from infrastructure concerns.
- **Defining explicit contracts** through port interfaces that adapters implement.
- **Enforcing dependency rules** where dependencies point inward only—domain never imports adapters.
- **Providing working examples** including a CLI indexer and an HTTP server with OIDC authentication.
- **Serving as a template** for developers and AI coding agents that need a consistent, well-documented structure.

---

## Key Features

- ✅ **Hexagonal Architecture**: Domain at the center, adapters on the outside, application layer wiring everything together.
- ✅ **Domain-Driven Design**: Aggregates, entities, value objects, domain events, and services.
- ✅ **Port and Adapter Pattern**: Inbound ports for driving the application, outbound ports for driven infrastructure.
- ✅ **Explicit Dependency Injection**: No global state, no init magic—all wiring happens in `cmd/*/main.go`.
- ✅ **Profile-Guided Optimization (PGO)**: Benchmark-driven binary optimization for production performance.
- ✅ **OIDC Authentication**: Integrated with Keycloak via `cloud-native-utils/security`.
- ✅ **Structured Logging**: JSON logs with `log/slog` and contextual fields.
- ✅ **Event-Driven Messaging**: Domain event publishing via `cloud-native-utils/messaging`.
- ✅ **Generic Repository Pattern**: Type-safe CRUD with `cloud-native-utils/resource`.
- ✅ **Embedded Assets**: Static files and templates via `embed.FS` for portable binaries.
- ✅ **Multi-Stage Docker Builds**: Optimized containers with scratch runtime for minimal image size.
- ✅ **Task Runner Workflow**: `just` commands for build, test, profile, run, and deployment.

---

## Architecture Overview

This project implements **Hexagonal Architecture** with three distinct layers:

```
┌──────────────────────────────────────────────────────────────┐
│                    Application Layer                         │
│                  cmd/cli/main.go                             │
│                  cmd/server/main.go                          │
│        (Entry points, wire adapters to domain)               │
└─────────────────────────┬────────────────────────────────────┘
                          │
┌─────────────────────────▼────────────────────────────────────┐
│                       Domain Layer                           │
│                   internal/domain/                           │
│    (Pure business logic, defines Ports as interfaces)        │
└─────────────────────────▲────────────────────────────────────┘
                          │
┌─────────────────────────┴────────────────────────────────────┐
│  ┌────────────────┐           ┌──────────────────┐           │
│  │ Inbound        │───────────│ Outbound         │           │
│  │ Adapters       │           │ Adapters         │           │
│  │ (Driving)      │           │ (Driven)         │           │
│  └────────────────┘           └──────────────────┘           │
│                     Adapters Layer                           │
│                 internal/adapters/                           │
└──────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

| Layer | Location | Responsibility |
|-------|----------|----------------|
| **Domain** | `internal/domain/` | Pure business logic. Defines Ports (interfaces). Zero infrastructure knowledge. |
| **Adapters** | `internal/adapters/` | Infrastructure implementations. Inbound = driving; Outbound = driven. |
| **Application** | `cmd/` | Entry points that wire adapters to domain services via dependency injection. |

### Dependency Rule

Source code dependencies point **inward** only:

- Adapters depend on Domain
- Domain depends on nothing (no infrastructure imports)
- Application layer depends on both (wires them together)

---

## Project Structure

```
go-ddd-hex-starter/
├── .justfile                 # Task runner commands (build, test, profile, run, serve)
├── go.mod                    # Go module definition (Go 1.25+)
├── CONTEXT.md                # AI/developer context documentation
├── VENDOR.md                 # Vendor library documentation (cloud-native-utils)
├── README.md                 # User-facing documentation (this file)
├── Dockerfile                # Multi-stage container build
├── docker-compose.yml        # Dev stack (Keycloak, Kafka, app)
├── cpuprofile.pprof          # PGO profile data (auto-generated)
├── coverage.pprof            # Test coverage data (auto-generated)
├── bin/                      # Compiled binaries (gitignored)
├── tools/                    # Build/dev scripts (Python helpers)
├── cmd/                      # Application entry points
│   ├── cli/                  # CLI application
│   │   ├── main.go           # Wires adapters, runs file indexing
│   │   ├── main_test.go      # Benchmarks for PGO
│   │   └── assets/           # Embedded assets (embed.FS)
│   └── server/               # HTTP server application
│       ├── main.go           # Wires adapters, starts server with OIDC
│       └── assets/
│           ├── static/       # CSS, JS (base.css, htmx.min.js, theme.css)
│           └── templates/    # HTML templates (index.tmpl, login.tmpl)
└── internal/
    ├── adapters/             # Infrastructure implementations
    │   ├── inbound/          # Driving adapters (filesystem, HTTP handlers, routing)
    │   │   ├── file_reader.go
    │   │   ├── router.go
    │   │   ├── http_index.go
    │   │   ├── http_login.go
    │   │   ├── http_view.go
    │   │   └── middleware.go
    │   └── outbound/         # Driven adapters (persistence, messaging)
    │       ├── file_index_repository.go
    │       └── event_publisher.go
    └── domain/               # Pure business logic
        ├── event/            # Domain event interface
        │   └── event.go
        └── indexing/         # Bounded Context: Indexing
            ├── aggregate.go      # Aggregate Root (Index)
            ├── entities.go       # Entities (FileInfo)
            ├── value_objects.go  # Value Objects (IndexID) + Domain Events
            ├── service.go        # Domain Service (IndexingService)
            ├── ports_inbound.go  # Interfaces for driving adapters
            └── ports_outbound.go # Interfaces for driven adapters
```

### Notes on Structure

- **Bounded Contexts**: Each major area of business logic lives in `internal/domain/<context>/` (e.g., `indexing`).
- **Ports**: Defined as interfaces in the domain layer (`ports_inbound.go`, `ports_outbound.go`).
- **Adapters**: Concrete implementations in `internal/adapters/inbound/` (driving) and `outbound/` (driven).
- **Wiring**: All dependency injection happens explicitly in `cmd/*/main.go`—no global state.
- **Tests**: Co-located with source files as `*_test.go`, following the Arrange-Act-Assert pattern.

---

## Conventions & Standards

### Coding Style Disclaimer

The coding style in this repository reflects a combination of widely used practices, prior experience, and personal preference, and is influenced by the Go projects on [github.com/andygeiss](https://github.com/andygeiss). There is no single "best" project setup; you are encouraged to adapt this structure, evolve your own style, and use this repository as a starting point for your own projects.

### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Files | `snake_case.go` | `file_reader.go`, `http_index.go` |
| Packages | lowercase, singular | `indexing`, `event`, `inbound` |
| Interfaces | Noun describing capability | `FileReader`, `IndexRepository` |
| Structs | PascalCase noun | `Index`, `FileInfo`, `EventPublisher` |
| Methods | PascalCase verb phrase | `ReadFileInfos`, `CreateIndex` |
| Constructors | `New<Type>()` | `NewIndex()`, `NewFileReader()` |
| Value Objects | PascalCase, `ID` suffix for identifiers | `IndexID` |
| Domain Events | `Event<Action>` | `EventFileIndexCreated` |
| HTTP Handlers | `Http<View/Action><Name>` | `HttpViewIndex`, `HttpViewLogin` |
| Middleware | `With<Capability>` | `WithLogging`, `WithSecurityHeaders` |
| Tests | `Test_<Struct>_<Method>_With_<Condition>_Should_<Result>` | `Test_Index_Hash_With_No_FileInfos_Should_Return_Valid_Hash` |
| Benchmarks | `Benchmark_<Struct>_<Method>_With_<Condition>_Should_<Result>` | `Benchmark_Main_With_Inbound_And_Outbound_Adapters_Should_Run_Efficiently` |

### Core Principles

- **Context Propagation**: Always pass `context.Context` as the first parameter.
- **Error Handling**: Return errors; don't panic except for unrecoverable situations.
- **Dependency Injection**: Wire dependencies via constructors in `cmd/*/main.go`.
- **No Global State**: Avoid `init()` functions for wiring; use explicit construction.
- **Testing**: Follow Arrange-Act-Assert (AAA) pattern with `cloud-native-utils/assert`.
- **Logging**: Use `log/slog` with JSON structured logs at the adapter/application layer only.

---

## Using This Repository as a Template

### What to Preserve

- Hexagonal architecture with domain → adapters → application layering.
- Dependency rule: dependencies point inward only.
- Port interface pattern in the domain layer.
- Constructor pattern: `New<Type>()` returning interface types.
- Test naming conventions and AAA (Arrange-Act-Assert) pattern.
- Context propagation through all layers.
- Benchmark-driven PGO workflow.

### What to Customize

- **Domain logic**: Replace or extend the `indexing` bounded context with your own.
- **Adapters**: Add database, HTTP, queue, or API adapters as needed.
- **Entry points**: Add CLIs, servers, workers under `cmd/`.
- **Embedded assets**: Replace contents of `cmd/*/assets/` with your own static files and templates.
- **Configuration**: Add environment variables or config files for your deployment.
- **Templates**: Customize UI in `assets/templates/`.

### Steps to Create a New Project

1. Clone or use this repository as a GitHub template.
2. Update module name in [go.mod](go.mod).
3. Find and replace `go-ddd-hex-starter` references throughout the codebase.
4. Clear or replace the example `indexing` bounded context in [internal/domain/indexing/](internal/domain/indexing/).
5. Create your own bounded contexts in `internal/domain/`.
6. Implement inbound and outbound adapters in `internal/adapters/`.
7. Wire adapters to domain services in [cmd/*/main.go](cmd/).
8. Write tests following the established conventions.
9. Run `just profile` to generate a PGO baseline.
10. Build with `just build` for an optimized production binary.

---

## Getting Started

### Prerequisites

- **Go 1.25+**: [Download](https://go.dev/dl/)
- **just**: Task runner (`brew install just` on macOS)
- **Docker/Podman**: For containerized workflows
- **Graphviz**: For PGO visualization (`brew install graphviz`)

### Installation

```bash
# Clone the repository
git clone https://github.com/andygeiss/go-ddd-hex-starter.git
cd go-ddd-hex-starter

# Install dependencies
just setup

# Copy example configuration
cp .env.example .env
cp .keycloak.json.example .keycloak.json
```

### Quick Start

```bash
# Run CLI application (file indexer)
just run

# Start HTTP server (with OIDC authentication)
just serve

# Run tests with coverage
just test

# Generate PGO profile
just profile

# Build optimized binary
just build

# Start dev stack (Keycloak, Kafka, app)
just up

# Stop dev stack
just down
```

---

## Running, Scripts, and Workflows

### Task Runner Commands (`just`)

| Command | Description |
|---------|-------------|
| `just build` | Build optimized binary with PGO to `bin/` |
| `just run` | Build and run the CLI application |
| `just serve` | Run the HTTP server (development mode) |
| `just test` | Run all unit tests with coverage |
| `just profile` | Run benchmarks and generate PGO profile |
| `just up` | Start Docker Compose dev stack (Keycloak, Kafka, app) |
| `just down` | Stop Docker Compose services |
| `just setup` | Install dependencies via Homebrew |

### Manual Commands

```bash
# Run all tests
go test ./...

# Run with verbose output and coverage
go test -v -coverprofile=coverage.pprof ./internal/...

# Run specific test
go test -v -run Test_Index_Hash ./internal/domain/indexing/

# Build without PGO
go build -o ./bin/cli ./cmd/cli/main.go

# Build with PGO
go build -ldflags "-s -w" -pgo cpuprofile.pprof -o ./bin/cli ./cmd/cli/main.go

# Run HTTP server
PORT=8080 go run ./cmd/server/main.go
```

### Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `APP_NAME` | Application display name | — |
| `APP_DESCRIPTION` | Application description | — |
| `APP_SHORTNAME` | Short name for Docker image | — |
| `PORT` | HTTP server port | `8080` |
| `OIDC_CLIENT_ID` | OIDC client identifier | — |
| `OIDC_CLIENT_SECRET` | OIDC client secret | — |
| `OIDC_ISSUER` | OIDC issuer URL | — |
| `OIDC_REDIRECT_URL` | OIDC callback URL | — |

---

## Usage Examples

### Example 1: CLI File Indexer

The CLI application indexes files in a directory and persists the result to JSON.

```bash
# Run the CLI
just run

# Or directly with go
go run ./cmd/cli/main.go
```

**What it does**:
1. Reads file metadata from `cmd/cli/assets/` using the `FileReader` adapter.
2. Creates an `Index` aggregate with `FileInfo` entities in the domain layer.
3. Persists the index to `file-index.json` using the `FileIndexRepository` adapter.
4. Publishes an `EventFileIndexCreated` domain event via the `EventPublisher` adapter.

### Example 2: HTTP Server with OIDC

The HTTP server provides a web UI with OIDC authentication.

```bash
# Start Keycloak and Kafka
just up

# In another terminal, run the server
just serve
```

**What it does**:
1. Starts an HTTP server on the configured port (default: `8080`).
2. Serves static assets (CSS, JS) and templates (HTML) from embedded `assets/` directory.
3. Protects routes with OIDC middleware using Keycloak.
4. Provides login, index, and view handlers.
5. Logs all requests with structured JSON logs.

### Example 3: Adding a New Bounded Context

To add a new domain context (e.g., `orders`):

1. Create directory: `internal/domain/orders/`
2. Define files:
   - `aggregate.go`: Order aggregate root
   - `entities.go`: OrderItem entities
   - `value_objects.go`: OrderID, Money value objects
   - `ports_inbound.go`: OrderService interface
   - `ports_outbound.go`: OrderRepository interface
   - `service.go`: OrderService implementation
3. Create adapters:
   - `internal/adapters/inbound/http_orders.go`: HTTP handler
   - `internal/adapters/outbound/postgres_order_repository.go`: Repository implementation
4. Wire in `cmd/server/main.go`:

```go
orderRepo := postgres.NewOrderRepository(db)
orderService := orders.NewOrderService(orderRepo)
router.HandleFunc("/orders", inbound.HttpViewOrders(orderService))
```

---

## Testing & Quality

### Testing Strategy

- **Unit Tests**: Test domain logic in isolation using mocks for repositories.
- **Adapter Tests**: Test adapters with real or in-memory backends.
- **Benchmark Tests**: Measure performance and generate PGO profiles.

### Running Tests

```bash
# Run all tests
just test

# Run tests with coverage report
go test -v -coverprofile=coverage.pprof ./internal/...
go tool cover -func=coverage.pprof

# Run specific test
go test -v -run Test_Index_Hash ./internal/domain/indexing/

# Run benchmarks
go test -bench=. -benchmem ./cmd/cli/
```

### Test Conventions

All tests follow the **Arrange-Act-Assert (AAA)** pattern:

```go
func Test_Index_Hash_With_No_FileInfos_Should_Return_Valid_Hash(t *testing.T) {
    // Arrange
    index := indexing.Index{ID: "empty-index", FileInfos: []indexing.FileInfo{}}

    // Act
    hash := index.Hash()

    // Assert
    assert.That(t, "empty index must have a valid hash (size of 64 bytes)", len(hash), 64)
}
```

### Assertions

Use `cloud-native-utils/assert` for readable test assertions:

```go
import "github.com/andygeiss/cloud-native-utils/assert"

assert.That(t, "error message describing what should be true", got, want)
```

---

## CI/CD

This template is designed to integrate with standard CI/CD pipelines. While specific workflow files are not included, the following approach is recommended:

### Recommended CI Pipeline

1. **Checkout**: Clone the repository.
2. **Setup Go**: Install Go 1.25+ on the runner.
3. **Dependencies**: Run `go mod download`.
4. **Tests**: Run `just test` to execute unit tests with coverage.
5. **Linting**: Run `gofmt -l .` and `go vet ./...`.
6. **Build**: Run `just build` to create optimized binaries.
7. **Profile**: Run `just profile` to generate PGO data (optional).
8. **Docker Build**: Build container image with `docker build -t myapp:latest .`.
9. **Push**: Push image to container registry.

### Profile-Guided Optimization (PGO)

PGO improves runtime performance by using real benchmark data:

```bash
# Generate PGO profile
just profile

# Build with PGO
just build
```

The `just profile` command:
1. Runs benchmarks in `cmd/cli/main_test.go`.
2. Merges CPU profiles into `cpuprofile.pprof`.
3. Generates `cpuprofile.svg` for visualization.

The `just build` command uses the PGO profile to optimize the binary.

---

## Limitations and Roadmap

### Current Limitations

- **No ORM**: Uses JSON file storage via `cloud-native-utils/resource`. For relational databases, implement custom repository adapters.
- **Basic HTTP Routing**: Uses `net/http` standard library. For complex routing, consider adding a router library.
- **Local OIDC Only**: Example uses Keycloak in Docker Compose. For production, configure external OIDC provider.
- **No GraphQL or gRPC**: Only HTTP REST and CLI examples are included.

### Roadmap

Future enhancements may include:

- [ ] PostgreSQL repository adapter example
- [ ] gRPC server example
- [ ] WebSocket support
- [ ] OpenAPI/Swagger documentation generation
- [ ] Kubernetes deployment manifests
- [ ] GitHub Actions workflow examples
- [ ] Load testing with `k6` or similar

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) if available, or open an issue to discuss new features.

---

## Resources

- **Documentation**:
  - [CONTEXT.md](CONTEXT.md): Architectural constraints and conventions for AI agents and developers.
  - [VENDOR.md](VENDOR.md): Documentation for `cloud-native-utils` library.
- **External Libraries**:
  - [cloud-native-utils](https://github.com/andygeiss/cloud-native-utils): Utilities for cloud-native Go applications.
  - [pkg.go.dev/go-ddd-hex-starter](https://pkg.go.dev/go-ddd-hex-starter): API documentation.
- **Learning Resources**:
  - [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
  - [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html)
  - [Go Project Layout](https://github.com/golang-standards/project-layout)

---

## Acknowledgments

This template is built on patterns and practices from the Go community and the projects on [github.com/andygeiss](https://github.com/andygeiss). Special thanks to the maintainers of:

- [cloud-native-utils](https://github.com/andygeiss/cloud-native-utils) for providing reusable components.
- The Go team for an excellent standard library and tooling.

---

**Happy coding!** If you find this template useful, please star the repository and share it with others. Feedback and contributions are always welcome.
