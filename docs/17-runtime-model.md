# 17 - Runtime Model

> Product: Hublio
> Version: 1.0
> Status: Architecture Freeze v1

---

# 1. Purpose

This document defines how Hublio executes business operations at runtime.

The Runtime Model describes

* Runtime objects
* Runtime lifecycle
* Runtime responsibilities
* Runtime data flow

It is the execution model used by every business operation.

---

# 2. Runtime Philosophy

Every business request follows the same runtime model.

```text
Business Intent
        │
        ▼
    Execution
        │
        ▼
 Execution Step(s)
        │
        ▼
Execution Result
```

The runtime model is intentionally simple.

Every business capability is executed through the same lifecycle.

---

# 3. Runtime Components

Version 1 defines the following runtime components.

* Intent
* Execution
* Execution Step
* Execution Context
* Execution Timeline
* Execution Snapshot
* Execution Result

No additional runtime abstractions are introduced.

---

# 4. Intent

Intent represents a business request submitted by a client.

Intent is the public contract of the platform.

Typical examples

* Create Invoice
* Cancel Invoice
* Sync Customer
* Import Product

An Intent contains

* Intent ID
* Resource
* Operation
* Connection
* Payload
* Status
* Created Time

An accepted Intent is immutable.

---

# 5. Execution

Execution is the runtime instance created from an accepted Intent.

Execution is an internal platform object.

Clients never create Executions directly.

An Execution contains

* Execution ID
* Intent ID
* Status
* Context
* Started Time
* Completed Time
* Result

Execution owns the runtime lifecycle.

---

# 6. Execution Steps

Execution is composed of one or more Execution Steps.

Each Step performs exactly one responsibility.

Typical Steps

1. Validate Request
2. Transform Request
3. Invoke Connector
4. Transform Response
5. Publish Events

Steps execute sequentially in Version 1.

---

# 7. Execution Context

Execution Context provides immutable runtime information.

It contains

* Organization
* Workspace
* Connection
* Credentials
* Correlation ID
* Trace ID
* Configuration
* Timeout

The Context is created before execution begins.

It remains read-only throughout execution.

---

# 8. Execution Timeline

The Timeline records the execution lifecycle.

Typical entries

* Created
* Queued
* Started
* Step Started
* Step Completed
* Retry Scheduled
* Completed
* Failed
* Cancelled

Timeline entries are append-only.

---

# 9. Execution Snapshot

Snapshots preserve important runtime data.

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

# 10. Execution Result

Every Execution produces exactly one result.

Possible results

* Succeeded
* Failed
* Cancelled
* Expired
* Dead Letter

Execution Results never change after completion.

---

# 11. Runtime Flow

Every execution follows the same flow.

```text
Intent

↓

Execution Created

↓

Execution Step 1

↓

Execution Step 2

↓

Execution Step N

↓

Execution Result

↓

Completed
```

The runtime flow is deterministic.

---

# 12. Retry

Retry belongs to the Orchestration Engine.

When a retry is required

* The current Execution remains unchanged.
* Retry metadata is recorded.
* Processing resumes according to the configured retry policy.

Retry does not create duplicate business operations.

---

# 13. Replay

Replay creates a new Execution.

Replay reuses

* Intent
* Context
* Snapshots

Replay generates

* New Execution ID
* New Timeline
* New Runtime Events

Historical Executions remain immutable.

---

# 14. Runtime Events

The runtime publishes events during execution.

Examples

* ExecutionCreated
* ExecutionQueued
* ExecutionStarted
* StepStarted
* StepCompleted
* ExecutionSucceeded
* ExecutionFailed
* ExecutionCompleted

Runtime Events synchronize platform components.

---

# 15. Runtime State

Execution state is managed exclusively by the Orchestration Engine.

State transitions are validated by the State Machine.

Execution state cannot be modified by

* Connectors
* Transformation Engine
* Event Platform
* External Systems

---

# 16. Runtime Ownership

| Runtime Object     | Owner                |
| ------------------ | -------------------- |
| Intent             | Intent Processor     |
| Execution          | Orchestration Engine |
| Execution Step     | Orchestration Engine |
| Execution Context  | Orchestration Engine |
| Execution Timeline | Orchestration Engine |
| Execution Snapshot | Orchestration Engine |
| Execution Result   | Orchestration Engine |

Ownership is exclusive.

No runtime object has multiple owners.

---

# 17. Runtime Persistence

Runtime data is persisted for

* Execution tracking
* Audit
* Replay
* Troubleshooting

The platform persists

* Intent
* Execution
* Execution Steps
* Snapshots
* Runtime Events

Customer business data remains owned by external systems.

---

# 18. Runtime Constraints

Version 1 intentionally excludes

* Parallel Execution
* Workflow Engine
* BPMN
* Dynamic Planning
* Human Approval
* Saga
* AI Planning

The runtime remains deterministic and sequential.

---

# 19. Runtime Guarantees

The Runtime guarantees

* Deterministic execution
* Replayability
* Auditability
* Observability
* Idempotent processing
* Provider independence

These guarantees apply to every business operation.

---

# 20. Guiding Principles

Clients express Business Intent.

The platform creates an Execution.

The Orchestration Engine coordinates Execution Steps.

Execution history is immutable.

Runtime data supports replay, auditing, and troubleshooting.

Keep the Runtime simple, predictable, and reliable.
