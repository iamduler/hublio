# 21 - Go Project Structure

> Product: Hublio
> Version: 1.0
> Status: Architecture Freeze v1

---

# 1. Purpose

This document defines the Go project structure of Hublio.

The project structure translates the Domain Model into a maintainable codebase.

The goals are

* Clear package boundaries
* Explicit dependencies
* High cohesion
* Low coupling
* Testability
* Scalability

The project structure must reflect the Architecture, not the database.

---

# 2. Design Principles

The Go codebase follows these principles.

* Domain First
* Package by Bounded Context
* Dependency Inversion
* Composition over Inheritance
* Small Interfaces
* Explicit Dependencies

Business logic belongs to the Domain.

Infrastructure depends on the Domain.

---

# 3. High-Level Structure

Hublio is a monorepo. The Go backend is a single module under `apps/api` (module
path `hublio`); the Next.js applications and shared TypeScript live alongside it.

```text
├── apps/
│   ├── api/                 # Go backend (module: hublio)
│   │   ├── cmd/
│   │   │   ├── api/         # REST API entrypoint
│   │   │   └── worker/      # background worker entrypoint
│   │   ├── internal/        # Go business modules (bounded contexts)
│   │   ├── migrations/      # SQL migrations
│   │   ├── tests/
│   │   ├── go.mod
│   │   └── sqlc.yml
│   ├── web/                 # Next.js user workspace (@hublio/web)
│   └── admin/               # Next.js admin console (@hublio/admin)
│
├── packages/               # Shared TypeScript packages
│   ├── ui/                  # @hublio/ui — shadcn + common components + theme
│   ├── config/              # @hublio/config — shared tsconfig base
│   └── sdk/                 # @hublio/sdk — types generated from openapi.yaml
│
├── api/openapi/            # OpenAPI spec (source of truth)
│
├── deploy/                 # Dockerfiles + docker-compose
│
├── system/                 # infra config (redis.conf)
│
├── scripts/                # dev/ops scripts
│
├── docs/
├── .github/
├── AGENTS.md
├── go.work
├── Makefile
├── package.json            # pnpm workspace root + turbo
├── pnpm-workspace.yaml
└── turbo.json
```

apps/
    Go backend (`api`) and Next.js applications (`web`, `admin`)

packages/
    Shared TypeScript packages (`ui`, `config`, `sdk`)

apps/api/internal/
    Go business modules (bounded contexts)
---

# 4. Application Entry Points

## cmd/api

Starts the REST API server.

Responsibilities

* HTTP Server
* Routing
* Middleware
* Dependency Injection

---

## cmd/worker

Starts background workers.

Responsibilities

* Queue Processing
* Execution Processing
* Retry
* Scheduled Jobs

Workers share the same Domain Model as the API.

---

# 5. Bounded Context Layout

Every Bounded Context follows the same structure.

```text
apps/api/internal/orchestration/

    application/

    domain/

    infrastructure/

    interfaces/
```

This layout is consistent across all contexts.

---

# 6. Domain Layer

The Domain layer contains business rules.

Typical contents

* Aggregates
* Entities
* Value Objects
* Domain Events
* Repository Interfaces

The Domain must not depend on infrastructure.

---

# 7. Application Layer

The Application layer coordinates business use cases.

Typical responsibilities

* Use Cases
* Commands
* Queries
* Transaction Management
* Event Publishing

Application Services orchestrate Aggregates.

---

# 8. Infrastructure Layer

Infrastructure implements technical concerns.

Examples

* PostgreSQL
* Redis
* HTTP Clients
* Queue
* Encryption
* Logging

Infrastructure depends on the Domain.

Never the opposite.

---

# 9. Interface Layer

The Interface layer exposes the application.

Examples

* REST Handlers
* Webhooks
* Queue Consumers
* CLI Commands

Interfaces translate external requests into Application use cases.

---

# 10. Connector Runtime

Each Connector lives in its own package.

```text
apps/api/internal/integration/connectors/

    misa/

    vnpt/

    nhanh/

    kiotviet/
```

Each Connector implements the same runtime contract.

No Connector depends on another Connector.

---

# 11. Shared Platform Packages

Shared platform functionality lives under

```text
apps/api/internal/platform/
```

Examples

* Authentication
* Authorization
* Configuration
* Middleware
* Validation
* Encryption
* Idempotency

Only cross-cutting concerns belong here.

---

# 12. Package Dependencies

Dependencies always point inward.

```text
Interfaces
      │
      ▼
Application
      │
      ▼
Domain
      ▲
      │
Infrastructure
```

The Domain is the center of the architecture.

---

# 13. Repository Pattern

Repository interfaces belong to the Domain.

Repository implementations belong to Infrastructure.

Example

```text
Domain

ExecutionRepository
```

↓

```text
Infrastructure

PostgresExecutionRepository
```

The Domain never imports PostgreSQL packages.

---

# 14. DTO Strategy

DTOs belong to the Interface layer.

The Domain never exposes DTOs.

Mappings occur in the Application layer.

The Domain works only with Domain objects.

---

# 15. Transaction Boundary

Transactions are managed in the Application layer.

Aggregates remain transaction-agnostic.

One transaction should modify one Aggregate.

---

# 16. Testing Strategy

Each layer has its own testing responsibility.

Domain

* Unit Tests

Application

* Use Case Tests

Infrastructure

* Integration Tests

Interfaces

* API Tests

Business rules should be tested without external dependencies.

---

# 17. Dependency Rules

Allowed

* Interface → Application
* Application → Domain
* Infrastructure → Domain

Not allowed

* Domain → Infrastructure
* Domain → Interface
* Application → Interface

Dependency direction must remain consistent.

---

# 18. Version 1 Constraints

Version 1 intentionally excludes

* Microservices
* Plugin Loading
* Runtime Module Discovery
* Hexagonal Frameworks
* Backend code generation

The project remains a modular monolith.

Note: frontend TypeScript **types** are generated from `api/openapi/openapi.yaml`
into `packages/sdk` (see AGENTS.md → OpenAPI). This consumes the single spec and
is not backend code generation.

---

# 19. Guiding Principles

The codebase reflects the Domain.

Packages represent Bounded Contexts.

Dependencies point inward.

Business logic remains independent of technology.

Keep the project structure simple, modular, and maintainable.
