# 09 - Platform API Specification

> Product: Hublio
> Version: 1.0
> Status: Architecture Freeze v1

---

# 1. Purpose

This document defines the public API architecture of Hublio.

The Platform API exposes business capabilities.

It never exposes provider-specific APIs.

Clients communicate only with Hublio.

---

# 2. Design Principles

The Platform API follows these principles.

* API First
* Intent Driven
* Canonical Models
* Provider Agnostic
* Versioned
* Secure
* Observable

Business capabilities remain stable even if connectors change.

---

# 3. API Categories

Hublio exposes two API groups.

## Management API

Used by

* Dashboard
* Administrators
* Internal Platform

Provides CRUD operations for platform configuration.

Examples

* Organizations
* Workspaces
* Connectors
* Connections
* Credentials
* API Keys
* Settings

---

## Platform API

Used by

* POS
* ERP
* CRM
* E-commerce
* Third-party Applications

Provides business operations through Business Intents.

---

# 4. Platform Entry Point

Every business request enters the platform as an Intent.

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
Execution
```

The client never creates Executions directly.

---

# 5. Intent API

Business operations are submitted through the Intent API.

Primary operations

* Create Intent
* Get Intent
* List Intents

Intent represents the public contract of the platform.

---

# 6. Execution API

Execution is an internal runtime concept.

Clients may query Execution status.

Supported operations

* Get Execution
* List Executions

Execution creation is internal.

Execution deletion is not supported.

---

# 7. Canonical Resources

Every Intent targets a Canonical Resource.

Examples

* Invoice
* Customer
* Company
* Product
* Order
* Payment
* Shipment

Resources are platform concepts.

They are not provider-specific.

---

# 8. Canonical Operations

Version 1 supports common operations.

Examples

* Create
* Update
* Cancel
* Get
* Search
* Sync

Operations are interpreted by the platform.

---

# 9. Request Model

Every Intent request contains

* Resource
* Operation
* Connection
* Payload

Optional metadata

* Idempotency Key
* Correlation ID

Payload always follows the Canonical Data Model.

---

# 10. Response Model

Every successful request returns

* Intent ID
* Execution ID
* Status
* Accepted Time

Long-running work continues asynchronously.

---

# 11. Asynchronous Processing

Business execution is asynchronous by default.

Client

↓

Intent Accepted

↓

Execution Created

↓

Execution Succeeded

↓

Result Available

Clients may

* Poll
* Receive Webhooks
* Query Execution

---

# 12. Authentication

Supported authentication methods

* API Key
* OAuth2
* JWT

Authentication identifies the caller.

---

# 13. Authorization

Authorization is evaluated using

* Organization
* Workspace
* Permissions

Authorization never depends on the selected provider.

---

# 14. Idempotency

All write operations support idempotency.

Repeated requests with the same idempotency key must not create duplicate business operations.

---

# 15. Versioning

The API uses URI versioning.

Example

/api/v1

Breaking changes require a new major version.

---

# 16. Pagination

Collection endpoints use cursor-based pagination.

Offset pagination is intentionally avoided.

---

# 17. Error Model

The platform returns Canonical Errors.

Examples

* ValidationError
* AuthenticationError
* AuthorizationError
* ConflictError
* TimeoutError
* RateLimitError
* InternalError

Provider-specific errors are translated before leaving the platform.

---

# 18. Observability

Every API request should expose

* Request ID
* Correlation ID
* Intent ID
* Execution ID (if available)

These identifiers support end-to-end tracing.

---

# 19. Platform Boundaries

The Platform API owns

* Request Validation
* Authentication
* Authorization
* Intent Submission

The Platform API does not own

* Business Execution
* Data Transformation
* Provider Communication

These responsibilities belong to downstream platform components.

---

# 20. Version 1 Constraints

Version 1 intentionally excludes

* GraphQL
* gRPC
* Public Event Streaming API
* Bulk Execution API
* Advanced Orchestration APIs (future)

REST is the only public API protocol in Version 1.

---

# 21. Guiding Principles

Clients express business intent.

The Platform API accepts the request.

The platform creates one or more Executions.

The Orchestration Engine coordinates execution.

The client interacts with stable platform concepts, never with provider-specific APIs.
