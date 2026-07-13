# 00 - Architecture Principles

> Product: Hublio
> Version: 1.0
> Status: Architecture Freeze v1

---

# 1. Purpose

This document defines the architectural principles of Hublio.

It is the foundation for all technical decisions, documentation, and source code.

If any future document conflicts with this document, this document takes precedence.

---

# 2. Product Vision

Hublio is a Business Orchestration Platform.

Its purpose is to orchestrate business operations between software systems using a unified execution model.

Hublio is not

* an ERP
* a CRM
* an Electronic Invoice Provider
* a Workflow Engine

Hublio coordinates communication between systems.

---

# 3. Core Principles

The platform follows these principles.

* API First
* Canonical Data Model
* Intent Driven
* Connector Agnostic
* Event Driven
* Stateless Services
* Security by Design
* Observability by Default
* Simplicity over Complexity

---

# 4. Architecture Overview

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
- Observability
```

---

# 5. Runtime Model

Every business request follows the same runtime model.

```text
Intent
    │
    ▼
Execution
    │
    ▼
Execution Step
```

This runtime model is the foundation of Hublio.

---

# 6. Core Components

The platform consists of six core components.

* Platform API
* Intent Processor
* Orchestration Engine
* Transformation Engine
* Connector Runtime
* Event Platform

Each component has one clear responsibility.

---

# 7. Canonical Data Model

All business processing uses Canonical Models.

Provider-specific models remain inside Connector Runtime.

Business logic must never depend on external APIs.

---

# 8. Connector Philosophy

A Connector is an Adapter.

A Connector owns

* Provider Authentication
* Provider APIs
* Provider DTOs
* Webhook Verification

The platform owns

* Execution
* Retry
* Scheduling
* Events
* Security
* Observability

---

# 9. Runtime Philosophy

Clients submit Business Intents.

The platform creates an Execution.

The Orchestration Engine coordinates Execution Steps.

Execution remains an internal runtime concept.

Clients never interact directly with Executions.

---

# 10. Platform Boundaries

Hublio owns

* Business Orchestration
* Execution Lifecycle
* Canonical Models
* Connector Management
* Platform Configuration

External systems own

* Business Data
* Business Rules
* Provider APIs

Hublio is not the system of record for customer business data.

---

# 11. Official Vocabulary

Platform

* Platform API
* Intent Processor
* Orchestration Engine
* Transformation Engine
* Connector Runtime
* Event Platform

Runtime

* Intent
* Execution
* Execution Step
* Execution Context
* Execution Timeline
* Execution Snapshot
* Execution Result

Integration

* Connector
* Connection
* Canonical Model

Multi-tenancy

* Organization
* Workspace

Only these terms should be used throughout the project.

---

# 12. Out of Scope (Version 1)

The following capabilities are intentionally excluded.

* Workflow Engine
* BPMN
* Saga
* Human Approval
* Rule Engine
* AI Planning
* Dynamic Execution Planning
* Parallel Execution
* Connector Marketplace

These features may be introduced in future versions without changing the core architecture.

---

# 13. Architecture Freeze

Architecture Freeze v1 defines the stable foundation of Hublio.

The following concepts must not change without a major architecture review.

* Business Intent
* Execution
* Execution Step
* Canonical Data Model
* Connector Runtime
* Platform API
* Orchestration Engine
* Transformation Engine
* Event Platform

All future documents and implementations must follow these principles.

---

# 14. Guiding Principle

Keep the platform simple.

Prefer explicit responsibilities.

Prefer composition over complexity.

Optimize for maintainability, reliability, and long-term evolution.
