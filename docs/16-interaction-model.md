# 16 - Interaction Model

> Product: Hublio
> Version: 1.0
> Status: Architecture Freeze v1

---

# 1. Purpose

This document defines how the core platform components collaborate to process business requests.

It describes

* Component interactions
* Responsibility boundaries
* Runtime flow
* Event propagation

It does not describe

* HTTP implementation
* Database operations
* Queue implementation
* Programming language

---

# 2. Design Principles

Component interactions follow these principles.

* Intent Driven
* Single Responsibility
* Loose Coupling
* Provider Agnostic
* Event Driven
* Observable

Each component performs one responsibility before passing control to the next component.

---

# 3. Primary Interaction

Every business request follows the same interaction pattern.

```text
Client
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
    ├──────────────┬──────────────┐
    ▼              ▼              ▼
Transformation   Event        Connector Runtime
Engine          Platform
    │                             │
    ▼                             ▼
Canonical Model            External System
```

This interaction model applies to every supported business operation.

---

# 4. Business Request Flow

The client submits a Business Intent.

```text
Client

↓

Platform API

↓

Validate Request

↓

Intent Processor

↓

Create Intent

↓

Create Execution

↓

Orchestration Engine
```

The Platform API validates requests.

The Intent Processor accepts business intent.

The Orchestration Engine owns execution.

---

# 5. Execution Flow

The Orchestration Engine coordinates every execution.

```text
Execution

↓

Validate Step

↓

Transform Request

↓

Connector Runtime

↓

Transform Response

↓

Complete Execution
```

Each Execution Step performs one responsibility.

Version 1 executes steps sequentially.

---

# 6. Transformation Flow

Business logic never communicates with provider models.

```text
Canonical Request

↓

Transformation Engine

↓

Provider Request

↓

Connector Runtime

↓

Provider Response

↓

Transformation Engine

↓

Canonical Response
```

Canonical Models are the only models shared across the platform.

---

# 7. Connector Flow

Every external communication follows the same pattern.

```text
Execution Step

↓

Connector Runtime

↓

Provider API

↓

Provider Response

↓

Canonical Response
```

The Connector Runtime is the only component allowed to communicate with external systems.

---

# 8. Event Flow

Platform components communicate through events.

```text
Execution

↓

Runtime Event

↓

Event Platform

↓

Subscribers

├── Audit

├── Metrics

├── Notification

└── Logging
```

Publishers never depend on subscribers.

Subscribers never modify execution state.

---

# 9. Retry Flow

Retry is coordinated by the Orchestration Engine.

```text
Execution Failed

↓

Retry Policy

↓

Queue

↓

Worker

↓

Execution Restarted
```

Connectors never retry automatically.

---

# 10. Replay Flow

Replay creates a new Execution.

```text
Replay Request

↓

Load Intent

↓

Load Snapshots

↓

Create Execution

↓

Run Execution
```

Historical Executions remain unchanged.

---

# 11. Scheduled Execution Flow

Scheduled execution follows the same runtime model.

```text
Scheduler

↓

Create Intent

↓

Intent Processor

↓

Execution

↓

Normal Processing
```

The Scheduler behaves like any other client.

---

# 12. Webhook Flow

Incoming webhooks follow a standardized process.

```text
Provider

↓

Connector Runtime

↓

Verify Signature

↓

Transform Payload

↓

Create Intent

↓

Execution
```

Webhook processing reuses the same runtime model as API requests.

---

# 13. Responsibility Matrix

| Component             | Primary Responsibility                        |
| --------------------- | --------------------------------------------- |
| Platform API          | Public API, authentication, validation        |
| Intent Processor      | Accept business intent                        |
| Orchestration Engine  | Manage execution lifecycle                    |
| Transformation Engine | Convert between canonical and provider models |
| Connector Runtime     | Communicate with external systems             |
| Event Platform        | Publish and distribute events                 |

Each responsibility belongs to exactly one component.

---

# 14. State Ownership

Platform API

* Request validation

Intent Processor

* Intent lifecycle

Orchestration Engine

* Execution lifecycle

Transformation Engine

* Data transformation

Connector Runtime

* Provider communication

Event Platform

* Event distribution

State ownership is never shared.

---

# 15. Interaction Boundaries

Components interact only through public interfaces.

Direct access to another component's internal state is prohibited.

Components should exchange

* Canonical Models
* Runtime Events
* Execution Context

Provider-specific models remain inside the Connector Runtime.

---

# 16. Version 1 Constraints

Version 1 intentionally excludes

* Parallel execution
* Workflow branching
* Human approval
* BPMN
* Dynamic routing
* Rule Engine

Every business request follows a single deterministic execution path.

---

# 17. Guiding Principles

Every client submits a Business Intent.

The platform creates an Execution.

The Orchestration Engine coordinates Execution Steps.

Specialized components perform specialized work.

The Event Platform synchronizes the platform.

Every interaction is deterministic, observable, and provider agnostic.
