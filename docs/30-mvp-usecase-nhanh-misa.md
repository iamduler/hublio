# 30 - MVP Use Case #1: Nhanh.vn ↔ MISA

> Product: Hublio
> Version: 1.0
> Status: Approved (Product Decisions + Freeze fan-out clarification applied)
> Last Updated: 2026-07-15

---

# 1. Purpose

MVP north star: đồng bộ đa danh mục giữa **Nhanh.vn** (origin) và **MISA** (destination) qua Hublio, multi-tenant, có webhook + poll, fan-out, reverse update, retry/DLQ, monitoring.

Architecture Freeze thắng về Runtime Model. Fan-out tuần tự/song song được phép dưới dạng **nhiều Execution từ một Intent** (xem §8 và AGENTS.md).

---

# 2. Problem

Khách dùng 2 hệ thống độc lập (CRUD API). Không nói chuyện trực tiếp → Hublio trung gian.

* Origin: Nhanh.vn (webhook và/hoặc pull)
* Destination chính: MISA
* Có thể kèm activity khác (email, ERP) từ cùng một sự kiện nguồn
* Reverse: cập nhật trạng thái ngược lại Nhanh.vn (trong MVP)

---

# 3. Product Decisions (locked)

| # | Topic | Decision |
| - | ----- | -------- |
| 1 | Canonical resources | **Multiple**. Ưu tiên **Invoice** trước; MVP phải mở rộng được nhiều danh mục (Product, Customer, …) giữa 2 hệ thống |
| 2 | Fan-out | **Có**. Hỗ trợ **tuần tự** và **song song** |
| 3 | Conditions | JSON filter trên cấu hình route (không Rule Engine). Toán tử: `>`, `<`, `>=`, `<=`, `eq`, `neq`, `between`, `in`. Kết hợp **AND / OR** (cây điều kiện) |
| 4 | Origin → dest | Xem **§5 đề xuất SyncRoute** (best practice) |
| 5 | Idempotency | Generate theo rule từ business fields (ví dụ `account_id + app_id + record_id`). Xem **§6** |
| 6 | Reverse update | **Có trong MVP** (Nhanh ← status từ kết quả MISA/Hublio) |
| 7 | Auth | Connector Auth cho **Nhanh + MISA** (credential types theo từng provider) |
| 8 | Webhook security | **Locked**: Hublio **tạo** webhook secret, lưu gắn Connection/SyncRoute; Nhanh gửi secret trong **header**; Hublio validate header == secret. Xem §7.1 |
| 9 | Poll watermark | **Locked**: có lưu **dấu vết (watermark/cursor)** của lần poll trước. Xem §7.2 |
| 10 | Retry | Tối đa **3 lần retry**. Delay: **1s → 3s → 10s**. Sau đó **Dead Letter** |

---

# 4. Runtime Flows

## 4.1 Webhook

```text
Nhanh.vn
  → HTTPS webhook → Hublio (validate shared secret in header)
  → Filter (JSON conditions)
  → Intent (idempotent)
  → Fan-out activities (sequential and/or parallel)
       → Execution(s) → Transform C→C → Connector → MISA / Email / ERP
  → Reverse activity → Nhanh.vn status update
```

## 4.2 Poll

```text
Scheduler (frequency on SyncRoute)
  → Pull Nhanh (since watermark/cursor)
  → Filter each record
  → Same Intent / fan-out / reverse path as 4.1
  → Advance watermark only after successful accept (policy below)
```

---

# 5. Origin → Destination — Best Practice Proposal

## 5.1 Why not only `Connection.config`?

`Connection` = một hệ thống đã xác thực trong Workspace (Nhanh **hoặc** MISA).

Gắn origin→dest vào một Connection sẽ:

* Không mô tả được 1 origin → N destinations
* Trộn credentials với routing logic
* Khó tái sử dụng cùng MISA connection cho nhiều nguồn

## 5.2 Recommendation: **SyncRoute** (configuration, Workspace-scoped)

Thêm **configuration entity** (không phải Workflow Engine, không BPMN):

```text
Workspace
  ├── Connection (Nhanh)     ← origin credentials + provider config
  ├── Connection (MISA)      ← destination credentials
  ├── Connection (Email/ERP) ← optional destinations
  └── SyncRoute              ← “luồng đồng bộ” origin → activities
```

