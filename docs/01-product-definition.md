# Product Definition

> Product: Hublio
> Version: 1.0
> Status: Approved
> Last Updated: 2026-07-15

---

# 1. Vision

Build the leading Business Integration Platform and Business Orchestration Platform for business software in Vietnam.

Hublio connects business systems and orchestrates business operations between them through a unified, Canonical-first runtime model.

Invoice integration is the first supported business domain, not the final product.

---

# 2. Mission

Provide a reliable, scalable, and developer-friendly platform that enables businesses to connect their software ecosystem with minimal engineering effort.

The platform should become the central integration and orchestration layer between:

- ERP
- POS
- CRM
- Accounting
- E-commerce
- Shipping
- Payment
- Electronic Invoice Providers
- Government Services

---

# 3. Elevator Pitch

Hublio is both a Business Integration Platform and a Business Orchestration Platform.

Instead of building point-to-point integrations between every pair of systems, software vendors integrate once with Hublio.

Clients submit Business Intents. The platform orchestrates Executions, transforms Canonical data, and communicates with external systems through Connectors.

The platform handles:

- Authentication
- Credential Management
- Canonical Data Model
- Canonical Transformation
- Validation
- Connector Mapping
- Retry
- Monitoring
- Logging
- Webhooks
- Execution
- Connector Management

---

# 4. Product Philosophy

The platform is NOT an invoice software.

The platform is NOT an ERP.

The platform is NOT a Workflow Engine.

The platform is NOT a BPMN or low-code automation tool.

Hublio is a Business Integration Platform + Business Orchestration Platform.

Everything should be designed around one principle:

> Connect Once. Orchestrate Everywhere.

---

# 5. Problem Statement

Today's software ecosystem suffers from point-to-point integrations.

Example:

POS A -> Invoice Provider A

POS A -> Invoice Provider B

POS A -> Invoice Provider C

POS B -> Invoice Provider A

POS B -> Invoice Provider B

POS B -> Invoice Provider C

As the number of software systems increases, the number of required integrations grows exponentially.

Problems:

- High development cost
- Difficult maintenance
- Different authentication methods
- Different API contracts
- Different data formats
- Difficult monitoring
- Difficult retry mechanism
- Vendor lock-in

---

# 6. Solution

Introduce a centralized Business Integration and Orchestration Platform.

Instead of integrating directly:

POS

↓

Invoice Provider

Software systems connect only to Hublio.

Hublio becomes responsible for:

- Authentication
- Connector Lifecycle
- Canonical Data Model
- Canonical Transformation
- Validation
- Provider Mapping (inside Connectors)
- Execution
- Retry
- Monitoring
- Logging
- Webhooks

---

# 7. Product Scope

## In Scope (Phase 1)

- Multi-tenant platform
- Public Platform API
- Connector Framework
- Invoice Integration
- Authentication
- Organization / Workspace / API Key hierarchy
- Credential Management
- Transformation Engine (Canonical → Canonical)
- Orchestration Engine
- Retry
- Replay
- Audit Log
- Monitoring
- Dashboard

---

## Future Scope

- Accounting Integration
- ERP Integration
- CRM Integration
- Shipping Integration
- Payment Integration
- Inventory Integration
- Government Services
- Advanced Orchestration
- Connector Marketplace
- Webhook Engine
- Event Automation
- Public SDK
- Billing Platform

---

## Out of Scope

The platform will NOT:

- Replace ERP software
- Replace Accounting software
- Replace POS software
- Replace Invoice Providers
- Become a Workflow Engine
- Become a BPMN engine
- Become a low-code automation platform like Zapier
- Store customer business transactions as a System of Record

---

# 8. Target Customers

Primary

- POS Vendors
- ERP Vendors
- Accounting Software Vendors
- CRM Vendors
- E-commerce Platforms

Secondary

- Enterprises
- System Integrators
- SaaS Companies

Future

- Government Integrations
- Public APIs
- Independent Developers

---

# 9. Core Value Proposition

Software vendors integrate once.

Hublio manages integrations and orchestrates business operations across downstream systems.

Benefits:

- Faster Integration
- Lower Development Cost
- Standardized API
- Unified Authentication
- Unified Monitoring
- Retry & Replay
- Better Reliability
- Lower Maintenance Cost

---

# 10. Core Principles

## API First

Everything should be accessible through APIs.

---

## Connector First

Business integrations are implemented as connectors.

The platform itself should remain connector-agnostic.

---

## Canonical Data Model

Every external system communicates through the platform's internal data model.

Never expose provider-specific models to the platform core.

---

## Intent Driven

Clients express what they want to accomplish.

The platform determines how the request is executed.

---

## Event Driven

Internal communication should be event-driven whenever appropriate.

---

## Stateless API

Public Platform APIs should remain stateless.

Long-running processes must execute asynchronously.

---

## Idempotency

Every execution must support idempotency.

Repeated requests should produce deterministic results.

---

## Reliability

Retry is a platform capability.

Not a connector responsibility.

---

## Observability

Every execution must be observable.

Every API call must be traceable.

Every failure must be diagnosable.

---

# 11. Success Metrics

Technical

- API Availability ≥ 99.9%
- API Response Time < 300ms
- Connector Success Rate ≥ 99%
- Retry Success Rate ≥ 95%

Business

- 20+ Connectors
- 100+ Organizations
- 1M+ Executions
- 99% Customer Data Integrity

---

# 12. Product Roadmap

Phase 1

Integration & Orchestration Foundation

Phase 2

Invoice Hub

Phase 3

Advanced Orchestration

Phase 4

Connector Marketplace

Phase 5

Webhook Engine

Phase 6

Event Automation

Phase 7

Developer Platform

---

# 13. Core Assets

The competitive advantage of the platform is NOT the user interface.

The long-term assets are:

1. Canonical Data Model
2. Connector Runtime SDK
3. Orchestration Engine
4. Transformation Engine
5. Connector Marketplace
6. Public API
7. Observability Platform

These assets should be reusable across all future integrations.

---

# 14. Technology Stack

Backend

- Go
- PostgreSQL
- Redis

Frontend

- Next.js
- TypeScript
- Tailwind CSS
- shadcn/ui
- TanStack Query
- React Hook Form
- Zod

Infrastructure

- Docker
- Nginx
- S3 Compatible Storage
- OpenTelemetry
- Prometheus
- Grafana

---

# 15. Guiding Principle

The platform should never depend on any specific software vendor.

Everything must communicate through standardized interfaces.

Invoice providers, ERP systems, POS systems, CRM systems, and future integrations are all treated as connectors.

The platform owns integration contracts and orchestration.

Connectors own provider-specific implementation.
