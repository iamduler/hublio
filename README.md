# Hublio

Business Integration Platform + Business Orchestration Platform.

Hublio kết nối hệ thống nghiệp vụ và điều phối Intent → Execution qua Canonical Model + Connectors.
Không phải Workflow Engine / ERP / CRM.

Chi tiết kiến trúc: [`AGENTS.md`](AGENTS.md), [`docs/`](docs/), checklist backend [`docs/29-backend-implementation-checklist.md`](docs/29-backend-implementation-checklist.md).

---

## Stack

| Layer | Tech |
| --- | --- |
| API / Worker | Go (`cmd/api`, `cmd/worker`) |
| Database | PostgreSQL |
| Cache / Work queue | Redis |
| Messaging (optional) | RabbitMQ |
| Dashboard (sau) | Next.js |

---

## Project layout

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
scripts/
docs/
```

---

## Prerequisites

* Go 1.25+
* PostgreSQL 16+ (local WSL hoặc Docker)
* Redis (bắt buộc cho cache + work queue)
* Docker / Docker Compose (tuỳ chọn, cho infra hoặc full stack)
* Make

Cài CLI tools một lần:

```bash
make install_tools
```

Cài đặt:

* `migrate` — golang-migrate (postgres tag)
* `sqlc` — generate typed queries

Đảm bảo `$(go env GOPATH)/bin` có trong `PATH` (Makefile đã tự thêm khi chạy `make`).

---

## Configuration

```bash
cp .env.sample .env
```

Biến quan trọng:

| Variable | Meaning |
| --- | --- |
| `DB_*` | Kết nối app tới PostgreSQL |
| `DB_ADMIN_*` | Superuser dùng bởi `scripts/create-db.sh` (tạo role/DB) |
| `REDIS_ADDRESS` | Redis host (work queue + cache) |
| `API_KEY` | Bootstrap API key cho machine routes `/api/v1/health`, queue, … |
| `JWT_SECRET_KEY` / `JWT_ENCRYPT_KEY` | JWT (32-byte key cho encrypt) |
| `ENABLE_API_DOCS` | Bật/tắt Scalar UI tại `/docs` (mặc định bật khi `DEVELOPMENT_MODE=development`) |
| `SERVER_PORT` | HTTP port (default `8080`) |

### Ghi chú WSL + Navicat

* User Navicat thường **không** có `CREATEROLE`.
* Dùng riêng app DB: `DB_USER=hublio`, `DB_NAME=hublio`.
* `DB_ADMIN_USER=postgres` với password trống → script fallback `sudo -u postgres` (peer auth trên Ubuntu/WSL).

---

## Local development

### 1) Database

**Postgres trên WSL (đang dùng):**

```bash
make db_create      # tạo role + database nếu chưa có
make migrate_up     # chạy migrations
make migrate_status # xem version hiện tại
```

Hoặc gộp:

```bash
make db_setup
```

**Postgres / Redis / RabbitMQ qua Docker (không chạy app trong container):**

```bash
# Trong .env cho Compose: DB_HOST=db, REDIS_ADDRESS=redis:6379 (từ trong network)
# Từ host API: thường map port và dùng localhost
make noapp
make db_setup
```

### 2) Redis

Work queue + cache cần Redis. Ví dụ:

```bash
# Docker Redis (hoặc máy đã cài redis-server)
docker run -d --name redis -p 6379:6379 redis:8.0-alpine
```

Đặt `REDIS_ADDRESS=localhost:6379` trong `.env` khi chạy API/worker trên host.

### 3) Generate SQL / build / test

```bash
make sqlc
make vet
make test
make build
make check          # vet + test + build
```

### 4) Chạy API + Worker

Hai terminal:

```bash
make server
make worker
```

Health:

```bash
curl -s http://localhost:8080/health

### API docs (Scalar — gần Scramble)

Khi `DEVELOPMENT_MODE=development` (hoặc `ENABLE_API_DOCS=true`):

