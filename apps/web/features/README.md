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

- Prefer `lib/api/client` (JWT → Go) for Identity / Integration feature APIs.
- Use `lib/api/bff-client` only for intents / executions / events (Next BFF holds
  the workspace API key). See `docs/24-nextjs-architecture.md` §8.1.
- Import DTO types from `@hublio/sdk` / `lib/api/sdk` when they map 1:1 to OpenAPI.