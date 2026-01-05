# CONTEXT.md

This file is the authoritative project context for AI coding agents, retrieval systems, and advanced developers working on this codebase.

---

## 1. Project Purpose

This repository is a **production-ready Go template** that demonstrates Domain-Driven Design (DDD) and Hexagonal Architecture (Ports & Adapters) patterns. It provides a complete foundation for building cloud-native applications with:

- Clean separation of domain logic from infrastructure concerns
- Event-driven architecture with pub/sub messaging
- OIDC authentication via Keycloak
- HTTP server with HTMX-powered UI
- Profile-Guided Optimization (PGO) for performance
- Docker/Podman containerization with multi-stage builds

The template serves as a **reference implementation** for AI coding agents and developers to spin up new Go projects following established architectural patterns and conventions.

---

## 2. Technology Stack

| Category | Technology |
|----------|------------|
| **Language** | Go 1.25+ |
| **Architecture** | Domain-Driven Design, Hexagonal Architecture |
| **HTTP Framework** | `net/http` (stdlib) with `cloud-native-utils/security` |
| **Frontend** | HTMX, HTML templates |
| **Authentication** | OpenID Connect (OIDC) via Keycloak |
| **Messaging** | Internal pub/sub or Apache Kafka via `cloud-native-utils/messaging` |
| **Build/Task Runner** | `just` (justfile) |
| **Containerization** | Docker/Podman with multi-stage builds |
| **Orchestration** | Docker Compose |
| **Profiling** | Go PGO with benchmark-driven profiles |
| **Vendor Library** | `github.com/andygeiss/cloud-native-utils` v0.4.8 |

---

## 3. High-Level Architecture

This template follows **Hexagonal Architecture** (Ports & Adapters) with **Domain-Driven Design** principles:

```
┌─────────────────────────────────────────────────────────────────┐
│                        cmd/ (Entry Points)                       │
│   ┌─────────────┐                      ┌─────────────┐          │
│   │   cli/      │                      │   server/   │          │
│   │  main.go    │                      │   main.go   │          │
│   └──────┬──────┘                      └──────┬──────┘          │
└──────────┼─────────────────────────────────────┼────────────────┘
           │                                     │
           ▼                                     ▼
┌─────────────────────────────────────────────────────────────────┐
│               internal/adapters/ (Infrastructure)                │
│   ┌────────────────────────┐    ┌────────────────────────┐      │
│   │       inbound/         │    │       outbound/        │      │
│   │  - HTTP handlers       │    │  - Repositories        │      │
│   │  - File readers        │    │  - Event publishers    │      │
│   │  - Event subscribers   │    │  - External services   │      │
│   │  - Router, Middleware  │    │                        │      │
│   └───────────┬────────────┘    └────────────┬───────────┘      │
└───────────────┼──────────────────────────────┼──────────────────┘
                │         Ports (Interfaces)   │
                ▼                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    internal/domain/ (Core)                       │
│   ┌────────────────────────┐    ┌────────────────────────┐      │
│   │        event/          │    │       indexing/        │      │
│   │  - Event interface     │    │  - Aggregate (Index)   │      │
│   │  - Publisher port      │    │  - Entities (FileInfo) │      │
│   │  - Subscriber port     │    │  - Value Objects       │      │
│   │  - Factory/Handler     │    │  - Domain Events       │      │
│   │                        │    │  - Ports (interfaces)  │      │
│   │                        │    │  - Service             │      │
│   └────────────────────────┘    └────────────────────────┘      │
└─────────────────────────────────────────────────────────────────┘
```

### Architectural Layers

| Layer | Location | Responsibility |
|-------|----------|----------------|
| **Domain** | `internal/domain/` | Pure business logic, aggregates, entities, value objects, domain events, port interfaces |
| **Adapters** | `internal/adapters/` | Infrastructure implementations (HTTP, filesystem, messaging, persistence) |
| **Application** | `cmd/` | Entry points, dependency wiring, lifecycle management |

### Data Flow

1. **Inbound adapters** receive external requests (HTTP, events, CLI)
2. **Domain services** orchestrate business workflows using domain objects
3. **Outbound adapters** persist state or publish events to external systems
4. **Domain events** decouple bounded contexts via pub/sub

---

## 4. Directory Structure (Contract)

