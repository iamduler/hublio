# 29 - Backend Implementation Checklist

> Product: Hublio
> Version: 1.0
> Status: Approved
> Last Updated: 2026-07-15

---

# 1. Purpose

Checklist triển khai **backend** theo Architecture Freeze.

Dùng để:

* Review trước khi code từng phase
* Theo dõi tiến độ
* Gate merge / PR

Mỗi feature phải đi theo thứ tự:

```text
Domain → Repository Interface → Application → Infrastructure → REST → Tests → Review
```

Nguồn đúng:

* `AGENTS.md`
* `docs/00` … `docs/28`
* `docs/20-database-schema.dbml`
* `docs/30-mvp-usecase-nhanh-misa.md` (MVP product north star: Nhanh.vn → MISA)

---

# 2. Preflight — sẵn sàng code chưa?

Trạng thái hiện tại (scaffold):

* [x] Module `hublio`, `cmd/api`, `cmd/worker`
* [x] BC packages: identity / integration / orchestration / transformation / events / platform
* [x] Platform infra: config, postgres, redis, auth JWT, middleware, messaging
* [x] Migration Identity (organizations, workspaces, users, workspace_users, api_keys)
* [x] Domain logic thực tế — Phase B Identity (Phase C+)
* [x] sqlc queries (identity select stubs + generate)
* [x] Management / Platform API — Phase B Identity auth + management
* [x] Worker consume work queue (`platform.health`; Execution jobs Phase D)

Gate Preflight (review trước Phase A):

* [ ] Product identity đã chốt: Business Integration + Orchestration (không Workflow Engine)
* [ ] Hierarchy: Organization → Workspace → API Key
* [ ] Mapping: Transformation = Canonical→Canonical; Connector = Canonical↔Provider DTO
* [ ] BC tên `platform` (không Administration)
* [ ] Execution terminal thành công = `succeeded` (không `completed`)
* [ ] Work queue thuộc Platform Infrastructure; Event Platform chỉ publish
* [ ] Local stack chạy được: Postgres, Redis, (queue), migrate up, `go build ./...`

---

# 3. Architecture Gates (mọi PR backend)

Mỗi PR phải tick trước khi merge:

* [ ] Không thêm BC / Aggregate / Runtime abstraction ngoài Freeze
* [ ] Dependency: Interfaces → Application → Domain ← Infrastructure
* [ ] Domain không import postgres / redis / gin / json provider DTO
* [ ] Business rules ở Domain
* [ ] Handler chỉ validate request + gọi use case + map response
* [ ] UUID v7 tạo ở Application (repo không generate ID)
* [ ] Một use case = một transaction boundary
* [ ] Events chỉ publish sau commit thành công
* [ ] Không log secrets / tokens / API keys / passwords
* [ ] Mọi operation có tenant: `organization_id` + `workspace_id` (khi áp dụng)
* [ ] Table-driven unit tests cho Domain invariants
* [ ] `go test ./...` và `go build ./...` pass
* [ ] Thêm/sửa/xóa HTTP route → cập nhật `api/openapi/openapi.yaml` cùng PR (không codegen)

---

# 4. Phase A — Platform Foundation (hardening scaffold)

Mục tiêu: infra đủ vững để gắn Domain.

## A1. Persistence & sqlc

* [x] Chốt path: `migrations/` + `internal/platform/persistence/queries` + `sqlc.yml`
* [x] Thêm queries tối thiểu cho Identity tables (hoặc placeholder readiness check)
* [x] `make sqlc` generate vào `internal/platform/persistence/sqlc`
* [x] Wrapper repo Infrastructure gọi sqlc — **không** để Domain phụ thuộc sqlc types
* [x] Transaction helper ở Application/platform (begin/commit/rollback), repo không tự commit

## A2. Cross-cutting platform

* [x] Correlation ID / Trace ID xuyên request → execution (sau này)
* [x] `apperr` map sang HTTP status nhất quán
* [x] Auth middleware: JWT user path + API Key path (Workspace-scoped)
* [x] API Key: hash at rest, chỉ trả plaintext một lần lúc create
* [x] Idempotency middleware/store skeleton (Redis/Postgres) — dùng cho Intent API sau
* [x] Structured logging: `correlation_id`, `request_id`, `organization_id`, `workspace_id`

