# Hublio Web (`@hublio/web`)

Next.js 16 user workspace for Hublio (Go API in `apps/api`).

## Stack

- Next.js 16 / React 19
- Tailwind CSS v4 + `@hublio/ui` (shadcn)
- next-intl (`en` default, `vi`)
- TanStack Query
- `@hublio/sdk` (OpenAPI types)

## Setup (from monorepo root)

```bash
pnpm install
pnpm --filter @hublio/web dev
```

Open [http://localhost:3000/en](http://localhost:3000/en).

Set `NEXT_PUBLIC_API_URL` to the Go API base including `/api/v1`
(e.g. `http://localhost:8080/api/v1`).

## API access

Hybrid (see `docs/24-nextjs-architecture.md` §8.1):

- JWT Identity / Integration CRUD → browser → Go
- Intents / Executions / Events → browser → Next BFF → Go (`X-API-KEY` server-side)

## Docs

- [AGENTS.md](./AGENTS.md)
- [docs/00_START_HERE.md](./docs/00_START_HERE.md)
- [docs/CHECKLIST.md](./docs/CHECKLIST.md)