* UI: [http://localhost:8080/docs](http://localhost:8080/docs)
* Spec: [http://localhost:8080/docs/openapi.yaml](http://localhost:8080/docs/openapi.yaml)

Spec nguồn / quy ước cập nhật: `api/openapi/openapi.yaml` — mỗi lần đổi HTTP route phải cập nhật file này cùng thay đổi (xem `AGENTS.md` → OpenAPI / API Docs). Không dùng codegen trừ khi được yêu cầu rõ.
curl -s http://localhost:8080/ready
```

Enqueue job kiểm tra worker (cần `API_KEY` trong `.env`):

```bash
make enqueue_health
# tương đương:
curl -sS -X POST -H "X-API-KEY: $API_KEY" http://localhost:8080/api/v1/platform/queue/health
```

### 5) Full stack Docker (API + Worker + infra)

```bash
# .env: DB_HOST=db, REDIS_ADDRESS=redis:6379, RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
make dev            # docker compose.dev up --build
# hoặc
make prod           # docker compose.prod up -d --build
make logs_prod
make stop_prod
```

Vào shell container API:

```bash
make bash
```

---

## Makefile cheat sheet

| Command | Mô tả |
| --- | --- |
| `make install_tools` | Cài `migrate`, `sqlc` |
| `make db_create` | Tạo role/DB PostgreSQL |
| `make migrate_up` | Apply migrations |
| `make migrate_down` | Rollback 1 step |
| `make migrate_status` | In version hiện tại |
| `make db_setup` | `db_create` + `migrate_up` |
| `make sqlc` | Generate Go từ SQL |
| `make server` | Chạy API |
| `make worker` | Chạy worker (Redis queue) |
| `make check` | `vet` + `test` + `build` |
| `make build` | Binary `bin/api`, `bin/worker` |
| `make enqueue_health` | Push job `platform.health` |
| `make noapp` | Chỉ infra Docker |
| `make dev` / `make prod` | Compose đầy đủ |
| `make stop_noapp` / `make stop_prod` | Dừng compose |

---

## Deployment (Version 1)

Nguyên tắc (xem [`docs/25-deployment-guide.md`](docs/25-deployment-guide.md)):

* Modular monolith: một binary API + process Worker riêng.
* PostgreSQL là source of truth; Redis không phải SoT.
* Work queue thuộc Platform Infrastructure (Redis → Worker), **không** thuộc Event Platform.

### Recommended flow

```text
1. Prepare .env (secrets, DB, Redis)
2. make install_tools
3. Provision Postgres + Redis
4. make db_setup
5. make build   (hoặc image Docker)
6. Start api + worker
7. Verify /health + /ready
```

### Production compose

```bash
cp .env.sample .env
# điền JWT_*, DB_*, REDIS_*, API_KEY, ...

make prod
make logs_prod

# migrate từ host (nếu DB expose) hoặc trong container:
make migrate_up
```

### Manual binary deploy

```bash
make build
./bin/api
./bin/worker
```

Chạy sau reverse proxy (Nginx), expose health/readiness cho orchestration:

* `GET /health` — liveness
* `GET /ready` — Postgres + Redis

### Rollback migration

```bash
make migrate_down
# hoặc
make migrate_goto version=N
```

---

## Useful docs

| Doc | Nội dung |
| --- | --- |
| [`AGENTS.md`](AGENTS.md) | Engineering constitution |
| [`docs/01-product-definition.md`](docs/01-product-definition.md) | Product scope |
| [`docs/03-platform-architecture.md`](docs/03-platform-architecture.md) | Component architecture |
| [`docs/20-database-schema.dbml`](docs/20-database-schema.dbml) | Schema |
| [`docs/25-deployment-guide.md`](docs/25-deployment-guide.md) | Deployment principles |
| [`docs/29-backend-implementation-checklist.md`](docs/29-backend-implementation-checklist.md) | Backend implementation phases |

---

## License

See [`LICENSE`](LICENSE).
