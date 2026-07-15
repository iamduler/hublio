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
apps/dashboard/src/
    app/
    features/
    components/
    lib/
    hooks/
    services/
    types/
    styles/
	packages/
		ui
		api-client
		config
		types
```

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

* UI Components
* Shared Components
* Feature Components

Reusable UI belongs in

```text
components/ui
```

---

# 8. API Layer

The frontend communicates only with Hublio REST APIs.

Provider APIs are never called directly.

API calls belong inside

```text
services/
```

or feature-specific API modules.

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

Dashboard authentication uses JWT.

API Keys are never used by the Dashboard.

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

Types are shared through API contracts.

The Dashboard should remain fast, accessible, and maintainable.
