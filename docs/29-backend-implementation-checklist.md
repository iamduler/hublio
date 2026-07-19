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
* [x] Worker consume work queue (`platform.health`; `orchestration.execution` — Phase D done)

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

* [x] Connector aggregate metadata (code, vendor, category, version, status SM)
* [x] States: Registered → Enabled ↔ Disabled → Removed (terminal)
* [x] Behaviors + events (`internal/integration/domain/connector.go`, `events.go`)
* [x] Unit tests SM (`internal/integration/domain/connector_test.go`)

## C2. Domain — Connection + Credential

* [x] Connection belongs to Workspace + references Connector
* [x] SM: Draft → Verifying → Active | VerificationFailed → Verifying; Active ↔ Disabled
* [x] Credential child; `domain.RotateCredential` revokes old + increments version
* [x] Invariants: only Active connection usable for Intent (`CanExecuteIntents`)
* [x] Unit tests (`connection_test.go`, `credential_test.go`)

## C3. Application

* [x] RegisterConnector (platform admin / seed) + SeedFakeConnector
* [x] CreateConnection / VerifyConnection / EnableConnection / DisableConnection
* [x] RotateCredential
* [x] Never store secrets plaintext (Application encrypts via `SecretEncryptor` port before Save)

## C4. Infrastructure

* [x] Migrations: connectors, connector_capabilities, connections, credentials (`migrations/000002_integration.up.sql`)
* [x] Repositories + sqlc (`internal/integration/infrastructure/*_repository.go`, `queries/integration.sql`)
* [x] Encryption for credential payload at rest (`AESSecretEncryptor`, `CREDENTIAL_ENCRYPTION_KEY`, JSONB `{"ciphertext": "..."}`)

## C5. Connector Runtime contract (skeleton)

* [x] Interface trong Integration Domain: `Runtime` (Verify / Health / Invoke) + `RuntimeRegistry`
* [x] Input/Output = canonical-ish `map[string]any` only ở boundary platform (no provider DTOs)
* [x] Provider DTO chỉ trong `internal/integration/connectors/<vendor>/` (none yet — Fake only)
* [x] Fake/Noop connector cho test Orchestration (`internal/integration/connectors/fake`)

## C6. Management API

* [x] Connectors list/get + register/enable/disable/remove (`/api/v1/integration/connectors...`)
* [x] Connections create/list/get + verify + enable/disable (`/api/v1/integration/workspaces/:workspaceId/connections...`)
* [x] Credentials rotate (no secret in responses)

**Exit criteria Phase C:** Active Connection + Fake Connector ready for Orchestration.

> Phase C completed 2026-07-17: Domain (Connector/Connection/Credential state machines + table-driven
> unit tests, all passing without DB), Application use cases, Fake Connector Runtime + Registry,
> Postgres migration/sqlc/repositories, AES-GCM credential encryption, Management HTTP API wired in
> `internal/platform/server/server.go`, OpenAPI paths added under `Integration` tag. Individual
> Capability enable/disable HTTP endpoints and SyncRoute are intentionally deferred (not in Phase C
> scope). `go build ./...` and `go test ./...` pass.

---

# 7. Phase D — Orchestration BC (core runtime)

## D1. Migrations Runtime

* [x] intents, executions, execution_steps, execution_snapshots, execution_timelines, idempotency_keys
* [x] Enums khớp Freeze (execution_status **không** có `completed`)
* [x] Indexes theo DBML

## D2. Domain — Intent

* [x] Intent aggregate + SM: Submitted → Accepted | Rejected | Expired
* [x] Value objects: resource, operation, connection, payload, status
* [x] Accepted Intent immutable
* [x] Unit tests

## D3. Domain — Execution

* [x] Execution aggregate + SM: Created → Queued → Running → Succeeded | Failed | Cancelled | Expired | DeadLetter
* [x] Failed → Queued (retry) | DeadLetter
* [x] Steps sequential v1: validate → transform_request → invoke_connector → transform_response → publish_event
* [x] Context, Timeline, Snapshot, Result
* [x] Unit tests mọi transition bất hợp lệ bị từ chối

## D4. Application — Intent Processor + Orchestration

* [x] SubmitIntent (idempotency key)
* [x] Validate + resolve Connection/Capability
* [x] Create Execution (không expose Execution create API)
* [x] Enqueue execution job (Platform Infrastructure queue)
* [x] Worker: claim job → run steps → update state → publish runtime events
* [x] Retry / Cancel use cases
* [ ] Timeout use case (deferred — no scheduler/deadline sweep yet)
* [ ] Replay use case (**deferred**: `executions.intent_id` is UNIQUE in v1 schema, so a
  second Execution for the same Intent is not representable; Replay needs either a schema
  change or a new Intent. `RetryExecution` covers the "run it again" need for Phase D by
  re-running the same Execution row: Failed → Queued → re-enqueued.)

## D5. Infrastructure

* [x] Repositories runtime tables (`internal/orchestration/infrastructure`)
* [x] Snapshot storage (JSONB) immutable (append-only, `ON CONFLICT (id) DO NOTHING`)
* [x] Queue job payload: execution_id, intent_id, organization_id, workspace_id, correlation_id

## D6. Platform API (Intent)

* [x] `POST` Intent (business entry) — `/api/v1/intents`
* [x] `GET` Intent / Execution status (tracking; client không tạo Execution)
* [x] Idempotency headers (`Idempotency-Key`, Postgres `idempotency_keys` source of truth)
* [x] Auth: API Key (Workspace-scoped) — simplest option meeting exit criteria; a JWT +
  workspace-membership variant for the same routes is deferred

**Exit criteria Phase D:** Submit Intent → Execution Succeeded với Fake Connector end-to-end.

