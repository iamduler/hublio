# 13 - Domain-Driven Design

> Product: Hublio
> Version: 1.0
> Status: Architecture Freeze v1

---

# 1. Purpose

This document defines the business domain model of Hublio.

The Domain Model represents business concepts independently of

* Database
* REST API
* Go packages
* Infrastructure
* External systems

The Domain is the heart of the platform.

---

# 2. Domain Philosophy

Hublio is a Business Orchestration Platform.

The Domain is responsible for

* Business Concepts
* Business Rules
* Business Consistency
* Business Language

Technology choices must never influence the Domain Model.

---

# 3. Design Principles

The Domain follows these principles.

* Business First
* Explicit Boundaries
* High Cohesion
* Low Coupling
* Rich Domain Model
* Simple Aggregates

Business behavior belongs inside the Domain.

Infrastructure supports the Domain.

---

# 4. Bounded Contexts

Version 1 defines the following Bounded Contexts.

## Identity

Responsibilities

* Organization
* Workspace
* User
* Authentication
* Authorization
* API Key

---

## Integration

Responsibilities

* Connector
* Connection
* Credentials
* Provider Integration

---

## Orchestration

Responsibilities

* Intent
* Execution
* Execution Step
* Retry
* Replay
* Scheduling

---

## Transformation

Responsibilities

* Canonical Model
* Mapping
* Validation
* Normalization

---

## Events

Responsibilities

* Runtime Events
* Business Events
* System Events

---

## Administration

Responsibilities

* Platform Configuration
* Monitoring
* Dashboard
* Maintenance

Each context owns its own business rules.

---

# 5. Ubiquitous Language

The following terms have fixed meanings.

## Intent

A business request submitted by a client.

---

## Execution

The runtime instance created from an accepted Intent.

---

## Execution Step

A single execution unit within an Execution.

---

## Connector

An adapter for one external system.

---

## Connection

A tenant-specific configuration of a Connector.

---

## Canonical Model

The provider-independent business model used inside Hublio.

---

## Runtime Event

A platform event describing execution progress.

---

## Business Event

A platform event describing a completed business outcome.

---

## System Event

A platform event describing platform administration.

Only these terms should be used throughout the project.

---

# 6. Aggregates

Version 1 defines six Aggregates.

* Organization
* Workspace
* Connector
* Connection
* Intent
* Execution

Every Aggregate owns its own consistency boundary.

Aggregates should remain small and focused.

---

# 7. Aggregate Responsibilities

## Organization

Owns tenant-level configuration.

---

## Workspace

Owns environment-level isolation.

---

## Connector

Owns provider capabilities.

---

## Connection

Owns provider configuration and credentials.

---

## Intent

Owns the public business request.

---

## Execution

Owns runtime state and execution lifecycle.

Each Aggregate has one clear responsibility.

---

# 8. Entities

Entities have stable identity.

Examples

* User
* Credential
* Execution Step

Entity identity never changes.

State may change according to business rules.

---

# 9. Value Objects

Value Objects have no identity.

Examples

* Email
* Money
* Timeout
* Retry Policy
* Correlation ID
* Trace ID
* Execution Result

Equal Value Objects are interchangeable.

---

# 10. Domain Events

Aggregates publish Domain Events after successful state changes.

Examples

* IntentAccepted
* ExecutionStarted
* ExecutionCompleted
* ConnectionActivated

Domain Events represent facts.

They never represent commands.

---

# 11. Business Invariants

Each Aggregate protects its own invariants.

Examples

Intent

* Cannot be modified after acceptance.

Execution

* Cannot complete twice.
* Cannot restart after completion.

Connection

* Cannot become Active before verification.

Business invariants belong inside Aggregates.

---

# 12. Aggregate Communication

Aggregates communicate through

* Application Services
* Domain Events

Aggregates should not reference each other directly.

Cross-aggregate consistency is eventually consistent.

---

# 13. Transaction Boundary

One transaction should modify one Aggregate.

Long-running business operations are coordinated by the Orchestration Engine.

Distributed transactions are intentionally avoided.

---

# 14. Infrastructure Boundary

Infrastructure implements technical concerns.

Examples

* Database
* Redis
* HTTP
* Queue
* Provider APIs

Infrastructure never contains business rules.

---

# 15. Persistence Boundary

Persistence stores Domain state.

Persistence does not define the Domain.

Database schema follows the Domain Model.

The Domain never follows the database.

---

# 16. Version 1 Constraints

Version 1 intentionally excludes

* Event Sourcing
* CQRS
* Specification Pattern
* Factory Pattern
* Domain Services as a default
* Complex Aggregate hierarchies

The Domain should remain simple and easy to understand.

---

# 17. Guiding Principles

The Domain is the source of truth.

Business rules belong inside the Domain.

Aggregates protect consistency.

Infrastructure supports the Domain.

Technology evolves.

Business concepts remain stable.
