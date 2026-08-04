# 01 — Admin Workspace (`apps/admin`)

> Platform / org administration console.  
> Legend: `[x]` done · `[~]` partial / scaffold · `[ ]` not started  
> App: `apps/admin` (`@hublio/admin`) · shared UI: `@hublio/ui`

Related: [00-foundation](./00-foundation.md) · [02-user-workspace](./02-user-workspace.md)

---

## 1. App scaffold

- [x] Next.js 16 + Tailwind + next-intl + `@hublio/ui`
- [x] Mission Control home (Figma layout; org count from `GET /identity/organizations`; health/KPIs muted until metrics API)
- [x] Auth gate for platform admins (JWT + `is_platform_admin`; seed `admin@hublio.local`)
- [x] Admin shell: top bar, nav, org context (from `/auth/me`)
- [x] Feature-first layout `features/<name>/` (mirror web)
- [x] `lib/api/client` (JWT → Go) — reuse pattern từ web; **không** cần BFF orchestration trừ khi admin chạy intents

---

## 2. Ma trận endpoint (admin scope)

### Organizations (JWT → Go)

Ops-core E2E is done: list / get / create / rename / suspend / activate / archive + org workspaces + org users (read-only).

DTO = `id`, `name`, `status`, `created_at`, `updated_at` only (no plan/country in schema).

- [x] `GET /identity/organizations` — list (platform admin); archived (soft-deleted) omitted
- [x] `GET /identity/organizations/:orgId` — org detail (platform admin or same-org)
- [x] `POST /identity/organizations` — create org + default workspace (platform admin; no owner user)
- [x] `PATCH /identity/organizations/:orgId` — rename
- [x] `POST /identity/organizations/:orgId/suspend`
- [x] `POST /identity/organizations/:orgId/activate`
- [x] `POST /identity/organizations/:orgId/archive` — soft archive (`archived` + `deleted_at`)
- [x] `GET /identity/organizations/:orgId/users` — list org users (platform admin or same-org; read-only)
- [ ] Server-side list page / filter / sort (UI is client-side over full array today)
- [ ] Export orgs (no endpoint; list Export button is toast shell only)
- [ ] Invite / create / suspend user from org detail (deferred)

### Connectors — platform catalog (JWT → Go)

- [ ] `POST /integration/connectors` — register connector
- [ ] `GET /integration/connectors` — catalog admin view
- [ ] `GET /integration/connectors/:id`
- [ ] `POST /integration/connectors/:id/enable`
- [ ] `POST /integration/connectors/:id/disable`
- [ ] `POST /integration/connectors/:id/remove`
- [ ] Marketplace / public catalog UI (nếu product cần)

### Workspaces (cross-tenant ops — chỉ khi có quyền platform)

- [x] List workspaces theo org — platform admin bypass on `GET .../workspaces`
- [x] Create workspace for any org as platform admin (`POST .../workspaces`; no auto membership when cross-tenant)
- [x] Force enable / disable workspace (platform admin without membership)
- [ ] Impersonation / support tools — **chỉ nếu product yêu cầu; không tự thêm**

---

## 3. Screens

- [x] Mission Control (shell + org KPI; metrics/workers blocked)
- [x] Organizations list / detail / create / rename / suspend / activate / archive (`DataTable` client-side)
- [x] Organizations detail tabs: **Overview** | **Workspaces** | **Users** (read-only list)
- [x] Organizations create screen (`/organizations/new`)
- [ ] Connector registry (create / enable / disable / remove)
- [ ] Platform health / usage live metrics (blocked nếu chưa có backend)
- [ ] Billing / plans (blocked — ngoài backend hiện tại)

---

## 2.1 Organizations remaining (API + UI)

Ops core (priorities 1–3) is done. Still deferred vs full Figma org console:

| Area | Status |
| --- | --- |
| List Export button | UI shell only (`toast` “not available”) |
| Bulk select / bulk actions | Not built (no bulk API) |
| List columns: plan, country, workspaces, users, last activity | Blocked — no fields / aggregates |
| Detail tabs beyond Overview + Workspaces + Users (Executions, Activity, Billing, Audit) | Deferred |
| Org user invite / CRUD | Deferred (workspace invite remains the path) |
| Org-level workspace roles on Users tab | Out of scope — roles stay workspace-scoped |
| Plan / country / industry / billing fields | Out of schema — keep in §5 |

**Done in ops-core slice:**

1. Platform-admin bypass on org workspaces + Workspaces tab
2. Admin create org + create screen
3. PATCH rename + archive

---

## 4. Gaps ưu tiên (khi bắt đầu Admin phase)

1. [x] Auth + admin shell + org context
2. [x] Org suspend / activate wired end-to-end
3. [x] Org workspaces cross-tenant (platform-admin bypass + Workspaces tab)
4. Connector register + enable/disable/remove UI
5. [x] i18n namespaces (`en` + `vi`) + auth/locale/theme switchers (shared `@hublio/ui`; console header too)
6. Verify theo [00-foundation §5](./00-foundation.md)

---

## 5. Ngoài phạm vi (chờ product / backend)

- [ ] Replay event từ admin
- [ ] Global marketplace publish flow
- [ ] Usage & billing / plan / country columns
- [ ] Cross-org analytics
- [ ] Org export / bulk delete
- [ ] Impersonate / audit / activity detail tabs (until APIs exist)
