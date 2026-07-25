# Hublio UI — Implementation Checklist

> Tracking sheet for the User Workspace phase.  
> Legend: `[x]` done · `[~]` partial / stub · `[ ]` not started  
> Last reviewed against codebase: 2026-07-25

---

## Monorepo migration (done)

The app moved into a pnpm + Turborepo monorepo. Paths below are now relative to
`apps/web` unless noted.

- [x] Go backend → `apps/api` (module path `hublio` unchanged); `go.work` added
- [x] This app (`hublio-ui`) → `apps/web` (`@hublio/web`)
- [x] Shared UI extracted → `packages/ui` (`@hublio/ui`): shadcn `ui/*`, `common/*`,
      `cn`, `toast`, and the design theme (`packages/ui/src/styles/theme.css`).
      `@/components/ui/*` → `@hublio/ui/ui/*`, `@/components/common/*` → `@hublio/ui/common/*`.
      `@/lib/utils` and `@/lib/toast` remain as thin re-export shims.
- [x] Admin console scaffolded → `apps/admin` (`@hublio/admin`, shared UI + next-intl)
- [x] Shared config → `packages/config` (`tsconfig.base.json`)
- [x] Generated API types → `packages/sdk` (`@hublio/sdk`, from `api/openapi/openapi.yaml`)
- [x] Dockerfiles + compose → `deploy/`; `Makefile`, CI, and docs updated
- [~] Migrate feature `lib/api` clients onto `@hublio/sdk` types (deferred — the
      hand-written, tested clients are intact; types available via `lib/api/sdk.ts`)

---

## API access policy (frozen)

Hybrid BFF — **do not** route all browser traffic through Next.js.

| Browser → | Auth | Used for |
| --- | --- | --- |
| Go (`NEXT_PUBLIC_API_URL` via `lib/api/client`) | JWT | auth, workspaces, connectors, connections, sync-routes, team, api-keys |
| Next BFF (`app/api/*` via `lib/api/bff-client`) | server `X-API-KEY` | intents, executions, events |

- [x] Documented in `docs/24-nextjs-architecture.md` §8.1 and `apps/web/AGENTS.md`
- [x] BFF only for API-key-only orchestration/events routes
- [ ] Optional later: httpOnly-only JWT + Next proxy (security hardening — separate decision)
- [ ] Optional later: Go accepts JWT on orchestration/events (would shrink BFF surface)

---

## Phase 0 — Foundations & theme

- [x] Re-skin tokens to Figma blue (`#2563EB` / `#EFF6FF` / `#FAFAFA`) in `app/globals.css`
- [x] Switch fonts to Inter + JetBrains Mono in `app/layout.tsx`
- [x] Update `DESIGN.md` to blue/Inter brand
- [x] `WorkspaceProvider` + active workspace cookie (`lib/workspace.ts`)
- [x] Workspace switcher in dashboard header
- [x] Bootstrap workspaces via `GET /identity/organizations/{orgId}/workspaces`

---

## Phase 1 — BFF for API-key-only routes

> Scope: intents / executions / events only. JWT CRUD stays browser → Go.
> Policy: see **API access policy (frozen)** above.

- [x] Mint + cache workspace API key server-side (`lib/api/bff.ts`, httpOnly cookie)
- [x] `POST /api/intents` (+ `Idempotency-Key`)
- [x] `GET /api/intents/[intentId]`
- [x] `GET /api/executions/[executionId]`
- [x] `GET /api/executions/[executionId]/timeline`
- [x] `POST /api/executions/[executionId]/cancel`
- [x] `POST /api/executions/[executionId]/retry`
- [x] `GET /api/events`
- [x] Browser `bff` client (`lib/api/bff-client.ts`)

---

## Phase 2 — Shared feature infrastructure

- [x] Feature-module layout `features/<name>/{api,components,hooks,schemas,types}`
- [x] TanStack Query + workspace-scoped `queryKeys` (`lib/query-keys.ts`)
- [x] Shared primitives: `StatusBadge`, `EnvBadge`, `RoleBadge`, `ConfirmDialog`, `KpiCard`, `CopyValue`, `PageHeader`, `FormattedDate`
- [x] shadcn `Form` + Zod + RHF form standard
- [x] Retrofit login / register to Zod + RHF
- [x] Loading / empty / error states reused across features
- [ ] Port remaining Figma primitives: `PlanBadge`, `FilterSelect`, `Pagination`, `ActionMenu`, `BulkBar`, `Sparkline` (only if a screen needs them)

