# Hublio

Business Orchestration Platform.

## Architecture

See `AGENTS.md` and `docs/`.

High-level Go layout:

```text
cmd/
  api/
  worker/
internal/
  identity/
  integration/
  orchestration/
  transformation/
  events/
  platform/
migrations/
docs/
```

## Local commands

```bash
make server
make worker
make migrate_up
make sqlc
make build
```
