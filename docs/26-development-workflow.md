# 26 - Development Workflow

> Product: Hublio
> Version: 1.0
> Status: Architecture Freeze v1

---

# 1. Purpose

This document defines the engineering workflow used to develop Hublio.

The goal is to ensure consistent implementation while preserving the architecture.

---

# 2. Development Philosophy

Every change should follow the Architecture Freeze.

Developers should extend the system rather than redesign it.

Business requirements drive implementation.

---

# 3. Feature Development Lifecycle

Every feature follows the same lifecycle.

```text id="f9mb0m"
Requirement

↓

Domain Analysis

↓

Aggregate Impact

↓

Database (if required)

↓

Application

↓

Infrastructure

↓

REST API

↓

Frontend

↓

Tests

↓

Review

↓

Merge
```

Implementation always begins with the Domain.

---

# 4. Branch Strategy

Recommended branches

* main
* develop
* feature/*
* bugfix/*
* hotfix/*

Short-lived branches are preferred.

---

# 5. Commit Strategy

Commits should be

* Small
* Atomic
* Reversible

One commit should represent one logical change.

---

# 6. Commit Messages

Use Conventional Commits.

Examples

* feat:
* fix:
* refactor:
* test:
* docs:
* chore:

---

# 7. Pull Requests

Every Pull Request should

* Explain the business requirement
* Describe implementation details
* Include testing evidence
* Reference related issues

Large Pull Requests should be avoided.

---

# 8. Code Review

Review focuses on

* Business correctness
* Architecture
* Maintainability
* Security
* Testing

Code style should be enforced by tooling rather than manual review.

---

# 9. Database Changes

Every schema change requires

* Updated DBML
* Migration
* Rollback strategy

Schema changes should remain backward compatible whenever practical.

---

# 10. API Changes

REST API changes require

* Documentation updates
* Backward compatibility review
* Versioning assessment

Breaking changes must be justified.

---

# 11. Frontend Changes

Frontend work follows

* API Contract
* Design System
* Feature Isolation

Business rules remain on the backend.

---

# 12. Testing Requirements

Minimum requirements

* Unit Tests for Domain
* Integration Tests for Infrastructure
* API Tests for REST
* End-to-End Tests for critical flows

Tests are mandatory for new features.

---

# 13. AI-assisted Development

AI coding assistants are encouraged.

Every generated change must comply with

* AGENTS.md
* Coding Standards
* Architecture Principles

Generated code must be reviewed before merging.

---

# 14. Continuous Integration

Every Pull Request should pass

* Formatting
* Linting
* Unit Tests
* Integration Tests
* Security Checks

Merge is blocked if required checks fail.

---

# 15. Release Process

Release workflow

```text id="wb9nr7"
Merge

↓

CI

↓

Build

↓

Deploy to Staging

↓

Verification

↓

Production
```

Deployments should be repeatable and automated.

---

# 16. Incident Handling

Production issues should follow

* Detection
* Mitigation
* Root Cause Analysis
* Permanent Fix
* Postmortem

Temporary fixes should be replaced with permanent solutions.

---

# 17. Documentation

Documentation must evolve with the codebase.

Architecture documents should only change when architecture changes.

Implementation guides may evolve continuously.

---

# 18. Version 1 Constraints

Version 1 intentionally excludes

* Trunk-Based Development
* Release Trains
* Monorepo-specific workflows
* Multi-team coordination processes

The workflow is optimized for a small engineering team.

---

# 19. Guiding Principles

Protect the architecture.

Keep changes small.

Automate repetitive work.

Review generated code.

Build incrementally while preserving long-term maintainability.
