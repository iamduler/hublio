# 19 - Data Model

> Product: Hublio
> Version: 1.0
> Status: Architecture Freeze v1

---

# 1. Purpose

This document defines the persistence model of Hublio.

It bridges the gap between the Domain Model and the physical database schema.

It defines

* Persistence Objects
* Ownership
* Relationships
* Cardinality
* Data Boundaries

It does **not** define

* SQL types
* Primary Keys
* Foreign Keys
* Indexes
* Constraints
* Database-specific features

Those are defined in the Database Schema (DBML).

---

# 2. Design Principles

The persistence model follows these principles.

* Domain Driven
* Aggregate First
* Normalized
* Provider Agnostic
* Immutable Runtime History
* Multi-tenant

The database exists to persist the Domain.

The Domain never follows the database.

---

# 3. Persistence Categories

Version 1 defines two categories of persistent data.

## Configuration Data

Configuration Data changes infrequently.

Examples

* Organization
* Workspace
* Connector
* Connection
* Credential
* API Key

Configuration Data is mutable.

---

## Runtime Data

Runtime Data is created while processing Business Intents.

Examples

* Intent
* Execution
* Execution Step
* Execution Snapshot
* Execution Timeline
* Event
* Audit Log

Runtime Data is append-oriented.

Completed runtime records are immutable.

---

# 4. Aggregate Mapping

| Aggregate    | Persistence Objects                                               |
| ------------ | ----------------------------------------------------------------- |
| Organization | Organization, User, API Key                                       |
| Workspace    | Workspace                                                         |
| Connector    | Connector                                                         |
| Connection   | Connection, Credential                                            |
| Intent       | Intent                                                            |
| Execution    | Execution, Execution Step, Execution Snapshot, Execution Timeline |

Each Aggregate owns its persistence objects.

---

# 5. Configuration Model

## Organization

Represents a tenant.

Owns

* Users
* API Keys
* Workspaces

---

## Workspace

Represents an isolated working environment.

Belongs to one Organization.

Owns

* Connections

---

## Connector

Represents an integration implementation.

Contains

* Metadata
* Supported Resources
* Supported Operations

Referenced by Connections.

---

## Connection

Represents a configured integration.

Contains

* Configuration
* Status
* Retry Policy
* Timeout

Owns

* Credentials

---

## Credential

Represents authentication data used by a Connection.

Credentials are encrypted.

Credentials are never exposed through public APIs.

---

# 6. Runtime Model

## Intent

Represents a Business Intent submitted by a client.

Contains

* Resource
* Operation
* Payload
* Status

Every accepted Intent creates one Execution.

---

## Execution

Represents one runtime instance.

Contains

* Status
* Context
* Result

Owns

* Execution Steps
* Snapshots
* Timeline

Execution history is immutable after completion.

---

## Execution Step

Represents one execution activity.

Contains

* Name
* Status
* Duration
* Retry Count
* Error Information

Belongs to exactly one Execution.

---

## Execution Snapshot

Stores important runtime payloads.

Examples

* Canonical Request
* Provider Request
* Provider Response
* Canonical Response

Snapshots support replay and troubleshooting.

---

## Execution Timeline

Records execution progress.

Examples

* Created
* Started
* Step Completed
* Retry Scheduled
* Completed

Timeline entries are append-only.

---

# 7. Event Model

Events are stored independently of Executions.

Supported categories

* Runtime Event
* Business Event
* System Event

Events are immutable.

Events are append-only.

---

# 8. Audit Model

Audit Logs record security-sensitive platform activities.

Examples

* Login
* API Key Creation
* Connection Update
* Credential Rotation
* Execution Replay

Audit Logs are immutable.

---

# 9. Relationships

```text
Organization
│
├── User
├── API Key
└── Workspace
      │
      └── Connection
              │
              ├── Credential
              └── Intent
                     │
                     └── Execution
                            ├── Execution Step
                            ├── Execution Snapshot
                            └── Execution Timeline

Connector
    │
    └── Connection
```

Events and Audit Logs are independent persistence objects.

---

# 10. Cardinality

Organization

* Owns many Users
* Owns many API Keys
* Owns many Workspaces

Workspace

* Owns many Connections

Connector

* Is referenced by many Connections

Connection

* Owns one or more Credentials
* Owns many Intents

Intent

* Creates exactly one Execution

Execution

* Owns many Execution Steps
* Owns many Execution Snapshots
* Owns many Timeline Entries

---

# 11. Tenant Isolation

Every business object belongs to exactly one Organization.

Workspace provides logical isolation within an Organization.

Tenant boundaries are enforced by the application layer.

---

# 12. Runtime History

The following objects are immutable after creation or completion.

* Intent (after Accepted)
* Execution (after Completed)
* Execution Snapshot
* Execution Timeline
* Event
* Audit Log

Historical records are never updated in place.

---

# 13. External Data Ownership

Hublio is not the system of record for customer business entities.

External systems own

* Invoices
* Customers
* Products
* Orders
* Payments

Hublio stores only the runtime information required for execution, replay, auditing, and troubleshooting.

---

# 14. Storage Strategy

Configuration Data

* Long-lived
* Frequently queried
* Occasionally updated

Runtime Data

* High volume
* Append-oriented
* Frequently written
* Rarely updated

Storage optimization should preserve auditability and replayability.

---

# 15. Evolution Strategy

The Data Model is stable.

The Database Schema may evolve.

Schema changes should preserve compatibility with the Data Model whenever practical.

---

# 16. Guiding Principles

The Data Model is derived from the Domain Model.

Aggregates own persistence.

Runtime history is immutable.

Provider-specific data remains outside the platform core.

The Database Schema is an implementation of this Data Model, not the other way around.
