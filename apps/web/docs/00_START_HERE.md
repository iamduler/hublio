# Start here — Hublio Web (`apps/web`)

1. Read [AGENTS.md](../AGENTS.md) and [DESIGN.md](../DESIGN.md).
2. Frontend architecture (httpOnly JWT proxy): [`docs/24-nextjs-architecture.md`](../../../docs/24-nextjs-architecture.md).
3. Backend OpenAPI: [`api/openapi/openapi.yaml`](../../../api/openapi/openapi.yaml).
4. App structure: [APP_STRUCTURE.md](./APP_STRUCTURE.md).
5. Progress: [CHECKLIST.md](./CHECKLIST.md) — done / partial / remaining work.
6. API ↔ FE wiring: [API_WIRING_CHECKLIST.md](./API_WIRING_CHECKLIST.md) → `00-foundation` / `01-admin-workspace` / `02-user-workspace`.

## Run (from repo root)

```bash
pnpm install
cp apps/web/.env.example apps/web/.env.local   # if present
pnpm --filter @hublio/web dev
```

Default locale prefix: `/en` (also `/vi`).

## API access

- **Auth** → Next `/api/auth/*` (httpOnly JWT cookies).
- **All dashboard Go APIs** → Next `/api/go/*` via `lib/api/client` (server Bearer + `X-Workspace-ID`).

Details: `docs/24-nextjs-architecture.md` §8.1 and `apps/web/AGENTS.md`.
