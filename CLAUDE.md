# CLAUDE.md

# Claude Code Instructions for Hublio

## Read First

Before generating or modifying any code, always read:

* AGENTS.md

AGENTS.md is the single source of truth for

* Architecture
* Domain Model
* Coding Standards
* Database Design
* Package Layout

If there is any conflict, AGENTS.md always wins.

---

# Your Role

You are a senior Go software engineer working on Hublio.

Your primary goal is to preserve the architecture while implementing features.

Do not redesign the system.

Build on top of the existing architecture.

---

# Development Workflow

For every task:

1. Understand the business requirement.
2. Identify the affected Bounded Context.
3. Identify the affected Aggregate.
4. Implement the Domain change first.
5. Implement the Application layer.
6. Implement Infrastructure.
7. Implement Interfaces.
8. Add tests.
9. Verify architecture rules.

Never skip these steps.

---

# Code Generation Rules

Always generate

* complete code
* compilable code
* production-ready code

Never generate

* TODO implementations
* placeholder methods
* fake repositories
* mock business logic

unless explicitly requested.

---

# Preferred Order

When implementing a new feature:

1. Domain
2. Repository Interface
3. Use Case
4. Repository Implementation
5. REST Handler
6. Tests

Never start from the REST API.

---

# Go Conventions

Prefer

* explicit constructors
* constructor injection
* small interfaces
* table-driven tests
* early returns

Avoid

* reflection
* globals
* hidden dependencies
* package-level mutable variables

---

# Architecture

Never violate

Interfaces

↓

Application

↓

Domain

↑

Infrastructure

The Domain must remain independent.

---

# Database

Never modify

* table names
* column names
* relationships

unless explicitly instructed.

Always follow

20-database-schema.dbml

---

# Connectors

Every external integration belongs inside

internal/integration/connectors

Never place provider-specific logic elsewhere.

---

# Pull Requests

Before finishing any task, verify

* Architecture respected
* Tests added
* Dependency direction correct
* No duplicated logic
* No business logic inside handlers
* No infrastructure leakage into Domain
* HTTP route changes update `api/openapi/openapi.yaml` in the same change (manual OpenAPI; no codegen unless requested)

---

# Final Rule

If you are uncertain, prefer consistency with the existing architecture over introducing a new abstraction.

When working inside the monorepo:

Identify the affected application.

Only modify packages related to the requested feature.

Avoid unnecessary cross-application changes.