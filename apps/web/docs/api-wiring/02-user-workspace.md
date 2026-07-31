# 02 — User Workspace (`apps/web`)

> Tenant user workspace — Integration + Orchestration UI.  
> Legend: `[x]` wired end-to-end · `[~]` client/hook có nhưng thiếu UI · `[ ]` chưa wire  
> App: `apps/web` (`@hublio/web`)  
> **Status (2026-07-30): Core CRUD + list members/intents/executions DONE** — auth/onboarding +
> connector toggle, sync-route PATCH/poll/watermark, events filters, Team/Intents/Executions lists.
> Remaining backend gaps: `GET /me`, cursor pagination.

Related: [00-foundation](./00-foundation.md) · [01-admin-workspace](./01-admin-workspace.md)

---

## 1. Ma trận endpoint ↔ FE

### Auth — MFA/2FA (BFF `/api/auth/*` → Go `/auth/*`)

- [x] `POST /auth/login` — trả `{ mfa_required, mfa_token }` khi user bật MFA (BFF **không** set cookie)
- [x] `POST /auth/mfa/verify` — TOTP `code` hoặc `recovery_code`; BFF set cookie khi thành công
- [x] `POST /auth/mfa/setup` (JWT) — secret + `otpauth_url` + recovery codes, hiện **1 lần**
- [x] `POST /auth/mfa/enable` (JWT) — xác nhận `code` 6 số
- [x] `POST /auth/mfa/disable` (JWT) — cần `password`
- [x] `GET /auth/mfa/status` (JWT) — enabled / pending / recovery count
- [x] Settings → Security — MFA enrollment UI

> BFF: `app/api/auth/login` (pass-through challenge) + `app/api/auth/mfa/{status,setup,enable,disable,verify}`.

### Auth — Password reset & email verify

- [x] `POST /auth/forgot-password` + UI `/(auth)/forgot-password`
- [x] `POST /auth/reset-password` + UI `/(auth)/reset-password?token=`
- [x] `POST /auth/verify-email/request` + `POST /auth/verify-email` + UI `/(auth)/verify-email?email=`
- [x] Register → redirect verify-email (OTP gửi best-effort sau register)
- [x] `POST /auth/refresh` — rotate tokens; BFF `/api/auth/refresh` + silent retry in `proxyGoWithJWT`

### Auth — Onboarding wizard (OAuth hybrid)

- [x] Organization → Workspace → Invite team (Skip) → Complete
- [x] Routes: `/onboarding/organization|workspace|invite|complete`

### Identity (JWT → Go)

- [x] `GET /identity/organizations/:orgId` — bootstrap
- [x] `GET /identity/organizations/:orgId/workspaces` — list + switcher
- [x] `POST /identity/organizations/:orgId/workspaces` — create workspace
- [x] `POST /identity/workspaces/:id/enable`
- [x] `POST /identity/workspaces/:id/disable`
- [x] `POST /identity/workspaces/:id/members` — invite member
- [x] `GET /identity/workspaces/:id/members` — list members
- [x] `GET /identity/workspaces/:id/api-keys`
- [x] `POST /identity/workspaces/:id/api-keys`
- [x] `POST .../api-keys/:keyId/disable`
- [x] `POST .../api-keys/:keyId/rotate`

> Org suspend/activate và register connector → [01-admin-workspace](./01-admin-workspace.md).

### Integration — Connectors (JWT → Go)

- [x] `GET /integration/connectors`
- [x] `GET /integration/connectors/:id`
- [x] `POST .../connectors/:id/enable` — UI trên connector detail
- [x] `POST .../connectors/:id/disable` — UI + ConfirmDialog

### Integration — Connections (JWT → Go)

- [x] `GET .../connections`
- [x] `POST .../connections`
- [x] `GET .../connections/:cid`
- [x] `POST .../connections/:cid/verify`
- [x] `POST .../connections/:cid/enable`
- [x] `POST .../connections/:cid/disable`
- [x] `POST .../connections/:cid/credentials/rotate`

### Integration — Sync Routes (JWT → Go)

- [x] `GET .../sync-routes`
- [x] `POST .../sync-routes`
- [x] `GET .../sync-routes/:sid`
- [x] `PATCH .../sync-routes/:sid` — edit form (`/edit`, chỉ draft/disabled)
- [x] `POST .../sync-routes/:sid/enable`
- [x] `POST .../sync-routes/:sid/disable`
- [x] `DELETE .../sync-routes/:sid`
- [x] `POST .../sync-routes/:sid/webhook-secret/rotate`
- [x] `GET .../sync-routes/:sid/watermarks`
- [x] `PUT .../watermarks/:resourceType` — editor dialog

### Orchestration (JWT proxy → Go; machines still use X-API-KEY)

- [x] `GET /intents` — list (limit + optional status; no payload)
- [x] `POST /intents` — `/api/go` (+ `Idempotency-Key`)
- [x] `GET /intents/:id`
- [x] `GET /executions` — list (limit + optional status; slim)
- [x] `GET /executions/:id`
- [x] `GET /executions/:id/timeline`
- [x] `POST /executions/:id/cancel`
- [x] `POST /executions/:id/retry`
- [x] `POST /sync-routes/:id/poll` — nút “Poll now” trên detail
- [n/a] `POST /webhooks/sync-routes/:id` — inbound, không phải FE

### Events (JWT proxy → Go)

- [x] `GET /events`
- [x] Filter UI: `execution_id` (API query) + `category` (client-side trên page)

---

## 2. Gaps ưu tiên wire tiếp

- [x] Connector enable/disable: action + confirm trên `connector-detail.tsx`
- [x] Sync Route `PATCH` edit form
- [x] Sync Route poll: nút “Poll now” → `POST /sync-routes/:id/poll` qua `/api/go`
- [x] Watermark editor: UI cho `PUT watermarks/:resourceType`
- [x] Events filters: bind `execution_id`, `category`

---

## 3. Feature screens (tóm tắt)

| Area | Status |
| --- | --- |
| Auth login/register | [x] |
| Forgot / reset / verify | [x] |
| MFA challenge (login) | [x] |
| MFA enrol (Settings → Security) | [x] |
| Onboarding wizard (org→ws→invite→complete) | [x] |
| Create workspace | [x] |
| Dashboard KPIs + activity | [x] |
| Connectors list/detail + enable/disable | [x] |
| Connections CRUD + verify | [x] |
| Sync routes CRUD + PATCH/poll/watermark | [x] |
| Intents run + detail + list | [x] |
| Executions detail + cancel/retry + list | [x] |
| Events explorer + filters | [x] |
| API keys | [x] |
| Team invite + member list | [x] |
| Workspace settings | [x] general + security · members tab [ ] |

---

## 4. Ngoài phạm vi backend hiện tại

- [x] Intent list, Execution list, List workspace members
- [ ] Replay event, billing → Admin-phase / chưa backend
- [ ] Cursor/`pagination` envelope on list APIs
- [x] Token refresh (`POST /auth/refresh` + BFF `/api/auth/refresh` + proxy silent rotate)
- [ ] `GET /me`

---

## 5. Verify (user workspace)

Áp dụng [00-foundation §5](./00-foundation.md) với filter `@hublio/web`.
