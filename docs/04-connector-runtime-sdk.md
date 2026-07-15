# Connector Runtime SDK

> Product: Hublio
> Version: 1.0
> Status: Architecture Freeze v1

---

# 1. Purpose

The Connector Runtime SDK defines the contract between Hublio and external systems.

Every integration with an external system is implemented as a Connector.

The SDK ensures that all Connectors behave consistently regardless of the provider.

---

# 2. Design Principles

The Connector Runtime follows these principles.

* Adapter Pattern
* Stateless
* Connector Agnostic
* Versioned
* Testable
* Observable
* Replaceable

A Connector is an implementation detail.

Business orchestration always belongs to the platform.

---

# 3. Responsibilities

A Connector is responsible for

* Provider Authentication
* API Communication
* Webhook Verification
* Canonical → Provider DTO Mapping
* Provider Response → Canonical Mapping
* Provider Error Translation
* Health Check

Connectors own all Provider DTO mapping.

Transformation Engine never maps Provider DTOs.

A Connector is NOT responsible for

* Business Rules
* Execution Lifecycle
* Retry
* Scheduling
* Queue Management
* Authorization
* Persistence
* Observability

These responsibilities belong to the platform.

---

# 4. Connector Lifecycle

Every Connector follows the same lifecycle.

Registered

↓

Configured

↓

Verified

↓

Enabled

↓

Disabled

↓

Removed

Only enabled Connectors may execute requests.

---

# 5. Connector Categories

Version 1 supports the following connector categories.

## Source Connector

Produces business data.

Examples

* POS
* ERP
* CRM
* E-commerce

---

## Destination Connector

Consumes business data.

Examples

* Electronic Invoice
* Accounting
* Shipping

---

## Bidirectional Connector

Supports both inbound and outbound communication.

Examples

* ERP
* CRM

---

## Utility Connector

Provides platform capabilities.

Examples

* Email
* SMS
* Object Storage

---

# 6. Connector Manifest

Every Connector exposes metadata.

Required metadata

* Connector ID
* Name
* Vendor
* Version
* Category
* Description

The platform uses this information for discovery and administration.

---

# 7. Supported Resources

A Connector declares the Canonical Resources it supports.

Examples

* Invoice
* Customer
* Company
* Product
* Order
* Payment
* Shipment

The platform never assumes unsupported resources.

---

# 8. Supported Operations

Each supported resource declares available operations.

Examples

Invoice

* Create
* Cancel
* Get

Customer

* Create
* Update
* Search

Order

* Create
* Update
* Get

Capabilities are explicitly declared.

Nothing is implied.

---

# 9. Authentication

Supported authentication methods include

* API Key
* OAuth2
* Bearer Token
* Basic Authentication
* JWT

Credentials are supplied by the platform.

Connectors never persist credentials.

---

# 10. Runtime Context

Each Connector receives a read-only Runtime Context.

The context contains

* Organization
* Workspace
* Connection
* Credentials
* Configuration
* Execution ID
* Correlation ID
* Timeout

The Runtime Context is immutable during execution.

---

# 11. Data Flow

Every Connector follows the same communication flow.

```text
Canonical Request
        │
        ▼
Connector Runtime
        │
        ▼
Provider Request
        │
        ▼
External System
        │
        ▼
Provider Response
        │
        ▼
Connector Runtime
        │
        ▼
Canonical Response
```

The platform never allows direct communication between Connectors.

---

# 12. Provider Mapping

Each Connector owns

* Provider DTOs
* Request Serialization
* Response Deserialization

The platform owns

* Canonical Models
* Business Processing

Provider-specific models must never enter the platform core.

---

# 13. Error Translation

Each Connector converts provider-specific errors into Canonical Errors.

Examples

* AuthenticationError
* ValidationError
* ConflictError
* TimeoutError
* RateLimitError
* NetworkError
* ProviderUnavailable
* InternalConnectorError

The rest of the platform never handles provider-specific errors.

---

# 14. Webhook Handling

A Connector is responsible for

* Signature Verification
* Payload Validation
* Provider Payload Parsing
* Canonical Event Conversion

Business processing belongs to the Orchestration Engine.

---

# 15. Health Check

Every Connector supports Health Check.

Possible states

* Healthy
* Degraded
* Unavailable

Health information is used by the platform for monitoring.

---

# 16. Versioning

Every Connector has its own version.

Connector upgrades should not require platform upgrades.

Backward compatibility should be maintained whenever practical.

---

# 17. Security

A Connector must never

* Store credentials
* Store secrets
* Log access tokens
* Persist provider payloads
* Bypass platform authorization

Sensitive information is always managed by the platform.

---

# 18. Testing

Every Connector should provide

* Unit Tests
* Integration Tests
* Sandbox Validation

Connector behavior should be deterministic.

---

# 19. Runtime Boundaries

The Platform owns

* Intent Processing
* Execution Lifecycle
* Retry
* Replay
* Scheduling
* Events
* Security
* Observability

The Connector owns

* Provider Authentication
* Provider Communication
* Canonical ↔ Provider DTO Mapping
* Provider Error Translation
* Webhook Verification

Responsibilities must remain clearly separated.

---

# 20. Design Constraints

Version 1 intentionally excludes

* Dynamic Connector Loading
* Remote Connector Execution
* Connector Marketplace
* Connector Hot Reload
* Independent Connector Deployment

All Connectors are compiled into the platform application.

This keeps deployment, debugging, and operations simple.

---

# 21. Future Evolution

Future versions may support

* Connector Marketplace
* Remote Connector Runtime
* Connector Sandboxing
* Independent Connector Deployment

These capabilities must not change the SDK contract defined in this document.

---

# 22. Guiding Principles

A Connector is an Adapter.

It translates between Hublio and one external system.

It owns provider-specific implementation.

It never owns platform behavior.

It never owns business rules.

It never owns execution.

The Connector Runtime should remain simple, deterministic, and easily replaceable.
