# Start here — Hublio Web (`apps/web`)

1. Read [AGENTS.md](../AGENTS.md) and [DESIGN.md](../DESIGN.md).
2. Frontend architecture (hybrid BFF): [`docs/24-nextjs-architecture.md`](../../../docs/24-nextjs-architecture.md).
3. Backend OpenAPI: [`api/openapi/openapi.yaml`](../../../api/openapi/openapi.yaml).
4. App structure: [APP_STRUCTURE.md](./APP_STRUCTURE.md).
5. Progress: [CHECKLIST.md](./CHECKLIST.md) — done / partial / remaining work.

## Run (from repo root)

```bash
pnpm install
cp apps/web/.env.example apps/web/.env.local   # if present
pnpm --filter @hublio/web dev
```

Default locale prefix: `/en` (also `/vi`).

## API access (frozen)

- **JWT routes** → browser calls Go (`NEXT_PUBLIC_API_URL`) via `lib/api/client`.
- **API-key-only routes** (intents / executions / events) → browser calls Next
  `app/api/*` BFF; server holds the workspace API key and proxies to Go.

Details: `docs/24-nextjs-architecture.md` §8.1 and `apps/web/AGENTS.md`.
