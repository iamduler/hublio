# Product Definition

> Product: Hublio
> Version: 0.1
> Status: Draft
> Last Updated: 2026-07-09

---

# 1. Vision

Build the leading Integration Platform as a Service (iPaaS) for business software in Vietnam.

The platform provides a unified integration hub that allows software vendors and enterprises to connect multiple business systems through standardized APIs, eliminating point-to-point integrations.

Invoice integration is the first supported business domain, not the final product.

---

# 2. Mission

Provide a reliable, scalable, and developer-friendly integration platform that enables businesses to connect their software ecosystem with minimal engineering effort.

The platform should become the central integration layer between:

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

Our platform acts as a central integration hub between business software systems.

Instead of building individual integrations between every pair of systems, software vendors integrate only once with our platform.

The platform handles:

- Authentication
- Credential Management
- Data Mapping
- Data Transformation
- Validation
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

The platform is NOT a workflow automation tool.

The platform is an Integration Hub.

Everything should be designed around one principle:

> Connect Once. Integrate Everywhere.

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

Introduce a centralized Integration Hub.

Instead of integrating directly:

POS

↓

Invoice Provider

Software systems connect only to the Integration Hub.

The Integration Hub becomes responsible for:

- Authentication
- Connector Lifecycle
- Canonical Data Model
- Mapping
- Validation
- Transformation
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
- Organization Management
- Credential Management
- Transformation Engine
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
- Workflow Builder
- Connector Marketplace
- Public SDK
- Billing Platform

---

## Out of Scope

The platform will NOT:

- Replace ERP software
- Replace Accounting software
- Replace POS software
- Replace Invoice Providers
- Store business transactions permanently
- Become a workflow automation platform like Zapier

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

Our platform manages all downstream integrations.

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

Never expose provider-specific models to the core platform.

---

## Event Driven

Internal communication should be event-driven whenever possible.

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

Integration Platform Foundation

Phase 2

Invoice Hub

Phase 3

Workflow Engine

Phase 4

Connector Marketplace

Phase 5

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

The platform owns the orchestration.

Connectors own the implementation.