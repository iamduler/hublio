# AGENTS.md

# Hublio Engineering Constitution

Version: 1.0

Status: Architecture Freeze v1 (fan-out clarification: multi-Execution under one Intent)

---

# Purpose

This document defines the engineering rules that every AI coding agent and every developer **must follow** when contributing to Hublio.

The purpose is to preserve architecture consistency over time.

This document has higher priority than implementation preferences.

If generated code conflicts with this document, the generated code must be rejected.

---

# Product

Hublio is a Business Integration Platform and a Business Orchestration Platform.

Hublio is **NOT**

* ERP
* CRM
* Workflow Engine
* BPMN Engine
* Low-code Platform
* iPaaS Marketplace

The product focuses on integrating business systems and orchestrating business operations between external systems through a stable Canonical Model.

---

# Architecture

Architecture is frozen.

Do not introduce new architectural concepts.

Current architecture

Business Intent

↓

Execution

↓

Execution Step

↓

Connector Runtime

↓

External System

Everything must fit into this model.

---

# Architecture Principles

Always follow

* Domain Driven Design
* Clean Architecture
* Modular Monolith
* Package by Bounded Context

Never introduce

* Microservices
* CQRS
* Event Sourcing
* BPMN
* Workflow DSL
* Rule Engine
* Service Locator

unless explicitly requested.

---

# Domain

The Domain is the center of the application.

Business rules belong only inside the Domain.

The Domain must never depend on

* PostgreSQL
* Redis
* HTTP
* JSON
* REST
* Queue
* Frameworks

The Domain must remain pure Go.

---

# Dependency Direction

Dependencies always point inward.

Interfaces

↓

Application

↓

Domain

↑

Infrastructure

Never reverse this direction.

---

# Repository Layout

Hublio is a monorepo.

```text
apps/api/      # Go backend (module path: hublio)
apps/web/      # Next.js user workspace
apps/admin/    # Next.js admin (scaffold)
packages/      # shared JS/TS (ui, config, sdk)
api/openapi/   # OpenAPI source of truth
deploy/        # Dockerfiles + compose (deployment)
docs/          # architecture + schema docs
```

The Go module path remains `hublio`; all Go import paths are unchanged.

# Package Layout

Every Go bounded context (under `apps/api/internal/`) follows

```text
application/

domain/

infrastructure/

interfaces/
```

Do not invent additional architectural layers.

---

# Bounded Contexts

Current contexts

* Identity
* Integration
* Orchestration
* Transformation
* Events
* Platform

Do not create new contexts without approval.

---

# Aggregates

Current Aggregates

* Organization
* Workspace
* Connector
* Connection
* Intent
* Execution

Do not introduce additional Aggregates without approval.

Integration **configuration** entities (not Runtime Aggregates) may exist, e.g. **SyncRoute** (Workspace-scoped origin → destination fan-out). SyncRoute is not a Workflow.

---

# Runtime Model

Every business request becomes

Intent

↓

Execution(s)

↓

Execution Steps

↓

Succeeded / Failed / Terminal State

One Intent may spawn **one or more Executions** (fan-out).

Fan-out groups may be:

* sequential — start the next Execution after the previous reaches a terminal success (or stop per policy on failure)
* parallel — enqueue multiple Executions at once

Each Execution remains **internally sequential** (ordered Steps). Do not add parallel Steps inside a single Execution.

Fan-out is configuration-driven (e.g. SyncRoute destinations), not a Workflow Engine.

Do not introduce as Runtime concepts

* Workflow
* Task
* Activity (as a Runtime Aggregate / graph node)
* Pipeline
* Job Graph
* BPMN

Configuration may describe destination actions as “activities” on a SyncRoute without introducing an Activity Aggregate.

The Runtime Model is frozen with this fan-out clarification.

---

# Connectors

Every external system is implemented as a Connector.

A Connector

owns

* Authentication
* Provider API
* Provider DTO
* Error Translation

A Connector must never

* access the database
* publish events
* retry requests
* own business rules

---

# Inbound Webhooks and Polling

Allowed Integration concerns (not Workflow):

* Hublio may **generate and store** webhook secrets and validate them from request headers (constant-time compare). Never log secrets.
* Polling may persist a **watermark / cursor** (last poll trace) per SyncRoute and resource type in PostgreSQL.
* Redis is not the source of truth for watermarks or webhook secrets.

---

# Canonical Model

The Canonical Model is the source of truth.

Never expose provider models outside a Connector.

Never leak provider DTOs into

* Domain
* Application
* REST API

---

# Application Layer

Application coordinates use cases.

Application

may

* start transactions
* publish events
* call repositories

Application

must not

* implement business rules
* contain SQL
* contain HTTP handlers

---

# Infrastructure

Infrastructure implements technology.

Infrastructure depends on the Domain.

Never reverse the dependency.

Infrastructure includes

* PostgreSQL
* Redis
* HTTP
* Queue
* Encryption
* Logging

---

# REST API

REST handlers

only

* validate requests
* call Use Cases
* return responses

Handlers must not contain business logic.

---

# OpenAPI / API Docs

Source of truth for machine-readable HTTP API docs:

```text
api/openapi/openapi.yaml
```

