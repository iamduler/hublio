# API ↔ FE Wiring Checklists

Đã tách thành 3 file theo app / lớp:

| # | File | Phạm vi | Status (2026-07-29) |
| --- | --- | --- | --- |
| 00 | [api-wiring/00-foundation.md](./api-wiring/00-foundation.md) | httpOnly JWT proxy, quy trình wire, shared pkgs, DoD | **DONE** (còn manual smoke) |
| 01 | [api-wiring/01-admin-workspace.md](./api-wiring/01-admin-workspace.md) | `apps/admin` — org ops, connector registry | Scaffold only |
| 02 | [api-wiring/02-user-workspace.md](./api-wiring/02-user-workspace.md) | `apps/web` — ma trận endpoint, gaps | **DONE** core + lists (2026-07-30) |

Thứ tự làm việc: **00 → 02** (user workspace) → **01** (admin phase).

---

## Snapshot tiến độ

### Đã xong

- Monorepo (`apps/api`, `apps/web`, `apps/admin`, `packages/{ui,config,sdk}`)
- httpOnly JWT + `/api/go` proxy; Go `MachineOrJWTMiddleware` (JWT \| API key)
- Auth đầy đủ: login/register, forgot/reset, verify-email, MFA, onboarding wizard
- User workspace CRUD: connectors (+ enable/disable), connections, sync-routes (+ PATCH/poll/watermark), intents (+ list), executions (+ list), events (+ filters), api-keys, team invite + members, settings
- OpenAPI docs path fix (spec resolve từ `apps/api` cwd)

### Gaps đã wire (02 §2)

1. [x] Connector enable/disable UI
2. [x] Sync Route `PATCH` edit
3. [x] Sync Route “Poll now”
4. [x] Watermark editor
5. [x] Events filters (`execution_id` API + `category` client-side)

### Tiếp theo

- [x] Manual smoke: login → intent + Poll sync-route DONE (2026-07-30)
- [x] List members / intents / executions (API + FE)
- [x] Token refresh (`POST /auth/refresh` + BFF + silent proxy rotate)
- Dọn legacy `lib/api/bff.ts` / `bff-client.ts`
- Admin phase (`01`) khi product sẵn sàng
- Backend blockers: `GET /me`
- Optional: thêm `category` query param trên Go `GET /events` (hiện filter client-side)
- Note: demo MFA đang `pending_enrollment=true` (enrol dở) — disable/re-setup từ Settings → Security nếu cần
