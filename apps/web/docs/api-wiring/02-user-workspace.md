# 02 — User Workspace (`apps/web`)

> Tenant user workspace — Integration + Orchestration UI.  
> Legend: `[x]` wired end-to-end · `[~]` client/hook có nhưng thiếu UI · `[ ]` chưa wire  
> App: `apps/web` (`@hublio/web`)

Related: [00-foundation](./00-foundation.md) · [01-admin-workspace](./01-admin-workspace.md)

---

## 1. Ma trận endpoint ↔ FE

### Identity (JWT → Go)

- [x] `GET /identity/organizations/:orgId` — bootstrap
- [x] `GET /identity/organizations/:orgId/workspaces` — list + switcher
- [x] `POST /identity/organizations/:orgId/workspaces` — create workspace
- [x] `POST /identity/workspaces/:id/enable`
- [x] `POST /identity/workspaces/:id/disable`
- [x] `POST /identity/workspaces/:id/members` — invite member
- [x] `GET /identity/workspaces/:id/api-keys`
- [x] `POST /identity/workspaces/:id/api-keys`
- [x] `POST .../api-keys/:keyId/disable`
- [x] `POST .../api-keys/:keyId/rotate`

> Org suspend/activate và register connector → [01-admin-workspace](./01-admin-workspace.md).

### Integration — Connectors (JWT → Go)

- [x] `GET /integration/connectors`
- [x] `GET /integration/connectors/:id`
- [~] `POST .../connectors/:id/enable` — hook `useToggleConnector` có, **thiếu nút UI**
- [~] `POST .../connectors/:id/disable` — như trên

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
- [ ] `PATCH .../sync-routes/:sid` — edit chưa wire
- [x] `POST .../sync-routes/:sid/enable`
- [x] `POST .../sync-routes/:sid/disable`
- [x] `DELETE .../sync-routes/:sid`
- [x] `POST .../sync-routes/:sid/webhook-secret/rotate`
- [x] `GET .../sync-routes/:sid/watermarks`
- [~] `PUT .../watermarks/:resourceType` — API có, **thiếu UI editor**

### Orchestration (JWT proxy → Go; machines still use X-API-KEY)

- [x] `POST /intents` — `/api/go` (+ `Idempotency-Key`)
- [x] `GET /intents/:id`
- [x] `GET /executions/:id`
- [x] `GET /executions/:id/timeline`
- [x] `POST /executions/:id/cancel`
- [x] `POST /executions/:id/retry`
- [ ] `POST /sync-routes/:id/poll` — UI chưa wire (Go đã nhận JWT hoặc API key)
- [n/a] `POST /webhooks/sync-routes/:id` — inbound, không phải FE

### Events (JWT proxy → Go)

- [x] `GET /events`
- [ ] Filter UI theo `execution_id` / `category` (backend đã hỗ trợ query params)

---

## 2. Gaps ưu tiên wire tiếp

- [ ] Connector enable/disable: action + confirm trên `connector-detail.tsx` (hook đã có)
- [ ] Sync Route `PATCH` edit form
- [ ] Sync Route poll: nút “Poll now” → `POST /sync-routes/:id/poll` qua `/api/go`
- [ ] Watermark editor: UI cho `PUT watermarks/:resourceType`
- [ ] Events filters: bind `execution_id`, `category`, cursor

---

## 3. Feature screens (tóm tắt)

| Area | Status |
| --- | --- |
| Auth login/register | [x] |
| Forgot / verify / MFA | [~] stub / chưa backend |
| Create workspace | [x] |
| Dashboard KPIs + activity | [x] |
| Connectors list/detail | [x] · enable/disable UI [~] |
| Connections CRUD + verify | [x] |
| Sync routes CRUD + webhook/watermarks GET | [x] · PATCH/poll/PUT watermark [ ] |
| Intents run + detail | [x] · list [ ] (no list API) |
| Executions detail + cancel/retry | [x] · list [ ] (no list API) |
| Events explorer | [x] · filters [ ] |
| API keys | [x] |
| Team invite | [x] · member list [ ] (no list API) |
| Workspace settings | [x] general · members/security tabs [ ] |

---

## 4. Ngoài phạm vi backend hiện tại

- [ ] Intent list, Execution list → deep-link từ result / events
- [ ] List workspace members
- [ ] Forgot/reset password, verify email, MFA
- [ ] Replay event, billing → Admin-phase / chưa backend

---

## 5. Verify (user workspace)

Áp dụng [00-foundation §5](./00-foundation.md) với filter `@hublio/web`.
