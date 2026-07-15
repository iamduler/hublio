# Transformation Engine Specification

> Product: Hublio
> Version: 1.0
> Status: Approved

---

# 1. Purpose

The Transformation Engine is responsible for Canonical → Canonical transformations inside Hublio.

It never maps Provider DTOs.

Provider DTO mapping belongs exclusively to Connector Runtime.

Business logic must never contain Canonical transformation logic.

Connectors must never map directly to other connectors.

Every platform transformation follows:

Canonical Resource

↓

Transformation Engine (Canonical → Canonical)

↓

Canonical Resource

Examples

* Currency normalization
* Timezone normalization
* Field rename within Canonical Models
* Default values
* Schema validation
* Data enrichment within Canonical Models

---

# 2. Core Philosophy

The Transformation Engine transforms Canonical Models only.

It never executes business rules.

It never calls external APIs.

It never persists business data.

It never understands Provider DTOs.

Its responsibility is limited to:

- Canonical Field Rename
- Type Conversion within Canonical Models
- Normalization
- Enrichment
- Validation

---

# 3. Responsibilities

The Transformation Engine owns

- Canonical Field Rename
- Type Conversion
- Enum Normalization within Canonical Models
- Default Values
- Data Normalization
- Data Enrichment within Canonical Models
- Canonical Schema Validation

The Transformation Engine does NOT own

- Provider DTO Mapping
- Business Rules
- Authentication
- Retry
- Queue Management
- Workflow Logic
- Connector Communication / HTTP

---

# 4. Mapping Flow

Canonical Request

↓

Transformation Engine (Canonical → Canonical)

↓

Normalized Canonical Request

↓

Connector Runtime (Canonical → Provider DTO → HTTP → Provider Response → Canonical Response)

↓

Transformation Engine (Canonical → Canonical, if required)

↓

Normalized Canonical Response

The platform never transforms

Provider A

↓

Provider B

directly.

All provider-specific mapping stays inside Connector Runtime.

---

# 5. Transformation Layers

Every Canonical transformation may pass through layers such as:

Layer 1

Raw Canonical Resource

↓

Layer 2

Normalized Canonical Resource

↓

Layer 3

Validated Canonical Resource

Provider DTOs never enter these layers.

---

# 6. Mapping Types

The engine supports

Field Mapping

Type Conversion

Enum Mapping

Nested Mapping

Collection Mapping

Expression Mapping

Conditional Mapping

Default Value Mapping

Metadata Mapping

---

# 7. Field Mapping

Example

Provider

customerName

↓

Canonical

customer.name

↓

Provider

buyer_name

Field names should never leak into business logic.

---

# 8. Type Conversion

Supported conversions

String

Integer

Decimal

Boolean

Date

DateTime

UUID

ULID

Currency

Country Code

Language Code

Every conversion must be deterministic.

---

# 9. Enum Translation

Provider enums are translated into canonical enums.

Example

Provider A

PAID

↓

Canonical

Completed

↓

Provider B

Success

Business logic only understands canonical values.

---

# 10. Date & Time

Canonical format

RFC3339 UTC

Every provider date format must be converted.

Provider

09/07/2026

↓

Canonical

2026-07-09T00:00:00Z

---

# 11. Currency

Canonical currency

ISO 4217

Examples

VND

USD

JPY

Providers using numeric codes must be translated.

---

# 12. Country

Canonical country

ISO 3166-1 Alpha-2

Example

VN

US

JP

Providers may use

84

VNM

Vietnam

All become

VN

---

# 13. Language

Canonical language

ISO 639-1

Example

vi

en

ja

---

# 14. Address

Every connector maps into

Canonical Address

Standard attributes

Country

Province

District

Ward

Postal Code

Street

Additional fields belong to metadata.

---

# 15. Tax

Canonical tax model

Tax Type

Tax Rate

Tax Amount

Tax Category

Provider-specific tax codes remain inside connectors.

---

# 16. Unit

Canonical units

Piece

Box

Pack

Kg

Gram

Liter

Meter

Providers convert internally.

---

# 17. Money

Money is never represented by float.

Canonical Money

Currency

Amount

Precision

Rounding belongs to business rules.

---

# 18. Metadata

Every resource supports

metadata

Connector-specific extensions belong here.

Business logic must never depend on metadata.

---

# 19. Mapping Profiles

A Connection selects one Mapping Profile.

Mapping Profile defines

Field Mapping

Enum Mapping

Transformation Rules

Default Values

Validation Rules

Profiles are versioned.

---

# 20. Mapping Rules

Rules may include

Copy

Rename

Constant

Lookup

Convert

Calculate

Concatenate

Split

Trim

Uppercase

Lowercase

Replace

Format

Rules should remain deterministic.

---

# 21. Lookup Mapping

Lookup tables translate values.

Example

COD

↓

Cash

Province Code

↓

Province Name

Tax Code

↓

Tax Rate

Lookup data should be configurable.

---

# 22. Conditional Mapping

Conditional transformations

Example

If

Country = VN

↓

Tax = VAT

Else

GST

Conditions must remain side-effect free.

---

# 23. Validation

The Transformation Engine performs schema validation.

Examples

Required Field

Maximum Length

Data Type

Regex

Supported Enum

Business validation belongs elsewhere.

---

# 24. Transformation Pipeline

Incoming Payload

↓

Parse

↓

Normalize

↓

Validate

↓

Transform

↓

Canonical Resource

↓

Business Layer

↓

Canonical Resource

↓

Transform

↓

Serialize

↓

Outgoing Payload

---

# 25. Versioning

Mappings are versioned independently.

Connector Version

≠

Mapping Version

Execution stores

Mapping Version

for reproducibility.

---

# 26. Error Handling

Mapping errors

Unknown Field

Invalid Enum

Type Conversion

Missing Required Field

Unsupported Value

Invalid Format

Mapping errors stop execution before connector execution.

---

# 27. Performance

Mappings should

Avoid Reflection when possible.

Avoid Dynamic Evaluation.

Support High Throughput.

Cache lookup tables.

Be deterministic.

---

# 28. Extensibility

Future mapping capabilities

Expression Language

Custom Functions

AI-assisted Mapping

Visual Mapping Designer

Schema Evolution

Plugin-based Transformers

Architecture should support these without breaking compatibility.

---

# 29. Mapping Ownership

Connector Runtime owns

* Provider DTO
* Canonical ↔ Provider DTO Mapping
* Authentication
* Provider APIs
* HTTP Communication
* Error Translation

Transformation Engine owns

* Canonical → Canonical Transformation
* Field Rename
* Type Conversion
* Normalization
* Validation
* Enrichment within Canonical Models

Business Layer owns

* Business Rules
* Business Decisions
* Business State

Responsibilities must never overlap.

---

# 30. Design Principles

Every provider speaks its own language.

Hublio speaks one language: Canonical Data Models.

Connector Runtime translates between Canonical Models and Provider DTOs.

Transformation Engine keeps Canonical Models consistent, validated, and normalized.

Business logic should never know

- MISA
- VNPT
- Nhanh.vn
- Shopify
- SAP

Business logic only understands

Canonical Data Models.

Connectors protect the platform from vendor-specific APIs and DTOs.

Transformation Engine protects Canonical consistency inside the platform.
