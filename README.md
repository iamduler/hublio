# Hublio

Business Integration Platform + Business Orchestration Platform.

Hublio kết nối hệ thống nghiệp vụ và điều phối Intent → Execution qua Canonical Model + Connectors.  
Không phải Workflow Engine / ERP / CRM.

Chi tiết kiến trúc: [`AGENTS.md`](AGENTS.md), [`docs/`](docs/).

---

## Stack

| Layer | Tech |
| --- | --- |
| API / Worker | Go (`apps/api/cmd/api`, `apps/api/cmd/worker`) |
| Database | PostgreSQL |
| Cache / Work queue | Redis |
| Messaging (optional) | RabbitMQ |
| Frontend | Next.js 16 (`apps/web`, `apps/admin`) |
| Shared TS | `packages/ui`, `packages/config`, `packages/sdk` |

---

## Project layout

```text
apps/
  api/          # Go backend (module path: hublio)
    cmd/api/
    cmd/worker/
    internal/
    migrations/
  web/          # @hublio/web — user workspace
  admin/        # @hublio/admin — admin scaffold
packages/
  ui/           # @hublio/ui
  config/       # @hublio/config
  sdk/          # @hublio/sdk (OpenAPI types)
api/openapi/    # OpenAPI source of truth
deploy/         # Dockerfiles + compose
system/         # redis.conf, …
scripts/
docs/
```

---

## Prerequisites

| Tool | Version |
| --- | --- |
| Go | 1.25+ |
| Node.js | 20+ |
| pnpm | 10+ |
| Make | any |
| PostgreSQL | 16+ |
| Redis | 7/8 |
| Docker Compose | optional |

Enable pnpm once:

```bash
corepack enable
corepack prepare pnpm@10.32.1 --activate
```

---

## Setup & run — step by step

Làm lần lượt từ **root** của repo.

### 1. Backend `.env`

```bash
cp .env.sample .env
```

Chỉnh `.env` cho chạy trên **host** (không trong Docker):

```env
DEVELOPMENT_MODE=development
SERVER_PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=hublio
DB_PASSWORD=hublio
DB_NAME=hublio
DB_SSLMODE=disable

DB_ADMIN_USER=postgres
DB_ADMIN_PASSWORD=
DB_ADMIN_DB=postgres

REDIS_ADDRESS=localhost:6379
REDIS_DB=0

API_KEY=dev-api-key-change-me
JWT_SECRET_KEY=hublio-jwt-secret-key
JWT_ENCRYPT_KEY=12345678901234567890123456789012
CREDENTIAL_ENCRYPTION_KEY=dev-only-insecure-32-byte-key!!!

FRONTEND_URL=http://localhost:3000
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

> Nếu chạy `make server` trên host: dùng `localhost`.  
> Hostname `db` / `redis` / `rabbitmq` chỉ dùng khi process nằm trong Docker Compose.

### 2. Cài Go tools (một lần)

```bash
make install_tools
```

### 3. Start Postgres + Redis

**Option A — Docker infra only:**

```bash
make noapp
```

**Option B — đã có sẵn trên host:** đảm bảo Postgres + Redis đang chạy. Redis mẫu:

```bash
docker run -d --name redis -p 6379:6379 redis:8.0-alpine
```

### 4. Tạo DB + migrate

```bash
make db_setup
make migrate_status
```

### 5. Start API + Worker

```bash
# Terminal 1
make server

