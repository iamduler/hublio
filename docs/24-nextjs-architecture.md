# 24 - Next.js Architecture

> Product: Hublio
> Version: 1.0
> Status: Architecture Freeze v1

---

# 1. Purpose

This document defines the frontend architecture of Hublio.

The frontend provides a Dashboard for managing the platform.

Business logic remains on the backend.

---

# 2. Technology Stack

Framework

* Next.js 16
* React 19

Language

* TypeScript

Styling

* TailwindCSS
* shadcn/ui

State Management

* TanStack Query

Forms

* React Hook Form
* Zod

Charts

* Recharts

---

# 3. Design Principles

The frontend follows

* Feature-first
* Server-driven
* Component-based
* Type-safe
* Accessible

Business rules never belong to the frontend.

---

# 4. Project Structure

```text
apps/web/               # @hublio/web — user workspace
    app/                # App Router (route groups + [locale])
    features/           # feature-first modules (api, hooks, components)
    components/         # app-specific layout/composition components
    providers/
    lib/                # api client, bff, i18n, utils
    hooks/
    messages/           # next-intl translations (en, vi)
apps/admin/             # @hublio/admin — admin console
    app/
    i18n/
    messages/
packages/
    ui                  # @hublio/ui — shared components + design theme
    config              # @hublio/config — shared tsconfig base
    sdk                 # @hublio/sdk — types generated from openapi.yaml
```

Notes:

* Apps use the App Router directly at `app/` (no `src/` wrapper).
* The shared HTTP client stays in each app under `lib/api`; shared **types** come
  from `@hublio/sdk` (generated from `api/openapi/openapi.yaml`).

---

# 5. App Router

Use the App Router.

Pages belong only inside

```text
app/
```

Features never define routes.

---

# 6. Feature Structure

Each feature is isolated.

Example

```text
features/

    connections/

        api/

        components/

        hooks/

        schemas/

        types/
```

Features should not depend directly on each other.

---

# 7. Components

Component hierarchy

* UI Components (`@hublio/ui`)
* App Shared Components (`apps/web/components`)
* Feature Components (`features/*/components`)

Reusable primitives belong in

```text
packages/ui
```

App-specific layout / composition stays in `apps/web/components`.

---

# 8. API Layer

The frontend communicates only with Hublio REST APIs.

Provider APIs are never called directly.

API modules live in

```text
features/<name>/api.ts
```

Shared fetch helpers live in

```text
apps/web/lib/api/
```

Shared **DTO types** come from `@hublio/sdk` (generated from `api/openapi/openapi.yaml`).
Do not introduce a full generated runtime client unless explicitly requested.
Hand-written feature clients remain the default; adopt SDK types incrementally.

---

# 8.1 Browser → API access (httpOnly JWT proxy)

**Current rule (supersedes hybrid BFF for the dashboard).** Browser never holds
access/refresh tokens or workspace API keys.

```text
Browser
  ├─ /api/auth/* ──► Go /auth/*   (Next sets httpOnly hublio_session / hublio_refresh)
  │
  └─ /api/go/*    ──► Go /api/v1/* (Next attaches Bearer + X-Workspace-ID)
        identity, integration, intents, executions, events, …

Machine / external clients may still call Go directly with X-API-KEY
(orchestration, events, poll, platform). The dashboard does not mint UI API keys.
```

| Path | Browser calls | Why |
| --- | --- | --- |
| Auth | Next `/api/auth/*` | Set/clear httpOnly JWT cookies |
| All dashboard Go APIs | Next `/api/go/*` via `lib/api/client` | JWT stays httpOnly; proxy adds Bearer + workspace |
| Machine orchestration | Go + `X-API-KEY` | External/ops; not used by the browser |

Rules:

* Access and refresh tokens are **httpOnly** (`hublio_session` / `hublio_refresh`).
* Dashboard soft-gate in `proxy.ts` still reads the httpOnly session cookie on the request.
* Go orchestration/events accept **either** `X-API-KEY` **or** Bearer JWT + `X-Workspace-ID`
  (membership checked). See `MachineOrJWTMiddleware`.
* Workspace API keys must never be sent to the browser.
* Prefer `@hublio/sdk` types for request DTOs; keep hand-written feature `api.ts` clients.

---

# 9. State Management

Server State

* TanStack Query

Local State

* React State

Global state libraries are intentionally avoided in Version 1.

---

# 10. Forms

Every form uses

* React Hook Form
* Zod validation

Validation rules should match backend validation.

---

# 11. Authentication

Dashboard **user** authentication uses httpOnly JWT cookies set by Next
`/api/auth/*`. The browser never reads access/refresh tokens.

Workspace **API keys** remain for machine/external clients calling Go directly.
The dashboard UI does not mint or hold those keys.

---

# 12. Authorization

Permissions are evaluated by the backend.

The frontend only adapts the UI based on returned permissions.

---

# 13. Error Handling

Errors are displayed using canonical platform error messages.

Provider-specific errors are never shown.

---

# 14. Loading Strategy

Every asynchronous view should support

* Loading
* Empty
* Error
* Success

Skeleton loaders are preferred over spinners.

---

# 15. Tables

Data tables support

* Cursor pagination
* Filtering
* Sorting
* Column visibility
* Row actions

---

# 16. Internationalization

The UI should be internationalization-ready.

Version 1 ships with English.

Additional locales may be added later.

---

# 17. Accessibility

The frontend targets WCAG 2.2 AA.

Keyboard navigation and screen reader compatibility are required.

---

# 18. Performance

Prefer

* Server Components
* Lazy loading
* Route-level code splitting

Avoid unnecessary client-side rendering.

---

# 19. Testing

Recommended

* Vitest
* React Testing Library
* Playwright

Test business flows rather than implementation details.

---

# 20. Guiding Principles

The frontend is responsible for presentation and user interaction.

Business rules remain in the backend.

Features are isolated.

Types are shared through API contracts (`@hublio/sdk`).

Browser → API access is **hybrid** (JWT direct to Go; API-key routes via Next BFF).
Do not force every call through Next.js.

The Dashboard should remain fast, accessible, and maintainable.
