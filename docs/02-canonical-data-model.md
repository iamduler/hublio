# Canonical Data Model (CDM)

> Product: Hublio
> Version: 1.0
> Status: Approved
> Last Updated: 2026-07-15

---

# 1. Purpose

The Canonical Data Model (CDM) defines the universal business objects used internally by Hublio.

Every connector MUST communicate through these canonical models.

External systems MUST NEVER communicate directly with each other.

Example:

POS Order
    ↓
Canonical Order
    ↓
Invoice Provider

The Canonical Data Model is independent of:

- Database schema
- Public Platform API schema
- Connector implementation
- External vendor data models

---

# 2. Design Principles

## Connector Agnostic

Canonical models must not contain provider-specific fields.

Example:

❌ misa_tax_code

❌ nhanh_order_status

✔ tax

✔ order_status

---

## Business Oriented

Models describe business concepts.

Not implementation details.

Example:

Customer

instead of

MisaCustomerDTO

---

## Stable

Canonical models should evolve slowly.

Connector implementations may change frequently.

---

## Extensible

Every model supports extension through metadata.

Example

metadata

attributes

custom_fields

without changing the core model.

---

## Immutable Execution Data

Execution snapshots should never be modified.

Historical executions are immutable.

---

# 3. Canonical Resources

Hublio defines the following business resources.

Identity

- Organization
- Workspace
- User
- Member
- Role
- API Key

Business

- Company
- Customer
- Supplier
- Contact
- Address

Catalog

- Product
- Variant
- Unit
- Tax
- Currency

Sales

- Order
- Order Item
- Payment
- Shipment

Invoice

- Invoice
- Invoice Item

Integration

- Connector
- Connection
- Credential

Execution

- Flow
- Execution
- Execution Step

Observability

- Event
- Webhook
- Audit Log

---

# 4. Canonical Resource Rules

Every resource should have:

- id
- external_id
- source_system
- created_at
- updated_at
- metadata

Never use provider-specific identifiers as primary identifiers.

Example

Correct

id

external_id

Incorrect

misa_invoice_id

nhanh_order_id

---

# 5. Canonical Identity

Organization

Represents a customer using Hublio.

Workspace

Logical isolation inside an organization.

User

Human user.

Member

Relationship between user and workspace.

API Key

Machine identity.

---

# 6. Canonical Customer

Represents the legal customer regardless of source system.

Examples

POS Customer

ERP Customer

CRM Contact

Invoice Buyer

All become

Customer

Core Attributes

- id
- code
- name
- tax_code
- email
- phone
- addresses
- contacts
- metadata

---

# 7. Canonical Company

Represents a legal organization.

Examples

Supplier

Seller

Manufacturer

Distributor

Importer

Exporter

All become

Company

Core Attributes

- id
- legal_name
- display_name
- tax_code
- registration_number
- country
- addresses
- contacts

---

# 8. Canonical Product

Represents sellable products.

Core Attributes

- id
- sku
- name
- barcode
- description
- unit
- tax
- attributes
- metadata

Variants belong to Product.

---

# 9. Canonical Order

Represents a commercial transaction before invoicing.

Core Attributes

- id
- order_number
- customer
- items
- currency
- payment
- shipping
- total_amount
- tax_amount
- discount_amount
- status

An order may generate zero or multiple invoices.

---

# 10. Canonical Invoice

Represents a tax invoice.

Invoice lifecycle is independent from the provider.

Core Attributes

- id
- invoice_number
- customer
- seller
- items
- issue_date
- currency
- taxes
- subtotal
- total
- status

Provider-specific information must be stored separately.

---

# 11. Canonical Connector

Represents an integration package.

Examples

MISA

VNPT

Nhanh.vn

KiotViet

SAP

Salesforce

Shopify

The platform never contains provider logic.

Connector owns provider logic.

---

# 12. Canonical Connection

Represents a configured integration.

Example

Workspace A

↓

Nhanh Connector

↓

MISA Connector

Connection stores

- credentials
- mapping profile
- execution policy
- retry policy

---

# 13. Canonical Flow

Represents business orchestration.

Example

Order Received

↓

Validate

↓

Transform

↓

Create Invoice

↓

Wait Result

↓

Completed

Flow is provider independent.

---

# 14. Canonical Execution

Represents one execution instance.

Execution contains

- trigger
- payload
- status
- retry count
- execution logs
- execution steps
- timestamps

Execution data is immutable.

---

# 15. Canonical Event

Everything important generates an event.

Examples

ConnectionCreated

ExecutionStarted

ExecutionSucceeded

ExecutionFailed

WebhookReceived

InvoiceCreated

InvoiceCancelled

Events are immutable.

---

# 16. Canonical Status

Business status should be standardized.

Never expose vendor status internally.

Example

Pending

Processing

Succeeded

Failed

Cancelled

Instead of

MISA_STATUS_101

VNPT_STATUS_A

---

# 17. Canonical Error

Errors should also be standardized.

Every connector translates vendor errors into canonical errors.

Example

AuthenticationError

ValidationError

TimeoutError

RateLimitError

BusinessRuleViolation

ConnectorUnavailable

InternalPlatformError

---

# 18. Metadata

Every canonical resource supports metadata.

Metadata stores

- custom fields
- vendor extensions
- future compatibility

Business logic MUST NOT depend on metadata.

---

# 19. Mapping Rules

Every connector is responsible for:

External DTO

↓

Canonical Resource

↓

External DTO

No connector communicates directly with another connector.

---

# 20. Design Principles Summary

Hublio owns:

- Canonical Models
- Business Rules
- Execution
- Mapping
- Orchestration

Connectors own:

- Authentication
- Vendor APIs
- Vendor DTOs
- Vendor Error Translation
- Vendor Webhooks

External systems never know each other.

Every integration must pass through the Hublio Canonical Data Model.