## A3. Work queue (Platform Infrastructure)

* [x] Chọn implementation v1 (ví dụ Redis queue hoặc RabbitMQ queue) — **không** gắn Event Platform
* [x] Interface: `Enqueue(job)`, `Consume(handler)`, visibility timeout / ack
* [x] Worker `cmd/worker` consume job types: bắt đầu với no-op / health job
* [x] Document rõ: Event Platform ≠ Work Queue

## A4. CI backend tối thiểu

* [x] `go vet`, `go test`, `go build`
* [x] Optional: golangci-lint
* [x] Migration dry-run / schema check trên CI

**Exit criteria Phase A:** API health + migrate + worker idle + CI xanh.

> Phase A completed 2026-07-15: `/health` + `/ready`, Redis work queue + `platform.health`, sqlc + `WithinTransaction`, API key port (static/stub), idempotency Redis store, `.github/workflows/ci.yml`. golangci-lint deferred; CI runs vet/test/build.

---

# 5. Phase B — Identity BC

Thứ tự Aggregate: **Organization → Workspace → User/Membership → API Key**.

## B1. Domain — Organization

* [x] Aggregates / entities / value objects / status enums
* [x] Invariants: name unique (app-level + DB), suspend blocks new intents (rule ready)
* [x] Behaviors: Create, Update, Suspend, Activate, Archive
* [x] Domain events: OrganizationCreated / Updated / Suspended / Activated
* [x] Unit tests table-driven

## B2. Domain — Workspace

* [x] Aggregate + environment + status
* [x] Invariants: belongs to one Organization; disabled cannot execute intents
* [x] Behaviors: Create, Update, Enable, Disable
* [x] Child: API Key (thuộc Workspace)
* [x] Domain events: WorkspaceCreated / Enabled / Disabled
* [x] Unit tests

## B3. Domain — User & Workspace membership

* [x] User under Organization
* [x] `workspace_users` role: owner / admin / member
* [x] Auth behaviors: register/invite (scope v1), password hash interface (port), login
* [x] Unit tests cho status transitions

## B4. Domain — API Key (Workspace)

* [x] Create / Disable / Rotate
* [x] Store hash + prefix; never persist plaintext
* [x] Invariants: workspace-scoped; disabled key rejected
* [x] Domain events: ApiKeyCreated / Disabled / Rotated
* [x] Unit tests

## B5. Application (Identity use cases)

* [x] CreateOrganization (+ first owner user optional bootstrap)
* [x] CreateWorkspace
* [x] Invite/AddUserToWorkspace
* [x] CreateApiKey / DisableApiKey / RotateApiKey
* [x] Login / IssueToken / Logout (revoke)
* [x] UUID v7 generation
* [x] Transaction per use case
* [x] Publish domain/system events after commit

## B6. Infrastructure (Identity)

* [x] Postgres repositories implementing Domain ports
* [x] sqlc queries: orgs, workspaces, users, workspace_users, api_keys
* [x] Password hasher (bcrypt/argon2) ở infrastructure
* [x] Map DB rows ↔ Domain (không leak sqlc ra Application handlers)

## B7. Interfaces (Management API — Identity)

* [x] Routes dưới `/api/v1/...` (Management)
* [x] Organizations CRUD/lifecycle
* [x] Workspaces CRUD/lifecycle
* [x] Users / membership
* [x] API Keys create (return secret once) / list / disable / rotate
* [x] Auth endpoints: login, refresh, logout
* [x] OpenAPI stubs sync với `docs/23` (cập nhật dần) — `api/openapi/openapi.yaml` + Scalar `/docs`
* [ ] Integration tests API (auth + tenant isolation)

**Exit criteria Phase B:** tạo Org → Workspace → API Key → gọi API bằng API Key/JWT thành công.

