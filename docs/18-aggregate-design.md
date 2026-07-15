# 18 - Aggregate Design

> Product: Hublio
> Version: 1.0
> Status: Architecture Freeze v1

---

# 1. Purpose

This document defines the Aggregates of Hublio.

Each Aggregate represents a consistency boundary within the Domain.

This document describes

* Responsibilities
* Entities
* Value Objects
* Invariants
* Behaviors
* Published Events
* Persistence

It does not define

* Database tables
* Go packages
* REST APIs

---

# 2. Design Principles

Aggregates follow these principles.

* One Aggregate = One Consistency Boundary
* Small Aggregate
* Explicit Invariants
* Rich Behavior
* Low Coupling
* Event Driven

Every Aggregate owns its own state.

---

# 3. Aggregate Overview

Version 1 defines six Aggregates.

| Aggregate    | Domain        |
| ------------ | ------------- |
| Organization | Identity      |
| Workspace    | Identity      |
| Connector    | Integration   |
| Connection   | Integration   |
| Intent       | Orchestration |
| Execution    | Orchestration |

These Aggregates form the core Domain Model.

---

# 4. Organization Aggregate

## Responsibility

Represents a tenant.

Owns tenant-level configuration.

---

### Root Entity

Organization

---

### Child Entities

* User

---

### Value Objects

* Organization Name
* Status

---

### Invariants

* Organization ID is immutable.
* Suspended organizations cannot submit new Intents.
* Deleted organizations cannot access the platform.

---

### Behaviors

* Create
* Update
* Suspend
* Activate
* Archive

---

### Published Events

* OrganizationCreated
* OrganizationUpdated
* OrganizationSuspended
* OrganizationActivated

---

### Persisted Data

* Organization
* Users

---

# 5. Workspace Aggregate

## Responsibility

Provides logical isolation within an Organization.

Owns Workspace-scoped access credentials.

---

### Root Entity

Workspace

---

### Child Entities

* API Key

---

### Value Objects

* Workspace Name
* Environment

---

### Invariants

* Workspace belongs to exactly one Organization.
* Disabled Workspaces cannot execute Intents.
* API Keys belong to exactly one Workspace.

---

### Behaviors

* Create
* Update
* Enable
* Disable
* Create API Key
* Disable API Key
* Rotate API Key

---

### Published Events

* WorkspaceCreated
* WorkspaceEnabled
* WorkspaceDisabled
* ApiKeyCreated
* ApiKeyDisabled
* ApiKeyRotated

---

### Persisted Data

* Workspace
* API Keys

---

# 6. Connector Aggregate

## Responsibility

Represents an integration implementation.

---

### Root Entity

Connector

---

### Child Entities

None

---

### Value Objects

* Version
* Vendor
* Category
* Supported Resources
* Supported Operations

---

### Invariants

* Connector ID is immutable.
* Removed Connectors cannot be enabled.
* Disabled Connectors cannot be used by new Executions.

---

### Behaviors

* Register
* Enable
* Disable
* Remove

---

### Published Events

* ConnectorRegistered
* ConnectorEnabled
* ConnectorDisabled
* ConnectorRemoved

---

### Persisted Data

* Connector Metadata

---

# 7. Connection Aggregate

## Responsibility

Represents a configured integration between Hublio and an external system.

---

### Root Entity

Connection

---

### Child Entities

* Credential

---

### Value Objects

* Configuration
* Retry Policy
* Timeout
* Connection Status

---

### Invariants

* Connection belongs to one Workspace.
* Only verified Connections can become Active.
* Disabled Connections cannot execute Intents.

---

### Behaviors

* Configure
* Verify
* Activate
* Disable
* Rotate Credentials

---

### Published Events

* ConnectionCreated
* ConnectionVerified
* ConnectionActivated
* ConnectionDisabled
* CredentialRotated

---

### Persisted Data

* Connection
* Credential

---

# 8. Intent Aggregate

## Responsibility

Represents a business request submitted by a client.

---

### Root Entity

Intent

---

### Child Entities

None

---

### Value Objects

* Resource
* Operation
* Payload
* Intent Status

---

### Invariants

* Accepted Intents are immutable.
* Rejected Intents never create Executions.
* Every accepted Intent creates exactly one Execution.

---

### Behaviors

* Submit
* Accept
* Reject
* Expire

---

### Published Events

* IntentSubmitted
* IntentAccepted
* IntentRejected
* IntentExpired

---

### Persisted Data

* Intent

---

# 9. Execution Aggregate

## Responsibility

Represents runtime processing of an accepted Intent.

---

### Root Entity

Execution

---

### Child Entities

* Execution Step
* Execution Snapshot
* Execution Timeline

---

### Value Objects

* Execution Context
* Execution Result
* Retry Policy
* Timeout

---

### Invariants

* Execution belongs to exactly one Intent.
* Terminal Executions are immutable.
* Steps execute sequentially.
* Every Execution produces one Result.

---

### Behaviors

* Start
* Execute Step
* Retry
* Cancel
* Succeed
* Fail
* Replay

---

### Published Events

* ExecutionCreated
* ExecutionStarted
* StepStarted
* StepCompleted
* ExecutionSucceeded
* ExecutionFailed

---

### Persisted Data

* Execution
* Execution Steps
* Snapshots
* Timeline

---

# 10. Aggregate Relationships

```text id="hf9v21"
Organization
      │
      ▼
Workspace
      │
      ▼
Connection
      │
      ▼
Intent
      │
      ▼
Execution
```

Connector is referenced by Connection.

Aggregates communicate through identifiers and events.

---

# 11. Aggregate Communication

Aggregates never modify each other's state directly.

Communication occurs through

* Application Services
* Runtime Events

Cross-aggregate consistency is eventually consistent.

---

# 12. Aggregate Persistence

Each Aggregate owns its own persistence.

No Aggregate may update another Aggregate's data directly.

Persistence implementation is an infrastructure concern.

---

# 13. Version 1 Constraints

Version 1 intentionally excludes

* Aggregate inheritance
* Nested Aggregates
* Cross-Aggregate transactions
* Event Sourcing
* CQRS

Aggregates remain small and focused.

---

# 14. Guiding Principles

Aggregates represent business consistency.

Each Aggregate owns its state.

Each Aggregate protects its invariants.

Behavior is more important than data.

The Domain Model remains independent of infrastructure.
