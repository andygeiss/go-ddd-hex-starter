<p align="center">
<img src="https://github.com/andygeiss/go-ddd-hex-starter/blob/main/cmd/server/assets/static/img/icon-192.png?raw=true" width="100"/>
</p>

# Go DDD Hexagonal Starter

[![Go Reference](https://pkg.go.dev/badge/github.com/andygeiss/go-ddd-hex-starter.svg)](https://pkg.go.dev/github.com/andygeiss/go-ddd-hex-starter)
[![License](https://img.shields.io/github/license/andygeiss/go-ddd-hex-starter)](https://github.com/andygeiss/go-ddd-hex-starter/blob/master/LICENSE)
[![Releases](https://img.shields.io/github/v/release/andygeiss/go-ddd-hex-starter)](https://github.com/andygeiss/go-ddd-hex-starter/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/andygeiss/go-ddd-hex-starter)](https://goreportcard.com/report/github.com/andygeiss/go-ddd-hex-starter)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/f9f01632dff14c448dbd4688abbd04e8)](https://app.codacy.com/gh/andygeiss/go-ddd-hex-starter/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)
[![Codacy Badge](https://app.codacy.com/project/badge/Coverage/f9f01632dff14c448dbd4688abbd04e8)](https://app.codacy.com/gh/andygeiss/go-ddd-hex-starter/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_coverage)

A production-ready Go starter template demonstrating Domain-Driven Design (DDD) and Hexagonal Architecture (Ports & Adapters) patterns.

<p align="center">
<img src="https://github.com/andygeiss/go-ddd-hex-starter/blob/main/cmd/server/assets/static/img/login.png?raw=true" width="300"/>
</p>

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Usage](#usage)
- [Testing](#testing)
- [Configuration](#configuration)
- [Using as a Template](#using-as-a-template)
- [Common Pitfalls](#common-pitfalls)
- [License](#license)

---

## Overview

This template provides a clean foundation for building cloud-native Go applications with proper separation of concerns. It includes a working example domain (`indexing`) that demonstrates:

- File indexing with domain events
- OIDC authentication via Keycloak
- Event streaming with Apache Kafka
- Server-side rendering with Go templates and HTMX

Use this as a starting point for your own projects or as a reference implementation for DDD/Hexagonal patterns in Go.

---

## Features

- **Hexagonal Architecture** — Clear separation between domain logic, inbound adapters (HTTP, events), and outbound adapters (persistence, messaging)
- **Domain-Driven Design** — Aggregates, entities, value objects, domain events, and application services
- **Event-Driven Architecture** — Kafka-based event publishing and subscribing
- **OIDC Authentication** — Keycloak integration with session management
- **Embedded Assets** — Static files and templates compiled into the binary
- **Production Container** — Multi-stage Dockerfile producing a minimal scratch-based image (~5-10MB)
- **Profile-Guided Optimization** — PGO support for optimized builds
- **Developer Experience** — `just` command runner for common tasks

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                   cmd/ (Entry Points)                       │
│           server/main.go       cli/main.go                  │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│           internal/adapters/inbound/ (Driving)              │
│      HTTP handlers, event subscribers, file readers         │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│                  internal/domain/                           │
│     Pure business logic: aggregates, entities, services     │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│           internal/adapters/outbound/ (Driven)              │
│       Event publisher (Kafka), file-based repository        │
└─────────────────────────────────────────────────────────────┘
```

For detailed architectural documentation, see [CONTEXT.md](CONTEXT.md).

---

## Project Structure

```
go-ddd-hex-starter/
├── cmd/
│   ├── cli/                  # CLI demo application
│   │   └── main_test.go      # CLI unit and integration tests
│   └── server/               # HTTP server with OIDC auth
│       ├── main_test.go      # Server integration tests
│       └── assets/           # Embedded templates and static files
├── internal/
│   ├── adapters/
│   │   ├── inbound/          # HTTP handlers, event subscribers
│   │   └── outbound/         # Repositories, event publishers
│   └── domain/
│       ├── event/            # Event interfaces
│       └── indexing/         # Example bounded context (with tests)
├── tools/                    # Python development utilities
│   ├── *_test.py             # Python unit tests
│   ├── change_me_local_secret.py
│   └── create_pgo.py
├── .justfile                 # Command runner recipes
├── docker-compose.yml        # Local development stack
└── Dockerfile                # Production container build
```

---

## Prerequisites

- **Go 1.25+**
- **Python 3** (for development tooling)
- **Docker** or **Podman** (for container builds)
- **Docker Compose** (for local development stack)
- **just** (command runner) — install via `brew install just`
- **golangci-lint** (for code quality) — install via `brew install golangci-lint`

---

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/andygeiss/go-ddd-hex-starter.git
cd go-ddd-hex-starter
```

### 2. Install Dependencies

```bash
just setup
```

This installs `docker-compose`, `golangci-lint`, `just` and `podman` via Homebrew.

### 3. Configure Environment

```bash
cp .env.example .env
cp .keycloak.json.example .keycloak.json
```

### 4. Start Services

```bash
just up
```

This will:
1. Generate a random OIDC client secret
2. Build the Docker image
3. Start Keycloak, Kafka, and the application

### 5. Access the Application

| Service | URL |
|---------|-----|
| Application | http://localhost:8080 |
| Keycloak Admin | http://localhost:8180/admin |

Default Keycloak credentials: `admin` / `admin`

---

## Usage

### Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `just build` | `just b` | Build Docker image |
| `just up` | `just u` | Start all services |
| `just down` | `just d` | Stop all services |
| `just fmt` | — | Format Go code with golangci-lint |
| `just lint` | — | Run golangci-lint checks |
| `just test` | `just t` | Run Go + Python tests with coverage |
| `just serve` | — | Run HTTP server locally |
| `just run` | — | Run CLI demo locally |
| `just profile` | — | Generate PGO profiles |

### Running Locally (without Docker)

To run the server locally (requires Kafka and Keycloak running separately):

```bash
just serve
```

To run the CLI demo:

```bash
just run
```

---

## Testing

Run all tests (Go + Python) with coverage:

```bash
just test
```

This runs:
- Go unit tests with coverage (~65% coverage target)
- Python unit tests for development tools

Or run tests separately:

```bash
# Go tests only
go test -v ./internal/...

# Python tests only
cd tools && python3 -m unittest -v
```

Tests follow the naming convention:
```
Test_<Struct>_<Method>_With_<Condition>_Should_<Result>
```

### Test Coverage

| Package | Coverage |
|---------|----------|
| `internal/domain/indexing` | ~87% |
| `internal/adapters/inbound` | ~54% |
| `internal/adapters/outbound` | ~67% |
| **Total** | **~65%** |

---

## Configuration

All configuration is via environment variables. See [.env.example](.env.example) for the complete list with documentation.

### Key Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `KAFKA_BROKERS` | Kafka broker addresses | `localhost:9092` |
| `OIDC_ISSUER` | Keycloak realm URL | `http://localhost:8180/realms/local` |
| `OIDC_CLIENT_ID` | OIDC client identifier | `template` |
| `OIDC_CLIENT_SECRET` | OIDC client secret | (generated) |

---

## Using as a Template

1. Update `go.mod` with your module path
2. Update all imports to match
3. Configure `.env.example` with your application metadata
4. Replace or extend `internal/domain/indexing/` with your domains
5. Implement adapters for your infrastructure

See [CONTEXT.md](CONTEXT.md) for detailed conventions and guidelines.

---

## Common Pitfalls

When adapting this template for your own project, watch out for these common issues:

### Module Path Changes

After renaming the module in `go.mod`, you must update **all** import paths across the codebase:

```bash
# Find all files with the old import path
grep -r "github.com/andygeiss/go-ddd-hex-starter" --include="*.go"

# Use sed or your editor to replace with your new module path
```

Missing imports will cause cryptic "package not found" errors at build time.

### Keycloak Configuration Alignment

The OIDC flow requires exact alignment between three places:

| Setting | Location | Must Match |
|---------|----------|------------|
| Realm name | Keycloak Admin UI | `OIDC_ISSUER` URL path |
| Client ID | Keycloak → Clients | `OIDC_CLIENT_ID` env var |
| Client Secret | Keycloak → Clients → Credentials | `OIDC_CLIENT_SECRET` env var |
| Valid Redirect URIs | Keycloak → Clients | Your app's callback URL |

**Tip:** After changing Keycloak settings, restart the application — secrets are read at startup.

### Kafka Topic Naming

Topics must be created before the application starts, or Kafka auto-creation must be enabled. Ensure:

- Topic names in code match what's configured in Kafka
- The `KAFKA_BROKERS` environment variable points to accessible brokers
- Network connectivity exists between your app container and Kafka

### Docker Compose Port Conflicts

Default ports may conflict with existing services:

| Service | Default Port | Change In |
|---------|--------------|-----------|
| Application | 8080 | `docker-compose.yml`, `PORT` env |
| Keycloak | 8180 | `docker-compose.yml`, `OIDC_ISSUER` |
| Kafka | 9092 | `docker-compose.yml`, `KAFKA_BROKERS` |

### Embedded Assets Not Updating

Go's `//go:embed` caches files at compile time. If you modify templates or static files:

```bash
# Force a clean rebuild
go build -a ./cmd/server
# Or use just
just build
```

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