> Phase B core completed 2026-07-15: Domain + use cases + Postgres repos + Management/Auth HTTP. Smoke verified: register → login JWT → create API key → `GET /api/v1/health` with `X-API-KEY`. OpenAPI sync + formal integration tests deferred. Logout (refresh revoke) implemented; refresh-token rotate endpoint deferred.

---

# 6. Phase C — Integration BC (không provider thật trước)

## C1. Domain — Connector

* [ ] Connector aggregate metadata (code, vendor, category, version, status SM)
* [ ] States: Registered → Enabled → Disabled → Removed
* [ ] Behaviors + events theo `docs/14` / `docs/18`
* [ ] Unit tests SM

## C2. Domain — Connection + Credential

* [ ] Connection belongs to Workspace + references Connector
* [ ] SM: Draft → Verifying → Active | VerificationFailed → Disabled
* [ ] Credential child; rotate credentials
* [ ] Invariants: only Active connection usable for Intent
* [ ] Unit tests

## C3. Application

* [ ] RegisterConnector (platform admin / seed)
* [ ] CreateConnection / VerifyConnection / Activate / Disable
* [ ] RotateCredential
* [ ] Never store secrets plaintext

## C4. Infrastructure

* [ ] Migrations: connectors, connector_capabilities, connections, credentials (từ DBML)
* [ ] Repositories + sqlc
* [ ] Encryption for credential payload at rest

## C5. Connector Runtime contract (skeleton)

* [ ] Interface trong Integration: Auth / Invoke / Health / VerifyWebhook
* [ ] Input/Output = **Canonical DTOs only** ở boundary platform
* [ ] Provider DTO chỉ trong `internal/integration/connectors/<vendor>/`
* [ ] Fake/Noop connector cho test Orchestration

## C6. Management API

* [ ] Connectors list/get
* [ ] Connections CRUD + verify + disable
* [ ] Credentials rotate (no secret in responses)

**Exit criteria Phase C:** Active Connection + Fake Connector ready for Orchestration.

---

# 7. Phase D — Orchestration BC (core runtime)

## D1. Migrations Runtime

* [ ] intents, executions, execution_steps, execution_snapshots, execution_timelines, idempotency_keys
* [ ] Enums khớp Freeze (execution_status **không** có `completed`)
* [ ] Indexes theo DBML

## D2. Domain — Intent

* [ ] Intent aggregate + SM: Submitted → Accepted | Rejected | Expired
* [ ] Value objects: resource, operation, connection, payload, status
* [ ] Accepted Intent immutable
* [ ] Unit tests

## D3. Domain — Execution

* [ ] Execution aggregate + SM: Created → Queued → Running → Succeeded | Failed | Cancelled | Expired | DeadLetter
* [ ] Failed → Queued (retry) | DeadLetter
* [ ] Steps sequential v1: validate → transform_request → invoke_connector → transform_response → publish_event
* [ ] Context, Timeline, Snapshot, Result
* [ ] Unit tests mọi transition bất hợp lệ bị từ chối

## D4. Application — Intent Processor + Orchestration

* [ ] SubmitIntent (idempotency key)
* [ ] Validate + resolve Connection/Capability
* [ ] Create Execution (không expose Execution create API)
* [ ] Enqueue execution job (Platform Infrastructure queue)
* [ ] Worker: claim job → run steps → update state → publish runtime events
* [ ] Retry / Timeout / Cancel / Replay use cases
* [ ] Replay = new Execution, reuse Intent/context/snapshots rules

## D5. Infrastructure

* [ ] Repositories runtime tables
* [ ] Snapshot storage (JSONB) immutable
* [ ] Queue job payload: execution_id, correlation_id, tenant ids

## D6. Platform API (Intent)

* [ ] `POST` Intent (business entry)
* [ ] `GET` Intent / Execution status (tracking; client không tạo Execution)
* [ ] Idempotency headers
* [ ] Auth: API Key (Workspace) / JWT with workspace context

**Exit criteria Phase D:** Submit Intent → Execution Succeeded với Fake Connector end-to-end.

---

# 8. Phase E — Transformation BC

## E1. Domain / Application

* [ ] Transform pipeline Canonical → Canonical only
* [ ] Capabilities: field rename, type convert, timezone, currency normalize, defaults, validation
* [ ] **Cấm** Provider DTO trong package này