```
.
├── .github/
│   └── agents/              # AI agent configuration files
├── cmd/
│   ├── cli/                 # CLI application entry point
│   │   ├── main.go
│   │   ├── main_test.go     # Benchmarks for PGO
│   │   └── assets/          # Embedded CLI assets
│   └── server/              # HTTP server entry point
│       ├── main.go
│       ├── main_test.go     # Benchmarks for PGO
│       └── assets/          # Embedded web assets (templates, static)
├── internal/
│   ├── adapters/
│   │   ├── inbound/         # Driving adapters (HTTP, filesystem, events)
│   │   │   ├── router.go           # HTTP route registration
│   │   │   ├── middleware.go       # HTTP middleware (logging, security)
│   │   │   ├── http_*.go           # HTTP handlers
│   │   │   ├── file_reader.go      # Filesystem adapter
│   │   │   └── event_subscriber.go # Event subscription adapter
│   │   └── outbound/        # Driven adapters (repositories, publishers)
│   │       ├── file_index_repository.go  # Persistence adapter
│   │       └── event_publisher.go        # Event publishing adapter
│   └── domain/
│       ├── event/           # Cross-cutting event abstractions
│       │   ├── event.go            # Event interface
│       │   ├── event_factory.go    # EventFactoryFn type
│       │   ├── event_handler.go    # EventHandlerFn type
│       │   ├── event_publisher.go  # EventPublisher port
│       │   └── event_subscriber.go # EventSubscriber port
│       └── indexing/        # Example bounded context
│           ├── aggregate.go        # Index aggregate root
│           ├── entities.go         # FileInfo entity
│           ├── value_objects.go    # IndexID value object
│           ├── events.go           # Domain events
│           ├── ports_inbound.go    # FileReader port
│           ├── ports_outbound.go   # IndexRepository port
│           └── service.go          # IndexingService
├── tools/                   # Python utilities for development
│   ├── change_me_local_secret.py   # Keycloak secret rotation
│   └── create_pgo.py               # PGO profile generation
├── .env.example             # Environment variable template
├── .justfile                # Task runner configuration
├── .keycloak.json.example   # Keycloak realm template
├── docker-compose.yml       # Development stack (Keycloak, Kafka, App)
├── Dockerfile               # Multi-stage production build
├── go.mod                   # Go module definition
├── CONTEXT.md               # This file (AI agent context)
├── README.md                # Human-facing documentation
└── VENDOR.md                # Vendor library documentation
```

### Rules for New Code

#### Adding a new bounded context (domain)

1. Create a new package under `internal/domain/<context>/`
2. Define aggregates, entities, and value objects
3. Define domain events in `events.go`
4. Define inbound ports in `ports_inbound.go`
5. Define outbound ports in `ports_outbound.go`
6. Implement the domain service in `service.go`
7. Add tests in `*_test.go` files within the same package

#### Adding a new inbound adapter

1. Place implementation in `internal/adapters/inbound/`
2. Implement the port interface defined in the domain
3. Name HTTP handlers as `http_<feature>.go`
4. Register routes in `router.go`
5. Add tests in `<adapter>_test.go`

#### Adding a new outbound adapter

1. Place implementation in `internal/adapters/outbound/`
2. Implement the port interface defined in the domain
3. Prefer reusing `cloud-native-utils/resource` for persistence
4. Prefer reusing `cloud-native-utils/messaging` for pub/sub
5. Add tests in `<adapter>_test.go`

#### Adding new entry points

1. Create a new package under `cmd/<name>/`
2. Wire dependencies in `main.go`
3. Embed assets using `//go:embed assets`
4. Add benchmarks in `main_test.go` for PGO

---

## 5. Coding Conventions

### 5.1 General

- Keep modules small and focused on a single responsibility
- Prefer composition over inheritance
- Domain logic must have zero infrastructure dependencies
- Use dependency injection via interfaces (ports)
- Pass `context.Context` through all layers for cancellation and timeouts
- Services orchestrate workflows; adapters handle I/O

### 5.2 Naming

| Element | Convention | Example |
|---------|------------|---------|
| **Packages** | Lowercase, single word | `indexing`, `inbound`, `outbound` |
| **Files** | Snake_case | `file_reader.go`, `event_publisher.go` |
| **Aggregates** | PascalCase noun | `Index`, `Order`, `User` |
| **Entities** | PascalCase noun | `FileInfo`, `LineItem` |
| **Value Objects** | PascalCase with `ID` suffix for identifiers | `IndexID`, `OrderID` |
| **Interfaces (ports)** | PascalCase noun describing capability | `FileReader`, `IndexRepository`, `EventPublisher` |
| **Services** | PascalCase with `Service` suffix | `IndexingService` |
| **Events** | PascalCase with `Event` prefix | `EventFileIndexCreated` |
| **Event topics** | Constant with `EventTopic` prefix | `EventTopicFileIndexCreated` |
| **HTTP handlers** | `HttpView<Name>` or `Http<Action><Resource>` | `HttpViewIndex`, `HttpViewLogin` |
| **Test functions** | `Test_<Struct>_<Method>_With_<Condition>_Should_<Result>` | `Test_IndexingService_CreateIndex_With_Mockup_Should_Return_Two_Entries` |
| **Benchmark functions** | `Benchmark_<Struct>_<Method>_With_<Condition>_Should_<Result>` | `Benchmark_FileReader_ReadFileInfos_With_1000_Entries_Should_Be_Fast` |

