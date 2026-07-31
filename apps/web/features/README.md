# Features

Feature-first modules live here (see Hublio `docs/24-nextjs-architecture.md`).

Suggested layout per feature:

```text
features/<name>/
  api.ts            # or api/
  components/
  hooks.ts          # or hooks/
  schemas.ts
  types.ts
```

Features must not define App Router routes — pages stay under `app/`.

## Calling the API

- Use `lib/api/client` → Next `/api/go` (Bearer JWT + `X-Workspace-ID`) for all
  dashboard Go APIs (identity, integration, intents, executions, events).
- Use `lib/api/auth` → Next `/api/auth/*` for login/refresh/me/MFA (httpOnly cookies).
- Import DTO types from `@hublio/sdk` / `lib/api/sdk` when they map 1:1 to OpenAPI.