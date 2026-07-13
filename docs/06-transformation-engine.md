# Transformation Engine Specification

> Product: Hublio
> Version: 1.0
> Status: Draft

---

# 1. Purpose

The Transformation Engine is responsible for translating data between external systems and Hublio Canonical Data Models.

The Transformation Engine is the only component that understands how to transform data structures.

Business logic must never contain mapping logic.

Connectors must never map directly to other connectors.

Every transformation follows:

External DTO

↓

Canonical Resource

↓

Canonical Resource

↓

External DTO

---

# 2. Core Philosophy

The Transformation Engine translates data.

It never executes business logic.

It never calls external APIs.

It never persists business data.

Its responsibility is limited to:

- Mapping
- Transformation
- Normalization
- Enrichment
- Validation

---

# 3. Responsibilities

The Transformation Engine owns

- Field Mapping
- Type Conversion
- Enum Translation
- Default Values
- Data Normalization
- Data Enrichment
- Schema Validation

The Transformation Engine does NOT own

- Business Rules
- Authentication
- Retry
- Queue
- Workflow
- Connector Communication

---

# 4. Mapping Flow

External Payload

↓

Provider Mapper

↓

Canonical Resource

↓

Business Processing

↓

Canonical Resource

↓

Provider Mapper

↓

External Payload

The platform never transforms

Provider A

↓

Provider B

directly.

---

# 5. Mapping Layers

Every transformation passes through layers.

Layer 1

Raw Provider DTO

↓

Layer 2

Normalized Provider DTO

↓

Layer 3

Canonical Resource

↓

Layer 4

Target Provider DTO

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

Connector owns

Provider DTO

Authentication

Provider APIs

Transformation Engine owns

Canonical Transformation

Field Mapping

Type Conversion

Validation

Business Layer owns

Business Rules

Business Decisions

Business State

Responsibilities must never overlap.

---

# 30. Design Principles

Every provider speaks its own language.

Hublio speaks one language.

The Transformation Engine is the universal translator.

Business logic should never know

- MISA
- VNPT
- Nhanh.vn
- Shopify
- SAP

Business logic only understands

Canonical Data Models.

The Transformation Engine protects the platform from vendor-specific implementations.