## E2. Integration with Orchestration steps

* [ ] Step `transform_request` / `transform_response` gọi Transformation use cases
* [ ] Tests: normalize invoice-like canonical fixture (không MISA DTO)

**Exit criteria Phase E:** Step transform chạy trong Execution path, không đụng Provider DTO.

---

# 9. Phase F — Events BC + Observability wiring

## F1. Event Platform (internal)

* [ ] Event model: runtime / business / system
* [ ] Persist append-only `events`
* [ ] Publish after commit
* [ ] Delivery at-least-once; subscribers idempotent
* [ ] **Không** dùng Event Platform làm work queue Execution

## F2. Audit

* [ ] `audit_logs` cho API key create, connection changes, replay, login
* [ ] Không ghi secrets

## F3. Observability

* [ ] Metrics: execution success/fail, queue depth, latency
* [ ] Traces: request → intent → execution → step → connector
* [ ] Timeline API / query cho dashboard sau

**Exit criteria Phase F:** ExecutionSucceeded/Failed event persisted; audit basic works.

---

# 10. Phase G — First real Connector (e-invoice)

Chỉ sau Fake path xanh.

## G1. Connector package

* [ ] `internal/integration/connectors/<vendor>/` (ví dụ misa)
* [ ] Manifest + capabilities
* [ ] Auth + health
* [ ] Canonical Invoice → Provider DTO → HTTP → Canonical Response
* [ ] Error translation sang platform errors

## G2. End-to-end

* [ ] Connection verify với sandbox credentials
* [ ] Intent CreateInvoice → Succeeded / Failed có snapshot
* [ ] Replay / Retry tested
* [ ] Secrets không lộ log/API

**Exit criteria Phase G:** một Intent thật thành công trên sandbox provider.

---

# 11. Suggested sprint order

| Sprint | Scope | Deliverable |
| ------ | ----- | ----------- |
| S0 | Phase A | Hardened platform + queue + CI |
| S1 | Phase B | Identity Management API |
| S2 | Phase C | Connector/Connection + Fake connector |
| S3 | Phase D | Intent → Execution worker path |
| S4 | Phase E+F | Transform + Events/Audit |
| S5 | Phase G | First e-invoice connector |

Không nhảy S5 trước khi S3 xanh.

---

# 12. Per-feature mini checklist (copy vào PR)

```text
## Feature: _______________
BC: _______________  Aggregate: _______________

- [ ] Domain + unit tests
- [ ] Repository interface (Domain)
- [ ] Use case + transaction
- [ ] Infra repo + migration/sqlc (if needed)
- [ ] REST handler + DTO (Interfaces)
- [ ] AuthZ + tenant checks
- [ ] Events after commit (if any)
- [ ] Integration/API test
- [ ] Architecture Gates (section 3) passed
- [ ] Docs touched if contract changed
```

---

# 13. Explicit non-goals (đừng làm trong v1 backend)

* [ ] Workflow / BPMN / Saga / human approval
* [ ] Parallel Steps inside one Execution / dynamic planning
* [ ] Rule engine / AI planning
* [ ] Connector marketplace / hot-load plugins
* [ ] Microservices split
* [ ] CQRS / Event Sourcing frameworks
* [ ] Provider DTO trong Domain / Application / REST
* [ ] Transformation map Provider DTO
* [ ] Event Platform own Execution queue

---

# 14. Definition of Ready → Done (backend milestone)

## Ready to start coding a phase

* [ ] Phase trước đạt Exit criteria
* [ ] Aggregate & SM liên quan đã đọc lại
* [ ] DBML tables đã xác định
* [ ] API surface (Management vs Platform) đã clear

## Done

* [ ] Exit criteria của phase
* [ ] Tests theo `docs/27`
* [ ] Không regression Architecture Gates
* [ ] README/openapi cập nhật nếu public contract đổi

---

# 15. Guiding principle

Prefer the slowest correct layering over the fastest endpoint.

If unsure: Domain first, Adapter last, follow Freeze — never invent a new runtime concept.
