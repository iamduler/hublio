package nhanh

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

func TestConnector_VerifyGetAndUpdate_SandboxHTTP(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/product/list", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "tok-1" {
			_ = json.NewEncoder(w).Encode(apiResponse{Code: 0, ErrorCode: "ERR_INVALID_ACCESS_TOKEN"})
			return
		}
		_ = json.NewEncoder(w).Encode(apiResponse{Code: 1, Data: []any{}})
	})
	mux.HandleFunc("/bill/retail", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(apiResponse{Code: 1, Data: []retailBill{{
			ID:       99,
			OrderID:  55,
			Date:     "2026-07-15",
			Money:    110000,
			Customer: &billCustomer{ID: 1, Name: "Khach A", Mobile: "090", Address: "HN"},
			Products: []billProduct{{Code: "P1", Name: "Hang 1", Quantity: 1, Price: 100000, VAT: 10}},
		}}})
	})
	mux.HandleFunc("/order/edit", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(apiResponse{Code: 1})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	conn := New(WithHTTPClient(srv.Client()))
	config := map[string]any{
		"base_url":    srv.URL,
		"app_id":      123,
		"business_id": 456,
	}
	secret := map[string]any{"access_token": "tok-1"}

	if err := conn.Verify(context.Background(), domain.VerifyInput{
		ConnectionID: uuid.Must(uuid.NewV7()),
		Config:       config,
		Secret:       secret,
	}); err != nil {
		t.Fatalf("Verify: %v", err)
	}

	got, err := conn.Invoke(context.Background(), domain.InvokeInput{
		Capability: CapabilityInvoiceGet,
		Config:     config,
		Secret:     secret,
		Payload:    map[string]any{"id": 99},
	})
	if err != nil {
		t.Fatalf("invoice.get: %v", err)
	}
	if got.Payload["invoice_number"] != "99" {
		t.Fatalf("payload = %v", got.Payload)
	}

	upd, err := conn.Invoke(context.Background(), domain.InvokeInput{
		Capability: CapabilityInvoiceUpdateStatus,
		Config:     config,
		Secret:     secret,
		Payload:    map[string]any{"order_id": 55, "status": 2},
	})
	if err != nil {
		t.Fatalf("invoice.update_status: %v", err)
	}
	raw, _ := json.Marshal(upd)
	if strings.Contains(string(raw), "tok-1") {
		t.Fatalf("leaked access token: %s", raw)
	}
}

func TestConnector_Verify_BadToken(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/product/list", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(apiResponse{Code: 0, ErrorCode: "ERR_INVALID_ACCESS_TOKEN"})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	err := New(WithHTTPClient(srv.Client())).Verify(context.Background(), domain.VerifyInput{
		Config: map[string]any{"base_url": srv.URL, "app_id": 1, "business_id": 2},
		Secret: map[string]any{"access_token": "bad"},
	})
	if err == nil {
		t.Fatal("expected auth error")
	}
}
