# 14 - State Machines

> Product: Hublio
> Version: 1.0
> Status: Architecture Freeze v1

---

# 1. Purpose

This document defines the lifecycle of domain objects that change state over time.

State Machines ensure that every state transition is explicit, predictable, and validated.

Only Aggregates with meaningful business lifecycles have State Machines.

---

# 2. Design Principles

State Machines follow these principles.

* Explicit States
* Explicit Transitions
* Immutable History
* Deterministic Behavior
* Business Rule Enforcement

A state transition represents a business decision.

State changes must never occur implicitly.

---

# 3. Aggregates with State Machines

Version 1 defines State Machines for

* Intent
* Execution
* Connection
* Connector

Other Aggregates do not require lifecycle management.

---

# 4. Intent State Machine

Intent represents a business request submitted by a client.

## States

```text
Submitted
    │
    ▼
Accepted
    │
    ├──────────────┐
    ▼              ▼
Rejected      Expired
```

### State Definitions

**Submitted**

* Intent has been received.
* Validation has not completed.

**Accepted**

* Validation succeeded.
* Execution has been created.

**Rejected**

* Validation failed.
* No Execution is created.

**Expired**

* Intent was not processed within the configured lifetime.

### Allowed Transitions

* Submitted → Accepted
* Submitted → Rejected
* Submitted → Expired

Accepted is immutable.

Rejected is terminal.

Expired is terminal.

---

# 5. Execution State Machine

Execution represents runtime processing.

## States

```text
Created
    │
    ▼
Queued
    │
    ▼
Running
    │
 ┌──┴───────────────┐
 ▼                  ▼
Succeeded        Failed

Alternative terminal states

Cancelled
Expired
Dead Letter
```

### State Definitions

**Created**

Execution has been initialized.

**Queued**

Execution is waiting for a Worker.

**Running**

Execution Steps are being processed.

**Succeeded**

All Execution Steps completed successfully. This is the success terminal state.

**Failed**

Execution stopped because of an unrecoverable error.

**Cancelled**

Execution was cancelled by a user or the platform.

**Expired**

Execution exceeded its lifetime.

**Dead Letter**

Execution exhausted all retry attempts.

### Allowed Transitions

* Created → Queued
* Queued → Running
* Running → Succeeded
* Running → Failed
* Running → Cancelled
* Running → Expired
* Failed → Queued (Retry)
* Failed → Dead Letter

Succeeded, Cancelled, Expired and Dead Letter are terminal states.

---

# 6. Connection State Machine

A Connection represents a configured integration between Hublio and an external system.

## States

```text
Draft
    │
    ▼
Verifying
    │
 ┌──┴─────────────┐
 ▼                ▼
Active         Verification Failed
 │
 ▼
Disabled
```

### State Definitions

**Draft**

Configuration is incomplete.

**Verifying**

The platform is validating credentials and connectivity.

**Active**

The Connection is available for use.

**Verification Failed**

Verification was unsuccessful.

**Disabled**

The Connection is intentionally disabled.

### Allowed Transitions

* Draft → Verifying
* Verifying → Active
* Verifying → Verification Failed
* Verification Failed → Verifying
* Active → Disabled
* Disabled → Active

Only Active Connections may be used for Executions.

---

# 7. Connector State Machine

A Connector represents an integration implementation.

## States

```text
Registered
      │
      ▼
Enabled
      │
      ▼
Disabled
      │
      ▼
Removed
```

### State Definitions

**Registered**

Connector is installed but not enabled.

**Enabled**

Connector is available for Connections.

**Disabled**

Connector is temporarily unavailable.

**Removed**

Connector is permanently removed from the platform.

### Allowed Transitions

* Registered → Enabled
* Enabled → Disabled
* Disabled → Enabled
* Disabled → Removed

Removed is terminal.

---

# 8. State Transition Rules

Every transition must

* Validate business rules
* Record timestamp
* Record actor (if applicable)
* Produce audit information
* Publish appropriate events

No transition may bypass validation.

---

# 9. Runtime Events

State changes produce Runtime Events.

Examples

* IntentAccepted
* ExecutionCreated
* ExecutionStarted
* ExecutionFailed
* ConnectionActivated
* ConnectionDisabled
* ConnectorEnabled

Events are published after successful state transitions.

---

# 10. Persistence

Current state is stored with the Aggregate.

Historical transitions are preserved through

* Timeline
* Runtime Events
* Audit Records

State history must never be overwritten.

---

# 11. Error Handling

Invalid transitions must be rejected.

Examples

* Completing an Execution twice
* Activating an unverified Connection
* Using a Disabled Connector

Business rules always take precedence.

---

# 12. Observability

Every state transition should expose

* Previous State
* Current State
* Timestamp
* Correlation ID
* Execution ID (if applicable)

Transitions must be traceable end-to-end.

---

# 13. Version 1 Constraints

Version 1 intentionally excludes

* Parallel State Machines
* Hierarchical States
* Composite States
* Dynamic State Definitions
* User-defined State Machines

All lifecycle definitions are static and implemented in code.

---

# 14. Guiding Principles

Every important lifecycle is represented by a State Machine.

States are explicit.

Transitions are validated.

History is immutable.

Business rules determine when state changes are allowed.

Simple state models are preferred over flexible state models.