### 5.3 Error Handling & Logging

**Error handling:**
- Return errors up the call stack; let the entry point decide how to handle them
- Wrap errors with context using `fmt.Errorf("context: %w", err)` when adding information
- Domain logic returns domain-specific errors; adapters translate them
- Never panic except for truly unrecoverable programmer errors

**Logging:**
- Use `cloud-native-utils/logging.NewJsonLogger()` for structured JSON logs
- Use `logging.WithLogging(logger, handler)` middleware for HTTP request logging
- Log levels: `Info` for normal operations, `Error` for failures
- Include correlation data: method, path, duration for HTTP; topic, event type for messaging

### 5.4 Testing

**Framework:** Go stdlib `testing` package with `cloud-native-utils/assert`

**Patterns:**
- Follow Arrange-Act-Assert (AAA) pattern
- Use table-driven tests for multiple scenarios
- Mock dependencies using interface implementations
- Use `resource.NewMockAccess[K, V]()` for repository mocks

**Naming:** `Test_<Struct>_<Method>_With_<Condition>_Should_<Result>`

**Organization:**
- Unit tests in `*_test.go` files alongside source files
- Use `_test` package suffix for black-box testing (e.g., `indexing_test`)
- Benchmarks in entry point test files (`cmd/*/main_test.go`) for PGO

**Example:**
```go
func Test_IndexingService_CreateIndex_With_Mockup_Should_Return_Two_Entries(t *testing.T) {
    // Arrange
    sut, _ := setupIndexingService()
    path := "testdata/index.json"
    ctx := context.Background()

    // Act
    err := sut.CreateIndex(ctx, path)
    files, err2 := sut.IndexFiles(ctx, path)

    // Assert
    assert.That(t, "err must be nil", err == nil, true)
    assert.That(t, "err2 must be nil", err2 == nil, true)
    assert.That(t, "index must have two entries", len(files) == 2, true)
}
```

### 5.5 Formatting & Linting

- Use `gofmt` or `goimports` for formatting
- No explicit linter configuration in this template (relies on Go defaults)
- Follow standard Go idioms and effective Go guidelines

---

## 6. Agent-Specific Patterns

### Domain Event Flow

```
Domain Service → Event Publisher (port) → Outbound Adapter → Messaging Dispatcher
                                                                    │
                                                                    ▼
Domain Handler ← Event Subscriber (port) ← Inbound Adapter ← Messaging Dispatcher
```

### Event Structure

Events implement the `event.Event` interface:

```go
type Event interface {
    Topic() string
}
```

Events are defined in the bounded context (`internal/domain/<context>/events.go`):

```go
type EventFileIndexCreated struct {
    IndexID   IndexID `json:"index_id"`
    FileCount int     `json:"file_count"`
}

func (e *EventFileIndexCreated) Topic() string {
    return EventTopicFileIndexCreated
}
```

### Adding a New Domain Event

1. Define the event struct in `internal/domain/<context>/events.go`
2. Add a topic constant: `EventTopic<Name>`
3. Implement `Topic() string` method
4. Use builder pattern with `New<Event>()` and `With<Field>()` methods
5. Publish via the `EventPublisher` port in the domain service
6. Subscribe via `EventSubscriber.Subscribe(ctx, topic, factory, handler)`

### Adding a New Inbound Adapter (HTTP Handler)

1. Create `internal/adapters/inbound/http_<feature>.go`
2. Define response struct: `HttpView<Feature>Response`
3. Create handler function returning `http.HandlerFunc`
4. Use `templating.Engine` for rendering views
5. Register route in `router.go` with middleware chain

### Adding a New Outbound Adapter (Repository)

1. Create `internal/adapters/outbound/<name>_repository.go`
2. Implement the port interface from the domain
3. Prefer `resource.JsonFileAccess[K, V]` for file-based storage
4. Return the port interface type (not the concrete type)

---

## 7. Using This Repo as a Template

### What Must Be Preserved

| Element | Reason |
|---------|--------|
| Directory structure (`cmd/`, `internal/adapters/`, `internal/domain/`) | Enforces hexagonal architecture |
| Port/adapter pattern | Maintains testability and flexibility |
| Event-driven patterns | Enables loose coupling |
| `cloud-native-utils` dependency | Provides cross-cutting utilities |
| Testing conventions | Ensures consistent test quality |
| PGO workflow | Maintains performance optimization capability |

### What Should Be Customized