Interactive UI (Scalar, Scramble-like) is served at `/docs` when docs are enabled
(`DEVELOPMENT_MODE=development` or `ENABLE_API_DOCS=true`). Spec URL: `/docs/openapi.yaml`.

**Convention (no codegen for now):**

* Whenever you **add, change, rename, or remove** a public HTTP route, request/response
  body, path/query parameter, auth scheme, or status code → **update `api/openapi/openapi.yaml`
  in the same change**.
* Do not introduce Swagger/swag annotations, backend OpenAPI codegen, or a second competing
 spec file unless explicitly requested.
* Keep Scalar / docs wiring in `apps/api/internal/platform/docsui` only — no business logic there.

**Frontend SDK (allowed):** `packages/sdk` contains **TypeScript types generated from
`api/openapi/openapi.yaml`** (via `openapi-typescript`). This *consumes* the single source of
truth; it is not a competing spec and is not backend codegen. Whenever the spec changes,
regenerate with `pnpm --filter @hublio/sdk generate` and commit `packages/sdk/src/schema.d.ts`
in the same change. Frontend apps import DTO types from `@hublio/sdk`.

**Frontend API access (httpOnly JWT proxy):** Browser never holds JWT or workspace API keys.

* Auth → browser → Next `/api/auth/*` → Go `/auth/*` (Next sets httpOnly `hublio_session` /
  `hublio_refresh`).
* Dashboard Go APIs → browser → Next `/api/go/*` via `lib/api/client` → Go with Bearer +
  `X-Workspace-ID` (identity, integration, intents, executions, events).
* Machine / external clients may call Go directly with `X-API-KEY`. Orchestration/events
  accept either API key or JWT + `X-Workspace-ID` (`MachineOrJWTMiddleware`).
* Details: `docs/24-nextjs-architecture.md` §8.1 and `apps/web/AGENTS.md`.

Agents and developers must apply this without being reminded.

---

# Repository

Repository interfaces belong to the Domain.

Repository implementations belong to Infrastructure.

Never place repository interfaces inside Infrastructure.

---

# Transactions

Transactions belong to Application.

Repositories never begin or commit transactions.

One Use Case owns one transaction.

---

# Events

Events describe facts.

Never publish commands as events.

Publish events only after successful state changes.

Events are immutable.

---

# Error Handling

Errors are values.

Wrap infrastructure errors.

Never leak PostgreSQL or Redis errors to the API.

Prefer domain errors.

---

# Logging

Use structured logging.

Never log

* passwords
* secrets
* tokens
* API keys

Every log should include

* correlation_id
* request_id (if available)
* execution_id (if available)

---

# Context

Use context.Context only for

* cancellation
* timeout
* deadlines
* tracing

Never store business objects inside Context.

---

# UUID

All IDs use UUID v7.

Generated by the application layer.

Repositories never generate IDs.

---

# Database

The Database Schema is frozen.

Do not

* rename tables
* rename columns
* add new relationships

unless explicitly requested.

Use PostgreSQL.

Redis is not a source of truth.

---

# JSONB

JSONB is allowed only for

* config
* payload
* context
* snapshot
* metadata

Do not store relational data inside JSONB.

---

# Multi-tenancy

Every business operation belongs to

* Organization
* Workspace

Never bypass tenant isolation.

---

# Testing

Write

* Unit Tests
* Integration Tests

Prefer table-driven tests.

Business rules must be testable without PostgreSQL or Redis.

---

# Code Generation

Generated code must

* compile
* follow package boundaries
* follow dependency direction
* use constructor injection
* avoid globals
* avoid reflection

Never generate placeholder implementations unless requested.

---

# Simplicity

Always choose the simplest solution that satisfies the current requirements.

Avoid premature abstraction.

Avoid unnecessary interfaces.

Avoid generic frameworks.

Avoid over-engineering.

---

# Architecture Freeze

The following concepts are frozen.

* Canonical Model
* Runtime Model (including controlled fan-out: multi-Execution under one Intent)
* Aggregate Model
* Database Model
* Package Layout

Controlled clarifications (still Freeze-compatible):

* Parallel **Executions** under one Intent — allowed for SyncRoute fan-out
* Parallel **Steps** inside one Execution — forbidden
* SyncRoute as Integration configuration — allowed; Workflow Engine — forbidden
* Webhook shared-secret validation and poll watermarks — allowed Integration/Platform concerns

Do not redesign these concepts.

Build on top of them.

---

# Pull Request Checklist

Before proposing code, verify

* Architecture is respected
* Package boundaries are respected
* Dependency direction is correct
* Domain remains pure
* Business rules stay in Domain
* No duplicated logic
* Tests are included
* Public APIs remain backward compatible
* HTTP route changes include an update to `api/openapi/openapi.yaml`

If any rule is violated, stop and refactor before submitting.

---

# Guiding Principle

The architecture is optimized for long-term maintainability.

Prefer explicit code over clever code.

Prefer stable architecture over short-term convenience.

Every contribution should make Hublio easier to understand, easier to test, and easier to evolve.

Repository Strategy

Hublio is maintained as a monorepo.

AI must respect package boundaries.

Do not move code between apps without explicit instruction.