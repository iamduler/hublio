#!/usr/bin/env bash
# MISA meInvoice Web API smoke — source of truth:
#   https://doc.meinvoice.vn/webapi/intro.html
#
# Env:
#   MISA_TAX_CODE   (required) — MST, sent as taxcode header
#   MISA_USERNAME   (required)
#   MISA_PASSWORD   (required)
#   MISA_BASE_URL   (optional) — default https://testapp.meinvoice.vn/api
#   MISA_INV_SERIES (optional) — default AB/19E (must exist for your company)
#   MISA_INV_TEMPLATE (optional) — default 01GTKT0/001
#   MISA_INV_TYPE   (optional) — default 01GTKT
#
# Usage:
#   ./scripts/misa_webapi_smoke.sh           # OAuth only
#   ./scripts/misa_webapi_smoke.sh insert    # OAuth + SAInvoice/Insert draft
set -euo pipefail

BASE_URL="${MISA_BASE_URL:-https://testapp.meinvoice.vn/api}"
BASE_URL="${BASE_URL%/}"
TAX_CODE="${MISA_TAX_CODE:-}"
USERNAME="${MISA_USERNAME:-}"
PASSWORD="${MISA_PASSWORD:-}"
MODE="${1:-oauth}"

if [[ -z "$TAX_CODE" || -z "$USERNAME" || -z "$PASSWORD" ]]; then
  echo "error: set MISA_TAX_CODE, MISA_USERNAME, MISA_PASSWORD" >&2
  exit 1
fi

oauth() {
  echo "==> POST ${BASE_URL}/v2/oauth (Web API SoT)"
  RESP=$(curl -sS -X POST "${BASE_URL}/v2/oauth" \
    -H "taxcode: ${TAX_CODE}" \
    -H 'Content-Type: text/plain' \
    --data-raw "grant_type=password&username=${USERNAME}&password=${PASSWORD}")
  echo "$RESP" | python3 -m json.tool 2>/dev/null || echo "$RESP"
  TOKEN=$(echo "$RESP" | python3 -c 'import json,sys; print(json.load(sys.stdin).get("access_token") or "")' 2>/dev/null || true)
  if [[ -z "$TOKEN" ]]; then
    echo "error: no access_token in OAuth response" >&2
    exit 1
  fi
  export MISA_ACCESS_TOKEN="$TOKEN"
  echo "ok: access_token acquired"
}

insert_draft() {
  if [[ -z "${MISA_ACCESS_TOKEN:-}" ]]; then
    echo "error: MISA_ACCESS_TOKEN empty; run oauth first" >&2
    exit 1
  fi
  SERIES="${MISA_INV_SERIES:-AB/19E}"
  TEMPLATE="${MISA_INV_TEMPLATE:-01GTKT0/001}"
  INV_TYPE="${MISA_INV_TYPE:-01GTKT}"
  BODY=$(python3 - "$SERIES" "$TEMPLATE" "$INV_TYPE" <<'PY'
import json, sys, uuid
from datetime import datetime, timezone, timedelta

series, template, inv_type = sys.argv[1], sys.argv[2], sys.argv[3]
ref_id = str(uuid.uuid4())
detail_id = str(uuid.uuid4())
inv_date = datetime.now(timezone(timedelta(hours=7))).replace(microsecond=0).isoformat()

data = {
    "RefID": ref_id,
    "RefType": 0,
    "AccountObjectName": "Hublio Smoke Buyer",
    "AccountObjectAddress": "Ha Noi, Viet Nam",
    "AccountObjectTaxCode": "",
    "PaymentMethod": "TM/CK",
    "InvTypeCode": inv_type,
    "InvTemplateNo": template,
    "InvSeries": series,
    "InvNo": "<Chưa cấp số>",
    "InvDate": inv_date,
    "CurrencyCode": "VND",
    "ExchangeRate": 1,
    "TotalSaleAmountOC": 100000,
    "TotalSaleAmount": 100000,
    "TotalDiscountAmountOC": 0,
    "TotalDiscountAmount": 0,
    "TotalVATAmountOC": 10000,
    "TotalVATAmount": 10000,
    "TotalAmountOC": 110000,
    "TotalAmount": 110000,
    "VATRate": 10,
    "PublishStatus": 0,
    "IsInvoiceDeleted": False,
    "EInvoiceStatus": 1,
    "EntityState": 1,
}
detail = [{
    "RefDetailID": detail_id,
    "RefID": ref_id,
    "InventoryItemID": "SMOKE-1",
    "InventoryItemCode": "SMOKE-1",
    "InventoryItemName": "Hublio smoke line",
    "Description": "Hublio smoke line",
    "UnitName": "LAN",
    "Quantity": 1,
    "UnitPrice": 100000,
    "AmountOC": 100000,
    "Amount": 100000,
    "DiscountRate": 0,
    "DiscountAmountOC": 0,
    "DiscountAmount": 0,
    "VATRate": 10,
    "VATAmountOC": 10000,
    "VATAmount": 10000,
    "SortOrder": 1,
    "IsPromotion": False,
    "EntityState": 1,
}]
body = {
    "data": json.dumps(data, ensure_ascii=False),
    "detail": json.dumps(detail, ensure_ascii=False),
    "EntityState": 0,
}
print(json.dumps({"ref_id": ref_id, "body": body}, ensure_ascii=False))
PY
)
  REF_ID=$(echo "$BODY" | python3 -c 'import json,sys; print(json.load(sys.stdin)["ref_id"])')
  PAYLOAD=$(echo "$BODY" | python3 -c 'import json,sys; print(json.dumps(json.load(sys.stdin)["body"], ensure_ascii=False))')

  echo "==> POST ${BASE_URL}/SAInvoice/Insert (RefID=${REF_ID})"
  RESP=$(curl -sS -X POST "${BASE_URL}/SAInvoice/Insert" \
    -H "Authorization: Bearer ${MISA_ACCESS_TOKEN}" \
    -H 'Content-Type: application/json' \
    --data-raw "$PAYLOAD")
  echo "$RESP" | python3 -m json.tool 2>/dev/null || echo "$RESP"
  echo "ok: insert attempted; verify draft on https://testapp.meinvoice.vn (intro step 3)"
  echo "    RefID=${REF_ID}"
}

oauth
if [[ "$MODE" == "insert" ]]; then
  insert_draft
fi
