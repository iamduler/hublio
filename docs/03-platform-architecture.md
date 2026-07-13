# Platform Architecture

> Product: Hublio
> Version: 1.0
> Status: Architecture Freeze v1

---

# 1. Purpose

This document defines the logical architecture of Hublio.

It describes

* Core platform components
* Component responsibilities
* Runtime execution flow
* Platform boundaries
* Architectural principles

This document intentionally does **NOT** define

* Database schema
* API specification
* Deployment topology
* Programming language
* Infrastructure implementation

---

# 2. Vision

Hublio is a Business Orchestration Platform.

Its purpose is to orchestrate business operations between software systems through a unified platform.

External systems never communicate directly with each other.

All communication passes through Hublio.

---

# 3. Architecture Principles

The platform follows these principles.

## API First

Every capability is exposed through Platform APIs.

---

## Canonical First

All business processing uses Canonical Data Models.

Provider-specific models never enter the platform core.

---

## Intent Driven

Clients express **what** they want to accomplish.

The platform determines **how** the request is executed.

---

## Connector Agnostic

Business logic must never depend on any specific provider.

Connectors are replaceable implementation details.

---

## Event Driven

Platform components communicate through immutable Events whenever appropriate.

---

## Stateless Services

Application services remain stateless.

Persistent state belongs to the platform.

---

## Separation of Responsibilities

Each component has a single responsibility.

Responsibilities must never overlap.

---

# 4. High-Level Architecture

```text
                    Client Applications
                            │
                            ▼
                      Platform API
                            │
                            ▼
                     Intent Processor
                            │
                            ▼
                  Orchestration Engine
                            │
          ┌─────────────────┼─────────────────┐
          ▼                 ▼                 ▼
Transformation Engine   Event Platform   Connector Runtime
          │                                   │
          ▼                                   ▼
   Canonical Models                    External Systems

Infrastructure

- PostgreSQL
- Redis
- Object Storage
- Observability Stack
```

---

# 5. Core Components

## Platform API

Responsibilities

* Authentication
* Authorization
* Request Validation
* Idempotency
* Rate Limiting
* API Versioning

The Platform API is the only public entry point.

---

## Intent Processor

Responsibilities

* Validate Business Intent
* Resolve Connection
* Resolve Capability
* Create Execution
* Publish Runtime Events

The Intent Processor accepts business requests.

It never performs business execution.

---

## Orchestration Engine

Responsibilities

* Execution Lifecycle
* Execution Step Coordination
* Retry
* Replay
* Scheduling
* Timeout
* Cancellation

The Orchestration Engine coordinates execution.

It never communicates directly with external providers.

---

## Transformation Engine

Responsibilities

* Canonical Mapping
* Validation
* Normalization
* Type Conversion
* Data Enrichment

The Transformation Engine translates between external models and Canonical Models.

---

## Connector Runtime

Responsibilities

* Provider Authentication
* API Communication
* Provider DTO Mapping
* Error Translation
* Webhook Verification
* Health Check

Each Connector is an Adapter for one external system.

---

## Event Platform

Responsibilities

* Publish Events
* Subscribe Events
* Persist Events
* Route Events

The Event Platform synchronizes platform components.

---

# 6. Runtime Flow

Every Business Intent follows the same execution flow.

```text
Business Intent

↓

Intent Processor

↓

Execution Created

↓

Execution Steps

↓

Transformation

↓

Connector Runtime

↓

External System

↓

Transformation

↓

Execution Completed

↓

Events Published
```

The execution flow is identical regardless of the external provider.

---

# 7. Runtime Model

Execution is the runtime representation of an accepted Intent.

Each Execution contains

* Context
* Status
* Execution Steps
* Timeline
* Snapshots
* Events
* Result

Execution is an internal platform concept.

Clients interact only with Business Intents.

---

# 8. Execution Step

Execution is composed of one or more Steps.

Version 1 supports sequential execution.

Typical Steps include

* Validate Request
* Transform Request
* Call Connector
* Transform Response
* Publish Events

Each Step has one responsibility.

---

# 9. Canonical Data Flow

Every integration follows the same transformation pipeline.

```text
External Request

↓

Provider DTO

↓

Transformation Engine

↓

Canonical Model

↓

Business Processing

↓

Canonical Model

↓

Transformation Engine

↓

Provider DTO

↓

External Response
```

Business logic only works with Canonical Models.

---

# 10. Connector Architecture

Every Connector is isolated.

A Connector owns

* Authentication
* Provider APIs
* Provider DTOs
* Error Translation
* Webhook Verification

The Platform owns

* Execution
* Retry
* Scheduling
* Events
* Observability
* Security

---

# 11. Event Flow

Platform components communicate through Events.

Examples

* IntentAccepted
* ExecutionCreated
* ExecutionStarted
* StepCompleted
* ExecutionCompleted
* ExecutionFailed
* ConnectionActivated

Events are immutable.

Events are append-only.

---

# 12. Platform Boundaries

Hublio owns

* Business Orchestration
* Execution Lifecycle
* Canonical Models
* Transformation
* Event Processing
* Connector Management
* Platform Configuration

External systems own

* Business Data
* Business Rules
* Provider APIs
* Provider State

Hublio is not the System of Record for customer business data.

---

# 13. Cross-cutting Capabilities

Every component follows the same platform capabilities.

* Security
* Multi-tenancy
* Audit
* Observability
* Idempotency
* Correlation
* Logging
* Metrics
* Tracing

These capabilities are implemented consistently across the platform.

---

# 14. Scalability

The following services are stateless.

* Platform API
* Intent Processor
* Orchestration Engine
* Transformation Engine
* Connector Runtime
* Workers

Persistent state is stored outside application processes.

Horizontal scaling should not require application changes.

---

# 15. Architecture Constraints

Version 1 intentionally excludes

* Workflow Engine
* BPMN
* Saga
* Human Approval
* Rule Engine
* AI Planning
* Dynamic Execution Planning
* Parallel Execution

The platform focuses on reliable business orchestration with a simple execution model.

---

# 16. Guiding Principles

Clients express Business Intent.

The platform creates an Execution.

The Orchestration Engine coordinates Execution Steps.

The Transformation Engine translates data.

The Connector Runtime communicates with external systems.

The Event Platform synchronizes platform components.

Every component has one clear responsibility.

Architecture decisions should prioritize simplicity, maintainability, and long-term extensibility.
