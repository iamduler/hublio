# App structure (`apps/web`)

```text
app/
  [locale]/
    (public)/          # marketing landing
    (auth)/            # login, register, stubs
    (dashboard)/       # authenticated shell
  api/                 # Next BFF routes (intents, executions, events only)
components/
  layouts/             # top bar, sidebar, workspace header
  auth/                # app-specific auth chrome
features/              # feature-first modules (api, hooks, components, …)
i18n/                  # next-intl routing / navigation
lib/
  api/                 # JWT client, BFF helpers, errors, sdk type re-exports
  i18n/
  preferences/
  workspace.ts
messages/{en,vi}/
providers/
proxy.ts               # Next 16 request proxy (locale + auth gate)
```

Shared UI lives in `packages/ui` (`@hublio/ui`), not under `components/ui`.

Navigation must use `@/i18n/navigation` (not `next/link`).

## API modules

| Feature area | Client | Target |
| --- | --- | --- |
| auth, workspaces, connectors, connections, sync-routes, team, api-keys | `lib/api/client` (JWT) | Go `NEXT_PUBLIC_API_URL` |
| intents, executions, events | `lib/api/bff-client` | Next `app/api/*` → Go with `X-API-KEY` |
