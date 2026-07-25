# 31 — MISA meInvoice smoke (source of truth)

> Product: Hublio  
> Status: Approved for Phase G sandbox smoke  
> Last Updated: 2026-07-19

---

## 1. Source of truth

**Primary (this smoke):** [MeInvoice Web API — Bắt đầu thi công](https://doc.meinvoice.vn/webapi/intro.html)

| | Test | Production |
| - | ---- | ---------- |
| Base URL | `https://testapp.meinvoice.vn/api` | `https://app.meinvoice.vn/api` |
| Auth | `POST /v2/oauth` (intro: `/OAuth`) | same path on prod host |
| Insert draft | `POST /SAInvoice/Insert` | same |
| UI verify | [testapp.meinvoice.vn](https://testapp.meinvoice.vn) | app.meinvoice.vn |

Entity + sample payload (official):

* [EInvoice entity](https://testapp.meinvoice.vn/api/v2/Description/Entity/EInvoice.html)
* [InsertInvoiceParam sample](https://testapp.meinvoice.vn/api/v2/Description/ExampleParam/InsertInvoiceParam.html)
* API explorer (when available): [apitest.html](https://testapp.meinvoice.vn/api/v2/content/apitest.html)

**Related (Hublio connector today):** Integration Open API under the same doc portal —
[GetToken](https://doc.meinvoice.vn/itg/Doc/GetToken.html) (`testapi.meinvoice.vn/api/integration`).
That surface needs MISA-issued `appid` and publishes via `POST /invoice`. It is **not** the
Web API in §1. Do not mix hosts, auth, or payloads.

---

## 2. Two surfaces (do not mix)

| | Web API (SoT for this smoke) | Integration Open API (Hublio `misa` Runtime) |
| - | ---------------------------- | --------------------------------------------- |
| Host (test) | `testapp.meinvoice.vn/api` | `testapi.meinvoice.vn/api/integration` |
| Auth | OAuth password grant | JSON `/auth/token` |
| Needs `appid` | No | Yes |
| Create invoice | `/SAInvoice/Insert` → **draft** | `/invoice` → create/publish |
| Publish | Sign/publish on website | API (`SignType`, often HSM=2/3) |

---

## 3. Web API smoke (manual / script)

Credentials: meInvoice **login** + **MST** (tax code). No `appid`.

### 3.1 Get token (OAuth)

Per [intro](https://doc.meinvoice.vn/webapi/intro.html) step 1 — password grant; MST in header.

```bash
export MISA_TAX_CODE='0101243150'          # MST môi trường test
export MISA_USERNAME='...'                 # tài khoản đăng nhập testapp
export MISA_PASSWORD='...'

curl -sS -X POST 'https://testapp.meinvoice.vn/api/v2/oauth' \
  -H "taxcode: ${MISA_TAX_CODE}" \
  -H 'Content-Type: text/plain' \
  --data-raw "grant_type=password&username=${MISA_USERNAME}&password=${MISA_PASSWORD}"
```

Expect JSON with `access_token`, `token_type` ≈ `bearer`, `expires_in`, often `CompanyID` / `TaxCode`.

Pass:

```bash
export MISA_ACCESS_TOKEN='...'   # from access_token
```

### 3.2 Insert draft invoice (`/SAInvoice/Insert`)

Per intro step 2. Body is `MeInvoiceParam`: **`data` and `detail` are stringified JSON**
(not nested objects). See [InsertInvoiceParam](https://testapp.meinvoice.vn/api/v2/Description/ExampleParam/InsertInvoiceParam.html).

Required-ish fields inside `data` (see [EInvoice](https://testapp.meinvoice.vn/api/v2/Description/Entity/EInvoice.html)):

* `RefID` — new GUID
* `InvTypeCode`, `InvTemplateNo`, `InvSeries`, `InvNo` (or `<Chưa cấp số>`)
* `InvDate` — ISO datetime
* `AccountObjectName` (buyer)
* totals / `CurrencyCode` / `ExchangeRate`
* `EntityState`

Each `detail` row: matching `RefID`, `RefDetailID` (GUID), item name/qty/price, `VATRate`, `EntityState`.

```bash
# Prefer: scripts/misa_webapi_smoke.sh (builds escaped data/detail)
./scripts/misa_webapi_smoke.sh insert
```

Or minimal curl (replace GUIDs / series / template for your company):

```bash
REF_ID=$(uuidgen | tr '[:upper:]' '[:lower:]')
DETAIL_ID=$(uuidgen | tr '[:upper:]' '[:lower:]')

# data/detail must be JSON strings — use the script for correct escaping.
curl -sS -X POST 'https://testapp.meinvoice.vn/api/SAInvoice/Insert' \
  -H "Authorization: Bearer ${MISA_ACCESS_TOKEN}" \
  -H 'Content-Type: application/json' \
  -d @/tmp/misa_insert_body.json
```

Expect service result with success (field names vary: `success` / `Success`). Failures: auth,
invalid series/template, or malformed stringified JSON.

### 3.3 UI check (intro step 3)

1. Open [https://testapp.meinvoice.vn](https://testapp.meinvoice.vn)
2. Find the draft by buyer name / RefID
3. Sign and publish on the website (Web API does **not** auto-publish)

---

## 4. Hublio Intent smoke (Integration Open API — current Runtime)

Use when closing Phase G exit criteria: Intent → worker → Execution `succeeded`.

Docs: [GetToken](https://doc.meinvoice.vn/itg/Doc/GetToken.html) (+ publish docs under `/itg/Doc/`).

```text
1. Seed connectors (API boot): code=misa
2. Create Connection:
   config:  { "tax_code":"<MST>", "inv_series":"<ký hiệu>", "base_url":"https://testapi.meinvoice.vn/api/integration" }
   secret:  { "app_id":"<MISA appid>", "username":"...", "password":"..." }
3. POST .../connections/:id/verify  → Active (token + templates)
4. POST /api/v1/intents  capability=invoice.create  Canonical Invoice payload
5. Worker → Execution succeeded; response status=published
```

Canonical payload (minimum):

```json
{
  "invoice_number": "SMOKE-001",
  "issue_date": "2026-07-19",
  "currency": "VND",
  "customer": { "name": "Công ty Test", "tax_code": "0101243150", "address": "Hà Nội" },
  "items": [
    { "name": "Dịch vụ test", "quantity": 1, "unit_price": 100000, "vat_rate": "10%" }
  ],
  "total": 110000
}
```

`invoice_number` maps to meInvoice `RefID` when `id` / `ref_id` omitted.

---

## 5. Pass / fail

| Check | Pass |
| ----- | ---- |
| Web API §3.1 | `access_token` returned |
| Web API §3.2 | Insert success; draft visible on testapp |
| Web API §3.3 | Can sign/publish in UI |
| Hublio §4 | Connection Active + Execution `succeeded` |

Record MST, InvSeries / InvTemplateNo, and environment (testapp vs testapi) in the smoke notes.
Never commit passwords or `appid` into the repo.

---

## 6. Script

```bash
# From repo root — requires curl + python3 (JSON escape)
export MISA_TAX_CODE=...
export MISA_USERNAME=...
export MISA_PASSWORD=...
./scripts/misa_webapi_smoke.sh          # oauth only
./scripts/misa_webapi_smoke.sh insert   # oauth + SAInvoice/Insert
```
