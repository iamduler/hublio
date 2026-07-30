# 00 — Foundation (API ↔ FE wiring)

> Shared rules for both `apps/web` and `apps/admin`.  
> Legend: `[x]` done · `[~]` partial · `[ ]` not started  
> Policy: **httpOnly JWT + Next proxy** (all browser → Go via Next); orchestration/events also accept API key for machines.

Related: [01-admin-workspace](./01-admin-workspace.md) · [02-user-workspace](./02-user-workspace.md)

> **Status (2026-07-29): Foundation DONE** — optional A (httpOnly JWT + proxy),
> optional B (Go JWT on orchestration/events), soft gaps, and SDK request-DTO
> adoption all shipped and verified (Go tests + web build/lint/test green).
> Docs UI path fix (spec walk-up from `apps/api`) also shipped.
> Only remaining item is a manual login → run-intent smoke test (§5).
> Next workstream: [02-user-workspace §2 gaps](./02-user-workspace.md).

---

## 1. API access policy

| Browser → | Auth | Used for |
| --- | --- | --- |
| Next `/api/auth/*` | sets httpOnly JWT cookies | login / register / logout / session |
| Next `/api/go/*` (`lib/api/client`) | server attaches Bearer + `X-Workspace-ID` | all Go JWT routes (identity, integration, orchestration, events) |
| Go directly (`X-API-KEY`) | machine | external/ops clients only — **not** the browser |

- [x] Documented in root AGENTS + `apps/web/AGENTS.md` + `docs/24-nextjs-architecture.md` §8.1
- [x] httpOnly JWT + Next proxy (optional A)
- [x] Go accepts JWT on intents / executions / events / poll (`MachineOrJWTMiddleware` + `X-Workspace-ID`)
- [x] Soft gaps: BFF/client 401 clears session; intent/execution query keys workspace-scoped; Zod schemas for write forms

---

## 2. Quy trình wire 1 feature

Theo thứ tự — không bắt đầu từ UI:

- [x] **Kênh gọi**: browser → `lib/api/client` → `/api/go` (JWT). Auth → `/api/auth/*`. Machine clients may still use `X-API-KEY` on Go.
- [x] **Types**: request DTOs từ `@hublio/sdk` (`Schemas[...]`); response entities local until named in OpenAPI.
- [x] **API module** (`features/<name>/api.ts`): 1 hàm / endpoint, unwrap `SuccessEnvelope`.
- [x] **Schemas** (`schemas.ts`): Zod cho form ghi (auth, connections, team, workspaces, sync-routes, intents, api-keys).
- [x] **Hooks** (`hooks.ts`): TanStack Query; key từ `lib/query-keys.ts`; mutation invalidate đúng key.
- [x] **Components**: list / detail / form + `LoadingState` / `EmptyState` / `ErrorState`.
- [x] **Page**: chỉ compose, không fetch trực tiếp.
- [x] **i18n**: `messages/{en,vi}/<name>.json` + namespaces.
- [x] **Nav**: sidebar khi màn mới.
- [x] **Errors**: `getApiErrorMessage` / toast.
- [~] **Verify**: chạy lại build/lint/test sau mỗi batch wire (xem §5) — recurring per batch; last run green.

---

## 3. Shared packages

- [x] `@hublio/ui` — shadcn + common primitives + theme
- [x] `@hublio/config` — `tsconfig.base.json`
- [x] `@hublio/sdk` — types từ `api/openapi/openapi.yaml`
- [x] Feature request DTOs adopt `@hublio/sdk` (`Schemas`); response entities remain local until OpenAPI names them
- [x] Sau mỗi thay đổi OpenAPI: `pnpm --filter @hublio/sdk generate` + commit `schema.d.ts`

---

## 4. Cross-cutting (kiểm mỗi feature)

- [x] **Workspace scope**: `workspaceId` từ `useWorkspace()` / cookie → proxy gửi `X-Workspace-ID`
- [x] **Query keys**: workspace-scoped (kể cả intents / executions)
- [x] **Secrets**: JWT httpOnly; workspace API keys không cần cho UI dashboard nữa
- [x] **Envelope**: parse `SuccessEnvelope` / `ErrorEnvelope`
- [x] **Auth 401**: clear client presence + local user snapshot (`lib/api/client` + `bff-client`)
- [x] **Loading / Empty / Error**: primitives sẵn; áp dụng trên mọi view async
- [x] **i18n**: en + vi cho feature đã ship
- [x] **OpenAPI sync**: dual auth (bearer \| apiKey) + `X-Workspace-ID` documented

---

## 5. Verify (Definition of Done)

Last run: 2026-07-30 (manual smoke via BFF against `demo@hublio.local`).

- [x] `pnpm --filter @hublio/web build` xanh (42 routes, incl. `/api/go/[...path]`, `/api/auth/*`)
- [x] lint 0 error (6 warnings pre-existing on provider effects)
- [x] `pnpm --filter @hublio/web test` xanh (12 tests)
- [x] Go: `go -C apps/api build ./cmd/api` + middleware/orchestration/events tests xanh
- [x] Manual smoke (BFF): login → session → workspaces/connectors → create+verify connection → run intent (`echo`) → execution **succeeded** (needs `make worker`) · 401 no-cookie + bad password
- [x] Manual: SyncRoute create (schedule) + enable + **Poll now** — accepted 2 → watermark advanced; 2nd poll accepted 0 / exhausted
- [x] OpenAPI + `schema.d.ts` regenerated in same change

---

## 6. Follow-ups (ngoài scope foundation)

- [x] Manual e2e smoke — auth/intent + Poll sync-route DONE 2026-07-30
- [x] Wire gaps user workspace → [02-user-workspace §2](./02-user-workspace.md) — done 2026-07-29
- [ ] Dọn legacy `lib/api/bff.ts` + `bff-client.ts` sau khi mọi feature dùng `/api/go`
