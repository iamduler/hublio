<!-- BEGIN:nextjs-agent-rules -->
# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` before writing any code. Heed deprecation notices.
<!-- END:nextjs-agent-rules -->

# Hublio Web (`apps/web`) — AI notes

- Product: Hublio user workspace (Integration + Orchestration).
- Backend: Go API in this monorepo at `apps/api` — OpenAPI `api/openapi/openapi.yaml`.
- Frontend architecture: `docs/24-nextjs-architecture.md` (hybrid BFF rule).
- Business rules belong on the Go backend. This app is presentation + API integration only.
- Do not invent API endpoints. Only call paths documented in OpenAPI.
- Stack: Next.js 16, React 19, Tailwind v4, next-intl, TanStack Query, shadcn via `@hublio/ui`.
- Auth: JWT Bearer in cookies `hublio_session` / `hublio_refresh`. Dashboard soft-gated in `proxy.ts`.

## API access (do not change without explicit approval)

Hybrid — **not** “everything through Next.js”:

| Browser calls | Mechanism | Routes |
| --- | --- | --- |
| Go API directly | `lib/api/client` + JWT | auth, workspaces, connectors, connections, sync-routes, team, api-keys |
| Next BFF then Go | `lib/api/bff-client` → `app/api/*` → `X-API-KEY` | intents, executions, events |

- BFF exists to keep workspace API keys server-side. Do not expose them to the browser.
- Do not add BFF proxies for JWT CRUD unless there is a new secret or orchestration need.
- Prefer `@hublio/sdk` types for DTOs; keep hand-written feature `api.ts` clients (no full typed SDK client rewrite unless requested).