---

## Phase 3 — User Workspace features

### Auth

- [x] Login / register (wired to Go API)
- [~] Forgot password — UI stub only (no backend endpoint)
- [~] Verify email — UI stub only (no backend endpoint)
- [ ] Reset password screen
- [ ] MFA / OTP screen

### Onboarding / workspaces

- [x] Create workspace (`/dashboard/workspaces/new`)
- [x] Workspace settings: enable / disable
- [ ] Dedicated create-org onboarding flow (register already creates org)
- [ ] Invite-team onboarding step + “complete” screen

### Dashboard

- [x] Overview KPIs (connections / connectors / sync routes / events)
- [x] Recent activity feed from `GET /events` via BFF

### Connectors

- [x] Connector list + detail (capabilities)
- [ ] Enable / disable connector actions in UI (API hooks exist)
- [ ] Public marketplace catalog (Admin-phase / no backend)

### Connections

- [x] List / create / detail
- [x] Verify / enable / disable / rotate credentials
- [ ] Health / test connection dedicated screens
- [ ] Advanced config editor

### Sync Routes

- [x] List / create (single destination step) / detail
- [x] Enable / disable / delete / rotate webhook secret / watermarks
- [ ] Multi-step / parallel activity editor
- [ ] PATCH edit existing route

### Intents

- [x] Run intent form (connection + capability + JSON payload)
- [x] Intent detail
- [ ] Intent library / list (backend has no list endpoint)
- [ ] Intent configuration / versions screens

### Executions

- [x] Detail: steps + timeline + cancel + retry
- [ ] Execution list (backend has no list endpoint; deep-link from intent result / events)

### Events

- [x] Explorer table + payload inspector dialog
- [ ] Filter by execution / category in UI
- [ ] Replay event screen (no backend)

### API Keys / credentials

- [x] List / create / rotate / disable (plaintext shown once)
- [ ] Secrets manager / OAuth accounts screens (out of backend scope)

### Team

- [x] Add member (email + role)
- [ ] Member list / profile / roles matrix (backend has no list-members endpoint)
- [ ] Pending invitations / access activity

### Workspace settings

- [x] General: name / environment / status + enable/disable
- [ ] Members & roles tab (blocked on list-members)
- [ ] Security / usage & billing tabs (no backend)

---

## Phase 4 — Navigation, i18n, quality

- [x] Full User Workspace sidebar (Integration / Orchestration / Workspace sections)
- [x] Workspace switcher in header
- [x] Feature namespaces `en` + `vi` for all shipped screens
- [x] Harden client error handling for both Go error shapes (`lib/api/errors.ts`)
- [x] Vitest + Testing Library (envelope, errors, StatusBadge)
- [x] Lint gate + frontend job in `.github/workflows/ci.yml`
- [ ] Broader component / integration tests for feature modules
- [ ] E2E (Playwright) against Go API

---

## Backend gaps (blocks fuller UI)

These are **not frontend-only** — need Go work first:

- [ ] Token **refresh** endpoint (access TTL = 15 min; sessions expire hard)
- [ ] `GET /me` (or equivalent session bootstrap)
- [ ] List workspace **members**
- [ ] List **intents** / **executions** (paginated)
- [ ] Pagination on Identity / Integration list endpoints
- [ ] Optional: accept JWT on orchestration/events (would shrink BFF surface; hybrid rule still applies)
- [ ] Password reset / email verify / MFA endpoints

---

## Deferred — Admin Console (later phase)

Most need new Go endpoints before UI:

- [ ] Mission Control
- [ ] Platform Orgs / Users / Workspaces admin
- [ ] Connector publishing / marketplace management
- [ ] Cross-tenant executions / runtime metrics / queue monitor
- [ ] Infrastructure / worker / scheduler screens
- [ ] Platform Audit / Security / Compliance
- [ ] Billing / subscriptions / invoices
- [ ] Platform Settings / feature flags
- [ ] Help Center / knowledge base

---

## How to update

1. Tick items when merged.
2. Move `[~]` → `[x]` when stubs gain a real API.
3. When starting Admin Console, copy that section into a new checklist or keep tracking here.
