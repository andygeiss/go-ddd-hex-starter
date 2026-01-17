<p align="center">
<img src="https://github.com/andygeiss/hotel-booking/blob/main/cmd/server/assets/static/img/icon-192.png?raw=true" width="100"/>
</p>

# Go Hotel Booking

[![Go Reference](https://pkg.go.dev/badge/github.com/andygeiss/hotel-booking.svg)](https://pkg.go.dev/github.com/andygeiss/hotel-booking)
[![License](https://img.shields.io/github/license/andygeiss/hotel-booking)](https://github.com/andygeiss/hotel-booking/blob/master/LICENSE)
[![Releases](https://img.shields.io/github/v/release/andygeiss/hotel-booking)](https://github.com/andygeiss/hotel-booking/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/andygeiss/hotel-booking)](https://goreportcard.com/report/github.com/andygeiss/hotel-booking)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/f9f01632dff14c448dbd4688abbd04e8)](https://app.codacy.com/gh/andygeiss/hotel-booking/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)
[![Codacy Badge](https://app.codacy.com/project/badge/Coverage/f9f01632dff14c448dbd4688abbd04e8)](https://app.codacy.com/gh/andygeiss/hotel-booking/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_coverage)

A hotel reservation and payment management system built with Go, demonstrating Domain-Driven Design (DDD) and Hexagonal Architecture (Ports & Adapters) patterns.

<p align="center">
<img src="https://github.com/andygeiss/hotel-booking/blob/main/cmd/server/assets/static/img/login.png?raw=true" width="300"/>
</p>

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Domain Model](#domain-model)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Usage](#usage)
- [Testing](#testing)
- [Configuration](#configuration)
- [Using as a Template](#using-as-a-template)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

This repository provides a reference implementation for structuring Go applications with clean architecture principles. It demonstrates how to:

- Organize code using **Hexagonal Architecture** (Ports & Adapters)
- Apply **Domain-Driven Design** tactical patterns (aggregates, entities, value objects, domain events)
- Implement the **Saga Pattern** for distributed workflow orchestration
- Integrate authentication via **OIDC/Keycloak**
- Implement event-driven communication with **Apache Kafka**

The template includes a fully-featured **Booking** bounded context with reservations, payments, availability checking, and orchestrated workflows.

---

## Key Features

- **Complete Booking Domain** — Reservations, payments, availability checking, and guest management
- **Developer Experience** — `just` task runner, golangci-lint, comprehensive test coverage
- **Domain-Driven Design** — Aggregates, entities, value objects, domain services, and domain events
- **Event Streaming** — Kafka-based pub/sub for domain events
- **Hexagonal Architecture** — Clear separation between domain logic and infrastructure
- **OIDC Authentication** — Keycloak integration with session management
- **Production-Ready Docker** — Multi-stage build with PGO optimization (~5-10MB images)
- **Progressive Web App** — Service worker, manifest, and offline support for installable web apps
- **Saga Pattern** — Orchestrated booking workflow with compensation on failure

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Entry Points (cmd/)                      │
│                       server/main.go                        │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                  Inbound Adapters                           │
│   HTTP handlers, event subscribers                          │
│              internal/adapters/inbound/                     │
└──────────────────────────┬──────────────────────────────────┘
                           │ implements ports
┌──────────────────────────▼──────────────────────────────────┐
│                     Domain Layer                            │
│   Bounded contexts: booking/                                │
│   Aggregates, entities, value objects, services, ports      │
│                   internal/domain/                          │
└──────────────────────────┬──────────────────────────────────┘
                           │ defines ports
┌──────────────────────────▼──────────────────────────────────┐
│                  Outbound Adapters                          │
│   Event publisher, repositories, payment gateway, notifier  │
│              internal/adapters/outbound/                    │
└─────────────────────────────────────────────────────────────┘
```

### Bounded Contexts

| Context | Purpose |
|---------|---------|
| `booking` | Hotel reservations, payments, availability, and booking orchestration |

---

## Domain Model

The booking domain implements a complete hotel reservation system with two aggregate roots:

### Reservation Aggregate

```
Reservation (Aggregate Root)
├── ReservationID (Value Object)
├── GuestInfo (Entity)
│   ├── GuestID
│   ├── Email
│   └── Name
├── RoomID (Value Object)
├── DateRange (Value Object)
│   ├── CheckIn
│   └── CheckOut
├── Money (Value Object)
│   ├── Amount
│   └── Currency
└── ReservationStatus (Value Object)
    States: pending → confirmed → active → completed
                  ↘ cancelled
```

**Business Rules:**
- Minimum 1 night stay required
- Check-in must be in the future
- Cannot cancel within 24 hours of check-in
- Same-day checkout/check-in allowed (no overlap)
- Cancelled reservations don't block availability

### Payment Aggregate

```
Payment (Aggregate Root)
├── PaymentID (Value Object)
├── ReservationID (links to Reservation)
├── Money (Value Object)
├── PaymentStatus (Value Object)
│   States: pending → authorized → captured
│                  ↘ failed      ↘ refunded
└── PaymentAttempts (Entity Collection)
    └── PaymentAttempt
        ├── Sequence
        ├── Status
        ├── ErrorCode
        └── Timestamp
```

**Business Rules:**
- Authorization-Capture pattern (Authorize → Capture)
- Failed payments can be retried (max 3 attempts)
- Only captured payments can be refunded

### Booking Orchestration (Saga Pattern)

```
CompleteBooking Workflow:
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ 1. Create       │───▶│ 2. Authorize    │───▶│ 3. Capture      │
│    Reservation  │    │    Payment      │    │    Payment      │
│    (pending)    │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │                      │
                              ▼ (on failure)         ▼
                       Cancel Reservation     Refund + Cancel
                                                     │
                                                     ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │ 5. Send         │◀───│ 4. Confirm      │
                       │    Notification │    │    Reservation  │
                       └─────────────────┘    └─────────────────┘
```

### Domain Events

| Event | Trigger |
|-------|---------|
| `ReservationCreated` | New reservation created (pending) |
| `ReservationConfirmed` | Reservation confirmed after payment |
| `ReservationCancelled` | Reservation cancelled by guest or system |
| `PaymentAuthorized` | Payment authorization successful |
| `PaymentCaptured` | Payment captured (charged) |
| `PaymentFailed` | Payment attempt failed |
| `PaymentRefunded` | Payment refunded to guest |

---

## Project Structure

```
hotel-booking/
├── .justfile                     # Task runner commands
├── cmd/server/                   # HTTP server entry point
│   ├── main.go                   # Bootstrap: DI, server setup, lifecycle
│   └── assets/
│       ├── static/               # CSS, JS, images (embedded)
│       └── templates/            # HTML templates (*.tmpl, embedded)
├── docker-compose.yml            # Dev stack (Keycloak, Kafka, app)
├── Dockerfile                    # Multi-stage production build
├── internal/
│   ├── adapters/
│   │   ├── inbound/              # HTTP handlers, event subscribers
│   │   │   ├── router.go         # HTTP routing & middleware
│   │   │   ├── http_booking_*.go # Booking UI handlers
│   │   │   └── event_subscriber.go
│   │   └── outbound/             # Repositories, gateways, publishers
│   │       ├── file_*_repository.go      # JSON file storage
│   │       ├── repository_availability_checker.go
│   │       ├── mock_payment_gateway.go   # Payment simulation
│   │       ├── mock_notification_service.go
│   │       └── event_publisher.go
│   └── domain/
│       └── booking/              # Booking bounded context
│           ├── reservation.go    # Reservation aggregate
│           ├── payment.go        # Payment aggregate
│           ├── reservation_service.go
│           ├── payment_service.go
│           ├── orchestration_service.go  # Saga workflow
│           └── ports.go          # Interface definitions
└── reservations.json             # Runtime data (created by app)
```

---

## Getting Started

### Prerequisites

- **Docker** and **Docker Compose** (or Podman)
- **Go 1.24+**
- **golangci-lint** (for linting/formatting)
- **just** task runner

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/andygeiss/hotel-booking.git
   cd hotel-booking
   ```

2. **Install development tools:**
   ```bash
   just setup
   ```
   This installs `docker-compose`, `golangci-lint`, `just`, and `podman` via Homebrew.

3. **Configure environment:**
   ```bash
   cp .env.example .env
   cp .keycloak.json.example .keycloak.json
   ```

4. **Start the development stack:**
   ```bash
   just up
   ```
   This generates secrets, builds the Docker image, and starts Keycloak, Kafka, and the application.

5. **Access the application:**
   - **App:** http://localhost:8080/ui
   - **Keycloak Admin:** http://localhost:8180/admin (admin:admin)

---

## Usage

### Commands

| Command | Description |
|---------|-------------|
| `just build` | Build Docker image |
| `just down` | Stop all services |
| `just fmt` | Format code |
| `just lint` | Run linter |
| `just profile` | Generate CPU profile for PGO |
| `just setup` | Install development dependencies |
| `just test` | Run unit tests with coverage |
| `just test-integration` | Run integration tests |
| `just up` | Start full development stack |

### Booking Workflow

Once the application is running:

1. **Login** at http://localhost:8080/ui/login via Keycloak
2. **View Reservations** at `/ui/reservations` to see your bookings
3. **Create Reservation** at `/ui/reservations/new`:
   - Select a room and dates
   - Total is calculated automatically (nights × room price)
   - Submit to create a pending reservation
4. **View Details** at `/ui/reservations/{id}` to see reservation status
5. **Cancel Reservation** from the detail page (if >24 hours before check-in)

### API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/ui/` | GET | Dashboard (authenticated) |
| `/ui/login` | GET | Login page |
| `/ui/reservations` | GET | List user's reservations |
| `/ui/reservations/new` | GET | Reservation form |
| `/ui/reservations` | POST | Create reservation |
| `/ui/reservations/{id}` | GET | Reservation detail |
| `/ui/reservations/{id}/cancel` | POST | Cancel reservation |

---

## Testing

### Unit Tests

Run all unit tests with coverage:

```bash
just test
```

This generates `coverage.pprof` with coverage metrics.

### Integration Tests

Integration tests require external services (Kafka, Keycloak):

```bash
just test-integration
```

### Test Organization

- Unit tests are colocated with source files (`*_test.go`)
- Integration tests are tagged with `//go:build integration`
- Test fixtures live in `testdata/` directories

### Domain Test Coverage

The domain layer includes comprehensive tests for:

| Component | Test Coverage |
|-----------|---------------|
| **Reservation Aggregate** | State transitions, validation, overlap detection, cancellation rules |
| **Payment Aggregate** | Authorization-capture flow, retry logic, attempt tracking |
| **Value Objects** | Money formatting, ID types, date range calculations |
| **Business Rules** | 24-hour cancellation policy, minimum stay, future check-in |

---

## Configuration

Configuration is managed via environment variables. Copy `.env.example` to `.env` and customize:

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_NAME` | Display name for UI and PWA | `Hotel Booking` |
| `APP_DESCRIPTION` | Application description | `Hotel reservation and payment management system` |
| `APP_SHORTNAME` | Docker image/container name | `hotel-booking` |
| `APP_VERSION` | Version for PWA cache busting | `1.0.0` |
| `KAFKA_BROKERS` | Kafka broker addresses | `localhost:9092` |
| `OIDC_CLIENT_ID` | OIDC client ID | `hotel-booking` |
| `OIDC_CLIENT_SECRET` | OIDC client secret | Auto-generated |
| `OIDC_ISSUER` | Keycloak realm URL | `http://localhost:8180/realms/local` |
| `PORT` | HTTP server port | `8080` |

See `.env.example` for the complete list with documentation.

---

## Using as a Template

### Quick Start

1. **Clone and reinitialize:**
   ```bash
   git clone https://github.com/andygeiss/hotel-booking my-project
   cd my-project
   rm -rf .git && git init
   ```

2. **Update module path:**
   ```bash
   go mod edit -module github.com/yourorg/my-project
   # Update import paths in all .go files
   ```

3. **Configure project identity:**
   ```bash
   cp .env.example .env
   # Edit APP_NAME, APP_SHORTNAME, APP_DESCRIPTION, APP_VERSION
   ```

4. **Add your domain logic:**
   - Create bounded contexts in `internal/domain/`
   - Implement adapters in `internal/adapters/`
   - Wire up entry points in `cmd/`

### What to Keep

- Directory structure (`cmd/`, `internal/adapters/`, `internal/domain/`)
- Hexagonal architecture pattern
- `cloud-native-utils` as infrastructure library
- `context.Context` threading through all layers

### What to Customize

- Bounded contexts and domain logic (replace `booking/` with your domain)
- Static assets and templates in `cmd/server/assets/`
- Environment configuration in `.env`
- Docker Compose services as needed
- Swap mock adapters (payment gateway, notification service) for real implementations

---

## Contributing

1. Ensure all tests pass: `just test`
2. Ensure code is formatted and linted: `just fmt && just lint`
3. Follow hexagonal architecture patterns (ports in domain, adapters in adapters/)
4. Update documentation if architecture changes

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
