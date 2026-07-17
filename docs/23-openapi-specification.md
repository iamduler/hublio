# 23 - OpenAPI Specification

> Product: Hublio
> Version: 1.0
> Status: Architecture Freeze v1

---

# 1. Purpose

This document defines the REST API conventions for Hublio.

It establishes a consistent API design that is independent of any specific endpoint.

Detailed OpenAPI YAML files are maintained separately.

---

# 2. API Philosophy

The Hublio API is

* Resource-oriented
* RESTful
* Versioned
* Provider-agnostic
* Canonical Model based

Clients communicate only with Hublio concepts.

Provider APIs are never exposed.

---

# 3. Base URL

```text
/api/v1
```

Future breaking changes require a new version.

---

# 4. Authentication

Supported authentication methods

* API Key
* JWT (Dashboard)

API Keys are intended for machine-to-machine communication.

JWT is intended for Dashboard users.

---

# 5. Resource Groups

The API is organized into the following resource groups.

Identity

* Organizations
* Users
* Workspaces
* API Keys

Integration

* Connectors
* Connections
* Credentials

Orchestration

* Intents
* Executions

Platform

* Health
* Metrics
* Audit Logs

---

# 6. Resource Naming

Resources use plural nouns.

Examples

```text
/users

/workspaces

/connections

/intents

/executions
```

Actions should be represented as resources whenever possible.

Avoid verbs in URLs.

---

# 7. HTTP Methods

GET

Retrieve resources.

POST

Create resources.

PUT

Replace resources.

PATCH

Partially update resources.

DELETE

Soft delete configuration resources.

Runtime resources are never deleted.

---

# 8. Status Codes

200 OK

201 Created

202 Accepted

204 No Content

400 Bad Request

401 Unauthorized

403 Forbidden

404 Not Found

409 Conflict

422 Validation Error

429 Too Many Requests

500 Internal Server Error

---

# 9. Request Format

All requests use JSON.

```json
{
  "data": {}
}
```

---

# 10. Response Format

Successful responses follow

```json
{
  "data": {},
  "meta": {}
}
```

Collections

```json
{
  "data": [],
  "meta": {
    "pagination": {}
  }
}
```

---

# 11. Error Format

Errors follow a consistent structure.

```json
{
  "error": {
    "code": "CONNECTION_DISABLED",
    "message": "Connection is disabled.",
    "details": []
  }
}
```

Provider errors are translated into canonical platform errors.

---

# 12. Pagination

Cursor-based pagination is the default.

```text
GET /executions?cursor=...
```

Offset pagination is not supported.

---

# 13. Filtering

Filtering uses query parameters.

Example

```text
GET /executions?status=running
```

Multiple filters may be combined.

---

# 14. Sorting

Sorting uses

```text
sort=created_at

sort=-created_at
```

Negative prefix indicates descending order.

---

# 15. Idempotency

All POST endpoints that create business operations require

```text
Idempotency-Key
```

The platform guarantees exactly-once processing.

---

# 16. Correlation

Clients may send

```text
X-Correlation-ID
```

If omitted, the platform generates one.

Every response returns the effective Correlation ID.

---

# 17. Long-running Operations

Business operations are asynchronous.

Example

```text
POST /intents
```

↓

```
202 Accepted
```

↓

Execution created.

Clients poll

```text
GET /executions/{id}
```

or receive Webhooks.

---

# 18. API Versioning

Breaking changes require

```text
/api/v2
```

Minor additions remain within the current version.

---

# 19. Documentation Strategy

## Source of truth

```text
api/openapi/openapi.yaml
```

Interactive docs: Scalar UI at `/docs` (dev by default). Raw spec: `/docs/openapi.yaml`.

See also `AGENTS.md` → OpenAPI / API Docs.

## Manual sync (current)

OpenAPI is **maintained by hand**. There is **no** codegen / annotation generator in v1.

Whenever a PR adds, changes, or removes a public HTTP endpoint (path, method, body,
params, auth, status codes), update `api/openapi/openapi.yaml` **in the same PR**.

Do not:

* Skip the OpenAPI update “to follow up later”
* Invent a parallel swagger.json / annotations-only source of truth
* Add swag/apispec codegen unless product explicitly requests it

Split into multiple YAML files (auth, organizations, intents, …) is optional later
when the single file becomes hard to navigate — not required now.

---

# 20. Guiding Principles

The API exposes business capabilities, not implementation details.

Clients communicate using Canonical Models.

The API remains stable even when Connector implementations change.
