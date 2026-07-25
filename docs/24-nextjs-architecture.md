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

# 8.1 Browser → API access (hybrid BFF)

**Frozen rule for Version 1.** Not every browser call must go through Next.js.

```text
Browser
  ├─ JWT routes ──────────────► Go API (NEXT_PUBLIC_API_URL)
  │   auth, workspaces, connectors, connections,
  │   sync-routes, team, api-keys, …
  │
  └─ API-key-only routes ─────► Next.js BFF (/api/…) ──► Go API (X-API-KEY)
      intents, executions, events
```

| Path | Browser calls | Why |
| --- | --- | --- |
| JWT Identity / Integration CRUD | Go directly (`lib/api/client`) | User JWT already authorizes; Next adds no secret |
| Intent / Execution / Events | Next BFF (`lib/api/bff-client` → `app/api/*`) | Go requires workspace API key; key must stay server-side |

Rules:

* Use Next.js BFF **only** when Next adds security (hold secrets) or useful server orchestration.
* Do **not** proxy all JWT CRUD through Next “for consistency”.
* Workspace API keys must never be sent to the browser. Mint/cache them in
  `lib/api/bff.ts` (httpOnly cookie) and forward as `X-API-KEY` to Go.
* Browser JWT stays in cookies (`hublio_session` / `hublio_refresh`) for V1.
  Moving to httpOnly-only JWT + full Next proxy is a separate security decision,
  not the default architecture.

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

Dashboard **user** authentication uses JWT (browser → Go).

Workspace **API keys** are used only by the Next.js BFF (server → Go) for
routes that do not accept JWT. The browser never holds those API keys.

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