# Terminal 2
make worker
```

Verify:

```bash
curl -s http://localhost:8080/health
curl -s http://localhost:8080/ready
```

Smoke queue (cần `API_KEY`):

```bash
make enqueue_health
```

### 6. Cài frontend packages

```bash
pnpm install
```

### 7. Frontend env

```bash
cp apps/web/.env.example apps/web/.env.local
```

```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_APP_URL=http://localhost:3000
```

### 8. Start web

```bash
pnpm --filter @hublio/web dev
```

→ http://localhost:3000/en

Admin (tuỳ chọn):

```bash
pnpm --filter @hublio/admin dev
```

→ http://localhost:3001

### 9. Checklist URL

| Service | URL |
| --- | --- |
| Web | http://localhost:3000/en |
| API health | http://localhost:8080/health |
| API ready | http://localhost:8080/ready |
| API docs | http://localhost:8080/docs |
| Admin | http://localhost:3001 |

---

## Mỗi ngày (sau lần setup đầu)

```bash
make noapp                          # nếu dùng Docker infra
make server                         # terminal 1
make worker                         # terminal 2
pnpm --filter @hublio/web dev       # terminal 3
```

---

## API access (hybrid BFF)

Không bắt buộc mọi request qua Next.js.

* **JWT routes** → `lib/api/client` → Go  
* **API-key routes** (intents, executions, events) → `lib/api/bff-client` → Next `app/api/*` → Go  

Chi tiết: [`docs/24-nextjs-architecture.md`](docs/24-nextjs-architecture.md) §8.1.

---

## Makefile cheat sheet

| Command | Mô tả |
| --- | --- |
| `make install_tools` | Cài `migrate`, `sqlc` |
| `make db_setup` | Tạo DB + migrate |
| `make server` | Chạy API |
| `make worker` | Chạy worker |
| `make check` | vet + test + build |
| `make build` | Build binaries vào `apps/api/bin/` |
| `make noapp` | Chỉ infra Docker |
| `make dev` / `make prod` | Full compose |
| `make stop_noapp` / `make stop_prod` | Stop compose |
| `make bash` | Shell container `go-api` |

---

## Frontend notes

* UI dùng chung: `@hublio/ui`
* Types OpenAPI: `@hublio/sdk`

```bash
pnpm --filter @hublio/sdk generate
```

---

## Troubleshooting

### `failed to load ... apps/api/.env`

`make server` chạy với cwd = `apps/api`, nên `godotenv` tìm `.env` ở đó. Root `.env` vẫn được Makefile `export`, nên app vẫn nhận biến.

Hết warning:

```bash
ln -sf ../../.env apps/api/.env
```

### Redis `context deadline exceeded`

1. Redis chưa chạy.  
2. `.env` còn `redis:6379` trong khi chạy `make server` trên host → đổi thành `localhost:6379`.

### `DEVELOPMENT_MODE=production` — terminal gần như trống rồi `exit status 1`

Không phải lỗi `tracelog`. Ở production, log app ghi vào `apps/api/logs/app.log` (JSON).  
Error/Fatal vẫn in ra **stderr** (sau fix logger). Xem thêm:

```bash
tail -n 50 apps/api/logs/app.log
```

SQL tracer (`github.com/jackc/pgx/v5/tracelog`) ghi `apps/api/logs/sql.log` — field `"trace-id"` trống là bình thường khi chưa có request correlation id.

### `address already in use` / port conflict (vd. `:8080`)

Port đang bị process khác giữ (trên máy này thường là **Apache** trên `8080`).

```bash
ss -tlnp | rg ':8080'
```

Cách xử lý khuyến nghị — đổi port API:

```env
# .env / apps/api/.env
SERVER_PORT=8081
```

```env
# apps/web/.env.local
NEXT_PUBLIC_API_URL=http://localhost:8081/api/v1
```

Rồi chạy lại `make server`.

### Frontend không gọi được API

1. `curl http://localhost:8080/health` đã OK chưa?  
2. `apps/web/.env.local` đúng `NEXT_PUBLIC_API_URL` chưa?  
3. Restart `pnpm --filter @hublio/web dev` sau khi sửa env.

---

## Useful docs

| Doc | Nội dung |
| --- | --- |
| [`AGENTS.md`](AGENTS.md) | Engineering constitution |
| [`docs/24-nextjs-architecture.md`](docs/24-nextjs-architecture.md) | Frontend + hybrid BFF |
| [`apps/web/docs/00_START_HERE.md`](apps/web/docs/00_START_HERE.md) | Web start here |
| [`apps/web/docs/CHECKLIST.md`](apps/web/docs/CHECKLIST.md) | UI checklist |
| [`docs/01-product-definition.md`](docs/01-product-definition.md) | Product scope |
| [`docs/03-platform-architecture.md`](docs/03-platform-architecture.md) | Architecture |
| [`docs/20-database-schema.dbml`](docs/20-database-schema.dbml) | Schema |
| [`docs/25-deployment-guide.md`](docs/25-deployment-guide.md) | Deployment |
| [`docs/29-backend-implementation-checklist.md`](docs/29-backend-implementation-checklist.md) | Backend checklist |

---

## License

See [`LICENSE`](LICENSE).
