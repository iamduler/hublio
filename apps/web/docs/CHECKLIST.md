# Hublio UI — Implementation Checklist

> Tracking sheet for the User Workspace phase.  
> Legend: `[x]` done · `[~]` partial / stub · `[ ]` not started  
> Last reviewed against codebase: 2026-07-29  
> API wiring detail: [API_WIRING_CHECKLIST.md](./API_WIRING_CHECKLIST.md)

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

## API access policy

httpOnly JWT proxy — browser never holds tokens or workspace API keys.

| Browser → | Auth | Used for |
| --- | --- | --- |
| Next `/api/auth/*` | httpOnly cookies | login / register / logout / session |
| Next `/api/go/*` (`lib/api/client`) | server Bearer + `X-Workspace-ID` | identity, integration, intents, executions, events |

- [x] Documented in `docs/24-nextjs-architecture.md` §8.1 and `apps/web/AGENTS.md`
- [x] httpOnly JWT + Next proxy
- [x] Go accepts JWT on orchestration/events (`MachineOrJWTMiddleware`)

---

## Phase 0 — Foundations & theme

- [x] Re-skin tokens to Figma blue (`#2563EB` / `#EFF6FF` / `#FAFAFA`) in `app/globals.css`
- [x] Switch fonts to Inter + JetBrains Mono in `app/layout.tsx`
- [x] Update `DESIGN.md` to blue/Inter brand
- [x] `WorkspaceProvider` + active workspace cookie (`lib/workspace.ts`)
- [x] Workspace switcher in dashboard header
- [x] Bootstrap workspaces via `GET /identity/organizations/{orgId}/workspaces`

---

## Phase 1 — Auth proxy + orchestration JWT

> Supersedes the earlier API-key BFF for the dashboard UI.

- [x] `/api/auth/login|register|logout|session` (httpOnly cookies)
- [x] Catch-all `/api/go/[...path]` JWT proxy
- [x] Feature clients use `lib/api/client` → `/api/go` (including intents/executions/events)
- [x] Legacy `/api/intents|executions|events` routes rewired to JWT proxy
- [x] Removed deprecated `lib/api/bff.ts` + `bff-client.ts` (all traffic via `proxy-go` + `client`)
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

- [x] Login / register (wired to Go API via `/api/auth/*`)
- [x] Forgot password + reset password (`/forgot-password`, `/reset-password?token=`)
- [x] Verify email (`/verify-email`; register redirects here)
- [x] MFA challenge on login + enroll/disable in Settings → Security

### Onboarding / workspaces

- [x] Create workspace (`/dashboard/workspaces/new`)
- [x] Workspace settings: enable / disable
- [x] Onboarding wizard: org → workspace → invite (Skip) → complete
- [x] Routes: `/onboarding/organization|workspace|invite|complete`

### Dashboard

- [x] Overview KPIs (connections / connectors / sync routes / events)
- [x] Recent activity feed from `GET /events` via JWT proxy

### Connectors

- [x] Connector list + detail (capabilities)
- [x] Enable / disable connector actions in UI
- [ ] Public marketplace catalog (Admin-phase / no backend)

### Connections

- [x] List / create / detail
- [x] Verify / enable / disable / rotate credentials
- [ ] Health / test connection dedicated screens
- [ ] Advanced config editor

### Sync Routes

- [x] List / create (single destination step) / detail
- [x] Enable / disable / delete / rotate webhook secret / watermarks GET
- [x] PATCH edit existing route (`/sync-routes/[id]/edit`, draft|disabled only)
- [x] Poll now button → `POST /sync-routes/:id/poll`
- [x] Watermark editor UI → `PUT watermarks/:resourceType`
- [ ] Multi-step / parallel activity editor

### Intents

- [x] Run intent form (connection + capability + JSON payload)
- [x] Intent detail
- [x] Intent library / list (limit; no cursor yet)
- [ ] Intent configuration / versions screens

### Executions

- [x] Detail: steps + timeline + cancel + retry
- [x] Execution list (workspace-scoped; slim rows → detail)

### Events

- [x] Explorer table + payload inspector dialog
- [x] Filter by execution ID (API) + category (client-side)
- [ ] Replay event screen (no backend)

### API Keys / credentials

- [x] List / create / rotate / disable (plaintext shown once)
- [ ] Secrets manager / OAuth accounts screens (out of backend scope)

### Team

- [x] Add member (email + role)
- [x] Member list (roles shown; profile / roles matrix later)
- [ ] Pending invitations / access activity

### Workspace settings

- [x] General: name / environment / status + enable/disable
- [x] Security tab (MFA enroll / disable)
- [ ] Members & roles tab (list API exists; settings tab later)
- [ ] Usage & billing tabs (no backend)

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

- [x] Token **refresh** endpoint (access TTL = 15 min; rotate via refresh cookie)
- [x] `GET /me` (or equivalent session bootstrap)
- [x] List workspace **members**
- [x] List **intents** / **executions** (limit + optional status; no cursor yet)
- [ ] Pagination on Identity / Integration list endpoints
- [ ] Optional: accept JWT on orchestration/events (would shrink BFF surface; hybrid rule still applies)
- [x] Password reset / email verify / MFA endpoints

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
