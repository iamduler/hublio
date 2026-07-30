# API ↔ FE Wiring Checklists

Đã tách thành 3 file theo app / lớp:

| # | File | Phạm vi | Status (2026-07-29) |
| --- | --- | --- | --- |
| 00 | [api-wiring/00-foundation.md](./api-wiring/00-foundation.md) | httpOnly JWT proxy, quy trình wire, shared pkgs, DoD | **DONE** (còn manual smoke) |
| 01 | [api-wiring/01-admin-workspace.md](./api-wiring/01-admin-workspace.md) | `apps/admin` — org ops, connector registry | Scaffold only |
| 02 | [api-wiring/02-user-workspace.md](./api-wiring/02-user-workspace.md) | `apps/web` — ma trận endpoint, gaps | **DONE** core + 5 gaps (2026-07-29) |

Thứ tự làm việc: **00 → 02** (user workspace) → **01** (admin phase).

---

## Snapshot tiến độ

### Đã xong

- Monorepo (`apps/api`, `apps/web`, `apps/admin`, `packages/{ui,config,sdk}`)
- httpOnly JWT + `/api/go` proxy; Go `MachineOrJWTMiddleware` (JWT \| API key)
- Auth đầy đủ: login/register, forgot/reset, verify-email, MFA, onboarding wizard
- User workspace CRUD: connectors (+ enable/disable), connections, sync-routes (+ PATCH/poll/watermark), intents, executions, events (+ filters), api-keys, team invite, settings
- OpenAPI docs path fix (spec resolve từ `apps/api` cwd)

### Gaps đã wire (02 §2)

1. [x] Connector enable/disable UI
2. [x] Sync Route `PATCH` edit
3. [x] Sync Route “Poll now”
4. [x] Watermark editor
5. [x] Events filters (`execution_id` API + `category` client-side)

### Tiếp theo

- Manual smoke: login → dashboard → run intent / poll sync-route
- Dọn legacy `lib/api/bff.ts` / `bff-client.ts`
- Admin phase (`01`) khi product sẵn sàng
- Backend blockers: list members / intents / executions, token refresh, `GET /me`
- Optional: thêm `category` query param trên Go `GET /events` (hiện filter client-side)