| Element | Customization |
|---------|---------------|
| `internal/domain/indexing/` | Replace with your bounded contexts |
| `cmd/cli/` and `cmd/server/` | Adapt entry points for your use case |
| `.env.example` | Update environment variables for your app |
| `APP_NAME`, `APP_SHORTNAME`, `APP_DESCRIPTION` | Your application identity |
| Templates in `assets/templates/` | Your UI design |
| Keycloak realm configuration | Your OIDC setup |

### Steps to Create a New Project

1. **Clone/copy this template**
   ```bash
   git clone https://github.com/andygeiss/go-ddd-hex-starter my-project
   cd my-project
   rm -rf .git && git init
   ```

2. **Update project metadata**
   - Rename module in `go.mod`
   - Update `APP_NAME`, `APP_SHORTNAME`, `APP_DESCRIPTION` in `.env.example`
   - Update `README.md` with your project description
   - Update `LICENSE` if needed

3. **Configure local development**
   ```bash
   cp .env.example .env
   cp .keycloak.json.example .keycloak.json
   ```

4. **Replace example domain**
   - Remove or rename `internal/domain/indexing/`
   - Create your bounded contexts under `internal/domain/`
   - Implement aggregates, entities, value objects, events, ports, and services

5. **Implement adapters**
   - Create inbound adapters for your entry points
   - Create outbound adapters for your persistence and messaging needs

6. **Wire dependencies**
   - Update `cmd/*/main.go` to inject your services and adapters

7. **Add tests and benchmarks**
   - Write tests following naming conventions
   - Add benchmarks in `cmd/*/main_test.go` for PGO

8. **Run the stack**
   ```bash
   just up
   ```

---

## 8. Key Commands & Workflows

| Command | Description |
|---------|-------------|
| `just setup` | Install dependencies (docker-compose, just) via Homebrew |
| `just build` or `just b` | Build Docker image with PGO optimization |
| `just up` or `just u` | Generate secrets, build image, start all services |
| `just down` or `just d` | Stop and remove all Docker Compose services |
| `just test` or `just t` | Run all tests with coverage |
| `just serve` | Run HTTP server locally (outside Docker) |
| `just run` | Run CLI application locally |
| `just profile` | Generate CPU profiles for PGO |

### Environment Selection

- **Local development (without Docker):** Use `just serve` or `just run`; set `KAFKA_BROKERS=localhost:9092` in `.env`
- **Containerized development:** Use `just up`; default `KAFKA_BROKERS=kafka:9092` works

### Service URLs (after `just up`)

| Service | URL |
|---------|-----|
| Application | http://localhost:8080 |
| Keycloak Admin | http://localhost:8180/admin (admin:admin) |
| Kafka | localhost:9092 |

---

## 9. Important Notes & Constraints

### Security

- **Never commit secrets:** `.env` and `.keycloak.json` are local-only files
- **OIDC secrets:** Use `CHANGE_ME_LOCAL_SECRET` placeholder; `just up` rotates it automatically
- **Production secrets:** Use a proper secret store (Vault, K8s Secrets)
- **HTTP security headers:** Added via `WithSecurityHeaders` middleware

### Performance

- **PGO:** Run `just profile` to generate CPU profiles; `just build` uses them for optimization
- **Benchmarks:** Critical paths have benchmarks in `cmd/*/main_test.go`
- **Hot paths:** HTTP handlers, event processing, file I/O

### Platform Assumptions

- **OS:** macOS or Linux (uses Homebrew for setup)
- **Container runtime:** Podman for builds, Docker Compose for orchestration
- **Go version:** 1.25+ (uses `b.Loop()` in benchmarks)

### Deprecated/Experimental Areas

- None currently

### Do Not Touch

- `cloud-native-utils` integration patterns without reading `VENDOR.md`
- Port interface definitions in domain packages (changes cascade to adapters)

---

## 10. How AI Tools and RAG Should Use This File

### Priority

1. **Read this file first** before making architectural decisions or significant changes
2. Consult `README.md` for human-facing documentation and feature descriptions
3. Consult `VENDOR.md` for approved library usage patterns

### Constraints

- Treat directory structure and layering rules as hard constraints
- Prefer `cloud-native-utils` over custom implementations for cross-cutting concerns
- Follow naming conventions exactly for consistency
- Maintain port/adapter separation in all changes

### When Modifying the Codebase

1. Identify which layer the change belongs to (domain, adapter, cmd)
2. Respect the dependency direction: adapters depend on domain, never reverse
3. Define interfaces (ports) in the domain; implement them in adapters
4. Wire dependencies in entry points (`cmd/*/main.go`)
5. Add tests following the naming convention

### Updating This File

If architectural patterns or conventions change:

1. Update this file to reflect the new reality
2. Ensure changes are consistent with `README.md` and `VENDOR.md`
3. Test that the documented commands and workflows still work