### SyncRoute fields (logical)

* `id`, `workspace_id`, `name`, `status` (enabled/disabled)
* `source_connection_id` (origin)
* `resource_types[]` — e.g. `invoice`, `product` (multi-catalog MVP)
* `trigger`: `webhook` | `schedule` | `both`
* `schedule` — interval/cron khi poll
* `filter` — JSON condition tree (§3)
* `idempotency_rule` — template / field list (§6)
* `activities[]` — ordered **groups**:
  * `group_mode`: `sequential` | `parallel`
  * `steps[]`: `{ destination_connection_id, capability, mapping_key }`
* `reverse` — optional activity back to `source_connection_id` after primary success/failure
* `retry_policy` — override defaults (else 1s/3s/10s × 3)
* `watermark` — stored cursor per resource_type for poll (§7.2)

### Runtime mapping (keeps Freeze Intent/Execution)

```text
1 Source event / polled record
    → 1 Intent (business request)
    → N Executions (một Execution / activity step)
         parallel group: enqueue N jobs cùng lúc
         sequential group: enqueue step kế chỉ khi step trước succeeded
```

* Clients / webhooks **không** tạo Execution trực tiếp.
* SyncRoute chỉ là **cấu hình**; Orchestration sở hữu Execution lifecycle.
* Không phải Aggregate Runtime mới; là **Integration configuration**.  
  → Cần bổ sung vào schema (`sync_routes` + related) và checklist Integration — **architecture note**: không thêm Workflow; có thể ghi nhận SyncRoute như config root dưới Integration (không CQRS/BPMN).

### Alternative (không khuyến nghị cho MVP multi-dest)

Nhét route vào `connections.config` của source — chỉ ổn với 1 destination cố định.

---

# 6. Idempotency — Rule + Additional Proposals

Platform key phải **stable**, **tenant-safe**, và **fan-out-aware**.

## 6.1 Storage key shape (recommended)

```text
idempotency_key = hash(
  workspace_id
  + sync_route_id
  + resource_type
  + operation          // create | update | …
  + business_key       // from rule below
  + activity_id          // optional: per destination when needed
)
```

Unique trong `(organization_id, workspace_id)` hoặc global unique string đã namespace sẵn.

## 6.2 Business key rules (configurable per SyncRoute)

| Rule name | Formula (example) | Khi dùng |
| --------- | ----------------- | -------- |
| `provider_triple` | `account_id + app_id + record_id` | Nhanh-style IDs (đề xuất của bạn) — **default cho Nhanh** |
| `provider_event` | `provider_event_id` / `webhook_delivery_id` | Webhook có id giao nhận duy nhất |
| `canonical_natural` | `resource_type + canonical_natural_key` (e.g. invoice_number + issue_date) | Khi record_id đổi nhưng business number ổn định |
| `payload_hash` | `sha256(normalized_canonical_json)` | Fallback khi thiếu natural key (cẩn thận với field noise) |
| `source_cursor_item` | `resource_type + provider_updated_at + record_id` | Poll overlapping windows |

## 6.3 Fan-out guidance

* **Intent-level key**: chống duplicate accept cùng sự kiện nguồn.
* **Execution/activity-level key**: Intent key + `destination_connection_id` + `capability` — để retry MISA không bị chặn bởi Email đã succeeded (và ngược lại).

## 6.4 MVP default

1. Intent: `provider_triple` (hoặc `provider_event` nếu webhook có delivery id).
2. Mỗi activity Execution: Intent key + activity identity.
3. Luôn namespace bằng `workspace_id` + `sync_route_id`.

---

# 7. Concepts explained

## 7.1 Webhook security (locked)

Mục tiêu: chỉ chấp nhận webhook khi caller biết **secret do Hublio phát hành**.

### Flow

