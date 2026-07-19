package misa

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hublio/internal/integration/domain"

	"github.com/google/uuid"
)

func TestConnector_VerifyAndCreateInvoice_SandboxHTTP(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/auth/token", func(w http.ResponseWriter, r *http.Request) {
		var body tokenRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode token body: %v", err)
		}
		if body.Password == "" || body.AppID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(apiEnvelope{Success: false, ErrorCode: "UnAuthorize"})
			return
		}
		_ = json.NewEncoder(w).Encode(apiEnvelope{Success: true, Data: "sandbox-token"})
	})
	mux.HandleFunc("/invoice", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer sandbox-token" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.Header.Get("CompanyTaxCode"); got != "0101243150" {
			t.Errorf("CompanyTaxCode = %q", got)
		}
		var req createInvoiceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode invoice: %v", err)
		}
		if len(req.InvoiceData) != 1 {
			t.Fatalf("expected 1 invoice, got %d", len(req.InvoiceData))
		}
		inv := req.InvoiceData[0]
		if inv.BuyerLegalName == "" || len(inv.OriginalInvoiceDetail) == 0 {
			t.Fatalf("incomplete invoice payload: %+v", inv)
		}
		_ = json.NewEncoder(w).Encode(publishResponse{
			Success: true,
			PublishInvoiceResult: []publishInvoiceResult{{
				TransactionID: "txn-1",
				InvNo:         "0000001",
				InvSeries:     inv.InvSeries,
				InvDate:       inv.InvDate,
				RefID:         inv.RefID,
			}},
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	conn := New(WithHTTPClient(srv.Client()))
	config := map[string]any{
		"base_url":   srv.URL,
		"tax_code":   "0101243150",
		"inv_series": "1C25TYY",
	}
	secret := map[string]any{
		"app_id":   "app-1",
		"username": "user@example.com",
		"password": "secret",
	}

	if err := conn.Verify(context.Background(), domain.VerifyInput{
		ConnectionID: uuid.Must(uuid.NewV7()),
		Config:       config,
		Secret:       secret,
	}); err != nil {
		t.Fatalf("Verify: %v", err)
	}

	out, err := conn.Invoke(context.Background(), domain.InvokeInput{
		ConnectionID: uuid.Must(uuid.NewV7()),
		Capability:   CapabilityInvoiceCreate,
		Config:       config,
		Secret:       secret,
		Payload: map[string]any{
			"invoice_number": "INV-100",
			"issue_date":     "2026-07-15T10:00:00Z",
			"currency":       "VND",
			"customer": map[string]any{
				"name":     "Cong ty ABC",
				"tax_code": "0300000000",
				"address":  "Ha Noi",
			},
			"items": []any{
				map[string]any{
					"code":       "SKU-1",
					"name":       "San pham 1",
					"quantity":   2,
					"unit_price": 100000,
					"vat_rate":   "10%",
				},
			},
			"subtotal": 200000,
			"total":    220000,
		},
	})
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if out.Payload["status"] != "published" {
		t.Fatalf("status = %v, want published", out.Payload["status"])
	}
	if out.Payload["invoice_number"] != "0000001" {
		t.Fatalf("invoice_number = %v", out.Payload["invoice_number"])
	}
	if out.Metadata["transaction_id"] != "txn-1" {
		t.Fatalf("metadata = %v", out.Metadata)
	}
	// Secrets must never appear in metadata / payload.
	raw, _ := json.Marshal(out)
	if strings.Contains(string(raw), "secret") || strings.Contains(string(raw), "password") {
		t.Fatalf("response leaked secret material: %s", raw)
	}
}

func TestConnector_Verify_AuthFailure(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/token", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(apiEnvelope{Success: false, ErrorCode: "UnAuthorize"})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	conn := New(WithHTTPClient(srv.Client()))
	err := conn.Verify(context.Background(), domain.VerifyInput{
		Config: map[string]any{"base_url": srv.URL, "tax_code": "0101243150"},
		Secret: map[string]any{"app_id": "x", "username": "u", "password": "p"},
	})
	if err == nil {
		t.Fatal("expected auth error")
	}
}

func TestToProviderInvoice_RequiresSeries(t *testing.T) {
	t.Parallel()
	_, err := toProviderInvoice(map[string]any{
		"invoice_number": "INV-1",
		"issue_date":     "2026-07-15",
		"customer":       map[string]any{"name": "A"},
		"items":          []any{map[string]any{"name": "Item", "quantity": 1, "unit_price": 10}},
	}, connectionSettings{})
	if err == nil {
		t.Fatal("expected missing inv_series")
	}
}

func TestConnector_UnsupportedCapability(t *testing.T) {
	t.Parallel()
	_, err := New().Invoke(context.Background(), domain.InvokeInput{Capability: "echo"})
	if err == nil {
		t.Fatal("expected unsupported capability")
	}
}
