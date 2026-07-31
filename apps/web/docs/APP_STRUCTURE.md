# App structure (`apps/web`)

```text
app/
  [locale]/
    (public)/          # marketing landing
    (auth)/            # login, register, stubs
    (dashboard)/       # authenticated shell
  api/                 # Next BFF: /api/auth/*, /api/go/[...path], optional aliases
components/
  layouts/             # top bar, sidebar, workspace header
  auth/                # app-specific auth chrome
features/              # feature-first modules (api, hooks, components, …)
i18n/                  # next-intl routing / navigation
lib/
  api/                 # JWT client (`client`), proxy-go helpers, errors, sdk re-exports
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
| All dashboard Go APIs (identity, integration, intents, executions, events) | `lib/api/client` | Next `/api/go` → Go Bearer + `X-Workspace-ID` |
| Auth (login, refresh, me, MFA, …) | `lib/api/auth` | Next `/api/auth/*` (httpOnly cookies) |
