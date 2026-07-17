# 05 - Orchestration Engine

> Product: Hublio
> Version: 1.0
> Status: Architecture Freeze v1

---

# 1. Purpose

The Orchestration Engine is responsible for executing Business Intents.

It coordinates the complete lifecycle of an Execution.

It is the runtime coordinator of the platform.

---

# 2. Responsibilities

The Orchestration Engine owns

* Execution Creation
* Execution Lifecycle
* Execution Step Coordination
* Scheduling
* Retry
* Replay
* Timeout
* Cancellation
* Runtime Events
* Execution Timeline

It does not own

* Provider Communication
* Data Transformation
* Authentication
* Business Rules
* Provider DTOs

---

# 3. Runtime Model

Every Business Intent follows the same execution model.

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

This model is fixed for Version 1.

---

# 4. Execution Lifecycle

Every Execution follows a predictable lifecycle.

```text
Created
    │
    ▼
Queued
    │
    ▼
Running
    │
    ├──────────────┐
    ▼              ▼
Succeeded       Failed
```

Alternative terminal states

* Cancelled
* Expired
* Dead Letter

State transitions are validated by the platform.

---

# 5. Execution

Execution represents one runtime instance of an accepted Intent.

Every Execution contains

* Execution ID
* Intent ID
* Status
* Context
* Execution Steps
* Timeline
* Snapshots
* Events
* Result

Execution is immutable after completion.

---

# 6. Execution Steps

Execution is composed of one or more Steps.

Typical Steps include

1. Validate Request
2. Transform Request
3. Invoke Connector
4. Transform Response
5. Publish Events

Version 1 executes Steps sequentially.

Each Step has one responsibility.

---

# 7. Execution Context

Execution Context contains immutable runtime information.

* Organization
* Workspace
* Connection
* Credentials
* Correlation ID
* Trace ID
* Configuration
* Timeout

The Context remains unchanged throughout the Execution.

---

# 8. Scheduling

Supported scheduling modes

* Immediate
* Delayed
* Scheduled

All scheduled executions become normal Executions before processing.

---

# 9. Retry

Retry is owned exclusively by the Orchestration Engine.

Supported strategies

* Fixed Delay
* Exponential Backoff

Retry policy is configured by the platform.

Connectors must never perform automatic retries.

---

# 10. Replay

Replay creates a new Execution.

Replay reuses

* Intent
* Context
* Snapshots

Replay generates

* New Execution ID
* New Timeline
* New Runtime Events

Historical Executions are never modified.

---

# 11. Timeout

The Orchestration Engine monitors

* Queue Timeout
* Execution Timeout
* Connector Timeout

Timeout handling is deterministic and fully observable.

---

# 12. Cancellation

An Execution may be cancelled before completion.

Cancellation is cooperative.

Terminal Executions cannot be cancelled.

Cancelled Executions remain available for auditing.

---

# 13. Runtime Events

The Orchestration Engine publishes Runtime Events.

Examples

* ExecutionCreated
* ExecutionQueued
* ExecutionStarted
* StepStarted
* StepCompleted
* ExecutionSucceeded
* ExecutionFailed
* ExecutionCancelled

Runtime Events synchronize other platform components.

---

# 14. Execution Timeline

Every state transition is recorded.

Timeline entries include

* Timestamp
* Event
* Previous State
* Current State
* Execution ID
* Correlation ID

Timeline history is immutable.

---

# 15. Execution Snapshots

Snapshots preserve runtime information.

Supported snapshots

* Canonical Request
* Provider Request
* Provider Response
* Canonical Response

Snapshots support replay, troubleshooting, and auditing.

---

# 16. Failure Handling

Execution failures are classified.

* Validation Failure
* Transformation Failure
* Connector Failure
* Timeout
* Platform Failure

Every failure produces

* Runtime Event
* Timeline Entry
* Audit Record

---

# 17. Observability

Every Execution exposes

* Status
* Current Step
* Duration
* Retry Count
* Timeline
* Runtime Events
* Logs
* Trace

Every Execution must be traceable end-to-end.

---

# 18. Runtime Boundaries

The Orchestration Engine coordinates execution.

It never

* Calls provider APIs directly
* Performs data transformation
* Stores business data
* Implements provider-specific behavior

Those responsibilities belong to other platform components.

---

# 19. Scalability

The Orchestration Engine is stateless.

Execution state is stored in persistent storage.

Multiple workers may process Executions concurrently.

Horizontal scaling must not affect execution correctness.

---

# 20. Version 1 Constraints

Version 1 intentionally excludes

* Parallel Steps inside a single Execution
* Workflow Engine
* BPMN
* Saga
* Human Approval
* Dynamic Planning
* AI Planning

Clarification: the Orchestration Engine may create multiple Executions from one Intent for SyncRoute fan-out (sequential or parallel enqueue). Each Execution runs sequential Steps.

The execution model remains intentionally simple.

---

# 21. Guiding Principles

The client submits a Business Intent.

The platform creates one or more Executions.

The Orchestration Engine coordinates Execution Steps.

Specialized components perform specialized work.

The platform prioritizes simplicity, determinism, reliability, and observability over flexibility.
