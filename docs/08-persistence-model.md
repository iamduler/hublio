# 08 - Persistence Model

> Product: Hublio
> Version: 1.0
> Status: Architecture Freeze v1

---

# 1. Purpose

This document defines what data Hublio persists.

It describes persistence responsibilities and data categories.

It does **not** define

* Database schema
* Table structure
* Indexes
* Relationships
* Storage implementation

Those topics belong to the ERD.

---

# 2. Design Principles

Persistence follows these principles.

* Platform First
* Runtime Aware
* Immutable History
* Audit Friendly
* Replay Friendly
* Multi-tenant
* Provider Agnostic

Persistence exists to support platform execution.

Not to mirror external systems.

---

# 3. Data Categories

Hublio persists two categories of data.

## Configuration Data

Long-lived platform configuration.

Usually created and managed through the Dashboard.

Examples

* Organization
* Workspace
* Connector
* Connection
* Credentials
* API Keys
* Settings

---

## Runtime Data

Data created while executing Business Intents.

Examples

* Intent
* Execution
* Execution Step
* Events
* Snapshots
* Audit Records

Runtime data is append-oriented.

---

# 4. Configuration Data

Configuration Data defines how the platform operates.

Typical characteristics

* Mutable
* CRUD Operations
* Versioned where appropriate
* Rarely changes

Configuration data is shared by future Executions.

---

# 5. Runtime Data

Runtime Data represents platform activity.

Typical characteristics

* Immutable after completion
* Timestamped
* Traceable
* Auditable

Runtime data should never overwrite history.

---

# 6. Intent

Intent is persisted after successful validation.

Intent contains

* Intent ID
* Organization
* Workspace
* Resource
* Operation
* Connection
* Payload
* Status
* Created Time

Intent becomes immutable after acceptance.

---

# 7. Execution

Execution is created from an accepted Intent.

Execution contains

* Execution ID
* Intent ID
* Status
* Context
* Started Time
* Completed Time
* Result

Execution owns the runtime lifecycle.

---

# 8. Execution Steps

Each Execution consists of one or more Steps.

Each Step records

* Step Name
* Status
* Started Time
* Completed Time
* Duration
* Retry Count
* Error (if any)

Steps belong to exactly one Execution.

---

# 9. Snapshots

Snapshots preserve execution data.

Examples

* Canonical Request
* Provider Request
* Provider Response
* Canonical Response

Snapshots support

* Replay
* Audit
* Troubleshooting

Snapshots are immutable.

---

# 10. Events

Events record significant platform activities.

Version 1 defines

* Runtime Events
* Business Events
* System Events

Events are immutable and append-only.

---

# 11. Audit Records

Audit Records capture security-sensitive actions.

Examples

* Login
* Credential Rotation
* Connection Changes
* API Key Creation
* Execution Replay

Audit history must be immutable.

---

# 12. Business Data

Customer business data is **not** the responsibility of Hublio.

Hublio may temporarily process

* Invoice
* Customer
* Product
* Order

These records are used only for execution.

Hublio is not the system of record for customer business entities.

---

# 13. Multi-tenancy

Every persisted record belongs to

* One Organization
* One Workspace

Tenant isolation must be enforced at every layer.

---

# 14. Data Retention

Configuration Data

Retained until deleted or archived.

Runtime Data

Retained according to platform retention policies.

Retention periods should be configurable.

---

# 15. Data Integrity

Configuration Data

Supports updates.

Runtime Data

Never mutates after completion.

Historical accuracy has priority over storage optimization.

---

# 16. Storage Responsibilities

PostgreSQL

Stores

* Configuration Data
* Runtime Data
* Audit Records

Redis

Stores

* Cache
* Distributed Locks
* Queue State
* Temporary Runtime Data

Object Storage

Stores

* Attachments
* Large Payloads
* Snapshot Files
* Export Files

Each storage technology has one clear responsibility.

---

# 17. Backup Strategy

The platform backs up

* Configuration Data
* Runtime Data
* Audit Records
* Attachments

Backup procedures should be automated.

Restore procedures should be regularly validated.

---

# 18. Scalability

Configuration Data grows slowly.

Runtime Data grows continuously.

The persistence layer should support

* Efficient querying
* Archiving
* Retention
* Horizontal application scaling

Storage optimization should not compromise auditability.

---

# 19. Version 1 Constraints

Version 1 intentionally excludes

* Event Sourcing
* CQRS
* Multi-region Replication
* Cross-database Sharding
* Polyglot Persistence

A single relational database is sufficient for Version 1.

---

# 20. Guiding Principles

Persist only what the platform owns.

Separate configuration from runtime.

Preserve execution history.

Never overwrite runtime history.

Keep persistence simple, reliable, and auditable.