> Phase D completed 2026-07-17: Migration `000003_orchestration` (intents, executions,
> execution_steps, execution_snapshots, execution_timelines, idempotency_keys) + sqlc queries;
> Domain (`Intent`, `Execution` aggregates with state machines, `DefaultStepTypes` 5-step
> pipeline, table-driven unit tests, all pass without DB); Application (`SubmitIntent`,
> `RunExecution`, `GetIntent`, `GetExecution`, `CancelExecution`, `RetryExecution`,
> `ConnectionGateway`/`ConnectorGateway` ports); Infrastructure (Postgres repositories,
> `ConnectionGateway`/`ConnectorGateway` adapters wrapping Integration + Identity,
> `JobEnqueuer` on the Platform Redis queue); Interfaces (`POST /api/v1/intents`,
> `GET /api/v1/intents/:intentId`, `GET /api/v1/executions/:executionId`,
> `POST /api/v1/executions/:executionId/cancel`, `POST /api/v1/executions/:executionId/retry`,
> all API-Key/Workspace-scoped) wired in `internal/platform/server/server.go`;
> `cmd/worker/main.go` now consumes `orchestration.execution` jobs (own composition root,
> independent from the server package) alongside the existing `platform.health` no-op.
> `transform_request`/`transform_response` steps are intentionally passthrough — real
> Canonical↔Canonical mapping lands in Phase E. `go build ./...`, `go vet ./...`, and
> `go test ./...` all pass. OpenAPI updated under a new `Orchestration` tag. Timeout and
> Replay use cases are deferred (see D4 notes).
>
> **Enqueue-after-commit:** queue jobs are enqueued only after the DB transaction commits
> (HTTP handler / worker), avoiding a race where the worker reads uncommitted rows.
>
> **Smoke verified 2026-07-17:** register → Fake Connection verify → API key →
> `POST /api/v1/intents` → worker → Execution `succeeded`.

**Smoke steps (Postgres + Redis required):**

```text
1. migrate up (migrations/000001..000003)
2. POST /api/v1/auth/register, POST /api/v1/auth/login -> JWT
3. POST /api/v1/workspaces/:workspaceId/api-keys (JWT) -> capture plaintext key once
4. POST /api/v1/integration/workspaces/:workspaceId/connections against the Fake connector,
   then verify it (-> Active)
5. POST /api/v1/intents  with header  X-API-KEY: <key>  and  Idempotency-Key: <uuid>
   body: {"connection_id": "<connection-id>", "capability": "fake.echo", "payload": {"foo": "bar"}}
6. go run ./cmd/worker   (consumes orchestration.execution)
7. GET /api/v1/executions/:executionId with X-API-KEY -> poll until status = "succeeded"
```

---

# 8. Phase E — Transformation BC

## E1. Domain / Application

* [x] Transform pipeline Canonical → Canonical only (`internal/transformation/domain`: `Document`, `Operation`, `Pipeline`)
* [x] Capabilities: field rename, type convert, timezone, currency normalize, defaults, validation (`RenameField`, `ConvertType`, `NormalizeTimezone`, `NormalizeCurrency`, `SetDefault`, `ValidateRequired`)
* [x] **Cấm** Provider DTO trong package này (verified: no `internal/integration` import in `internal/transformation`)
* [x] `OperationSpec` + `BuildPipeline` factory so callers (Orchestration) can describe a Pipeline as data, without Domain knowing HTTP/DB
* [x] `Services.Transform` (`internal/transformation/application`) — no repositories; in-memory engine only

## E2. Integration with Orchestration steps

* [x] Step `transform_request` / `transform_response` gọi Transformation use cases (`run_execution.go` calls `Services.transformer()` instead of passthrough)
* [x] `Transformer` port on Orchestration Application + `TransformerAdapter` in Orchestration Infrastructure (wraps `transformationapp.Services`, never leaks Transformation types into Orchestration Domain/Application)
* [x] Wired in both composition roots: `internal/platform/server/server.go` and `cmd/worker/main.go` (worker runs the steps)
* [x] Tests: normalize invoice-like canonical fixture (không MISA DTO) — table-driven Domain tests per operation + full fixture (`internal/transformation/domain/pipeline_test.go`), Application-level fixture test (`internal/transformation/application/transform_test.go`), adapter capability-routing test (`internal/orchestration/infrastructure/transformer_adapter_test.go`)

**Exit criteria Phase E:** Step transform chạy trong Execution path, không đụng Provider DTO.

> Phase E completed 2026-07-18: Domain pipeline (`Document`/`Operation`/`Pipeline` +
> `RenameField`/`ConvertType`/`NormalizeTimezone`/`NormalizeCurrency`/`SetDefault`/`ValidateRequired`,
> all table-driven unit tests, no DB/HTTP dependency); `OperationSpec`/`BuildPipeline` factory;
> built-in `DefaultRequestPipelineSpec()`/`DefaultResponsePipelineSpec()` invoice normalization;
> Application `Services.Transform` (no repositories — pure in-memory engine); Orchestration
> `Transformer` port + `TransformerAdapter` (Infrastructure) applies the default invoice
> pipeline only when `capability` looks invoice-like (`strings.Contains(lower(capability),
> "invoice")`), otherwise an empty/identity spec — so the Fake connector's `fake.echo`
> capability still passes its payload through unchanged end-to-end. `run_execution.go`'s
> `transform_request`/`transform_response` steps now call the Transformer instead of the
> Phase D passthrough. Wired in `internal/platform/server/server.go` and `cmd/worker/main.go`
> (the worker is what actually runs Execution steps). No new HTTP routes, no OpenAPI change,
> no new migrations/tables (Transformation stays a pure in-memory engine per docs/06). `go
> build ./...`, `go vet ./...`, and `go test ./...` all pass.

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
