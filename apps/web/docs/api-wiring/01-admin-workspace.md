# 01 — Admin Workspace (`apps/admin`)

> Platform / org administration console.  
> Legend: `[x]` done · `[~]` partial / scaffold · `[ ]` not started  
> App: `apps/admin` (`@hublio/admin`) · shared UI: `@hublio/ui`

Related: [00-foundation](./00-foundation.md) · [02-user-workspace](./02-user-workspace.md)

---

## 1. App scaffold

- [x] Next.js 16 + Tailwind + next-intl + `@hublio/ui`
- [x] Placeholder home page
- [ ] Auth gate for platform admins (JWT + role check)
- [ ] Admin shell: top bar, nav, org context
- [ ] Feature-first layout `features/<name>/` (mirror web)
- [ ] `lib/api/client` (JWT → Go) — reuse pattern từ web; **không** cần BFF orchestration trừ khi admin chạy intents

---

## 2. Ma trận endpoint (admin scope)

### Organizations (JWT → Go)

- [ ] `GET /identity/organizations/:orgId` — org detail trong admin
- [ ] `POST /identity/organizations/:orgId/suspend`
- [ ] `POST /identity/organizations/:orgId/activate`
- [ ] Org list / search (cần endpoint list nếu chưa có)

### Connectors — platform catalog (JWT → Go)

- [ ] `POST /integration/connectors` — register connector
- [ ] `GET /integration/connectors` — catalog admin view
- [ ] `GET /integration/connectors/:id`
- [ ] `POST /integration/connectors/:id/enable`
- [ ] `POST /integration/connectors/:id/disable`
- [ ] `POST /integration/connectors/:id/remove`
- [ ] Marketplace / public catalog UI (nếu product cần)

### Workspaces (cross-tenant ops — chỉ khi có quyền platform)

- [ ] List workspaces theo org
- [ ] Force enable / disable workspace (reuse existing endpoints)
- [ ] Impersonation / support tools — **chỉ nếu product yêu cầu; không tự thêm**

---

## 3. Screens (chưa có)

- [ ] Organizations list / detail / suspend-activate
- [ ] Connector registry (create / enable / disable / remove)
- [ ] Platform health / usage (blocked nếu chưa có backend)
- [ ] Billing / plans (blocked — ngoài backend hiện tại)

---

## 4. Gaps ưu tiên (khi bắt đầu Admin phase)

1. Auth + admin shell + org context
2. Org suspend / activate wired end-to-end
3. Connector register + enable/disable/remove UI
4. i18n namespaces `admin/*` (en + vi)
5. Verify theo [00-foundation §5](./00-foundation.md)

---

## 5. Ngoài phạm vi (chờ product / backend)

- [ ] Replay event từ admin
- [ ] Global marketplace publish flow
- [ ] Usage & billing
- [ ] Cross-org analytics