1. Khi cấu hình SyncRoute / Connection nguồn Nhanh, Hublio **generate** một webhook secret (cryptographically random).
2. Secret được **lưu** (encrypted at rest) và hiện **một lần** (hoặc re-show có audit) để điền vào cấu hình webhook phía Nhanh.
3. Mỗi request từ Nhanh phải gửi secret trong **HTTP header** (tên header thống nhất, ví dụ `X-Hublio-Webhook-Secret` — chốt theo adapter Nhanh).
4. Hublio so khớp header với secret đã lưu (constant-time compare). Sai / thiếu → `401/403`, không tạo Intent.
5. Không log plaintext secret. Rotate secret = generate mới + disable secret cũ sau grace (MVP có thể rotate cứng).

Đây là shared-secret header validation (đơn giản, phù hợp Nhanh gửi kèm key). Có thể nâng HMAC body sau nếu provider hỗ trợ — không bắt buộc MVP.

## 7.2 Poll watermark / cursor (locked)

Hublio **bắt buộc lưu dấu vết** lần poll trước để lần sau chỉ kéo dữ liệu mới/thay đổi.

**Watermark** lưu theo `SyncRoute` + `resource_type`, ví dụ:

* `last_updated_at`
* và/hoặc `last_record_id`
* hoặc cursor provider trả về

### Policy

* Persist watermark sau khi batch được **accept** thành Intent (mặc định).
* Dùng overlap nhỏ + idempotency để an toàn khi clock skew / delay provider.
* Failure của destination Execution không rollback watermark Intent đã accept (retry/DLQ xử lý riêng).

---

# 8. Fan-out Sequential + Parallel (Freeze amendment)

**Allowed (v1.1 wording):**

* Một Intent có thể tạo **nhiều Execution** (fan-out theo SyncRoute).
* Group `parallel`: enqueue nhiều Execution cùng lúc.
* Group `sequential`: Execution sau chờ Execution trước (trong group) tới terminal success — hoặc dừng theo policy nếu fail.

**Still forbidden:**

* Workflow Engine / BPMN / Saga / dynamic planning
* Parallel **Steps inside a single Execution** (mỗi Execution vẫn tuần tự: validate → transform → invoke → …)
* Rule Engine

See `AGENTS.md` and `docs/00`, `docs/17`.

---

# 9. Retry Policy (locked)

```text
Attempt fail → wait 1s  → retry 1
Fail again   → wait 3s  → retry 2
Fail again   → wait 10s → retry 3
Fail again   → Dead Letter
```

* Áp dụng per Execution (per activity).
* Có thể override trên SyncRoute / Connection `retry_policy` JSON.
* DLQ: status `dead_letter` + monitoring + manual Replay.

---

# 10. Filter JSON (shape sketch)

```json
{
  "op": "AND",
  "conditions": [
    { "field": "status", "operator": "in", "value": ["confirmed", "paid"] },
    {
      "op": "OR",
      "conditions": [
        { "field": "total", "operator": ">", "value": 0 },
        { "field": "total", "operator": "between", "value": [1, 1000000] }
      ]
    }
  ]
}
```

Operators: `>`, `<`, `>=`, `<=`, `eq`, `neq`, `between`, `in`.  
Nesting: `AND` / `OR`.  
Evaluation trên Canonical (hoặc normalized source DTO trước map — **ưu tiên Canonical sau transform nhẹ / pre-filter trên provider fields documented**).

---

# 11. Engineering slice (updated)

**MVP in**

* Identity multi-tenant
* SyncRoute config + Connections Nhanh/MISA (+ utility later)
* Invoice-first Canonical; registry mở cho resource_types khác
* Webhook (shared secret in header, Hublio-generated) + Poll (persisted watermark)
* Filter engine (JSON)
* Fan-out sequential/parallel via multi-Execution
* Reverse update Nhanh
* Retry 1s/3s/10s × 3 + DLQ
* Auth connectors Nhanh + MISA
* Monitoring: Execution timeline, counts, DLQ list

**Still out**

* BPMN / visual workflow designer
* Generic Rule Engine / scripting
* Connector Marketplace
* Unbounded dynamic planning

---

# 12. Follow-ups

1. ~~Amend Freeze wording for parallel Executions~~ — done in AGENTS / docs Freeze.
2. Add `sync_routes` (+ webhook secret + watermark) to DBML when starting Integration design.
3. Confirm Nhanh header name for webhook secret + MISA auth docs.
4. Freeze Invoice Canonical fields for first mapping.
