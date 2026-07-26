<!-- BEGIN:nextjs-agent-rules -->
# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` before writing any code. Heed deprecation notices.
<!-- END:nextjs-agent-rules -->

# Hublio Web (`apps/web`) — AI notes

- Product: Hublio user workspace (Integration + Orchestration).
- Backend: Go API in this monorepo at `apps/api` — OpenAPI `api/openapi/openapi.yaml`.
- Frontend architecture: `docs/24-nextjs-architecture.md` (§8.1 httpOnly JWT proxy).
- Business rules belong on the Go backend. This app is presentation + API integration only.
- Do not invent API endpoints. Only call paths documented in OpenAPI.
- Stack: Next.js 16, React 19, Tailwind v4, next-intl, TanStack Query, shadcn via `@hublio/ui`.
- Auth: httpOnly JWT cookies `hublio_session` / `hublio_refresh` set by `/api/auth/*`. Dashboard soft-gated in `proxy.ts`.

## API access (do not change without explicit approval)

| Browser calls | Mechanism | Routes |
| --- | --- | --- |
| Next `/api/auth/*` | sets httpOnly cookies | login, register, logout, session |
| Next `/api/go/*` via `lib/api/client` | server Bearer + `X-Workspace-ID` | identity, integration, intents, executions, events |

- Browser never reads JWT or workspace API keys.
- Go still accepts `X-API-KEY` for machine clients on orchestration/events.
- Prefer `@hublio/sdk` types for request DTOs; keep hand-written feature `api.ts` clients (no full typed SDK client rewrite unless requested).
