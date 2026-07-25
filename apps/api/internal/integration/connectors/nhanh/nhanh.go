// Package nhanh implements the Nhanh.vn Connector Runtime (origin / reverse-update).
// Provider DTOs stay inside this package.
package nhanh

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"hublio/internal/integration/domain"
)

// Code is the Connector code registered in the platform catalog.
const Code = "nhanh"

const (
	CapabilityInvoiceGet          = "invoice.get"
	CapabilityInvoiceList         = "invoice.list"
	CapabilityInvoiceUpdateStatus = "invoice.update_status"
)

// Connector is the Nhanh.vn Runtime.
type Connector struct {
	httpClient *http.Client
}

type Option func(*Connector)

func WithHTTPClient(c *http.Client) Option {
	return func(conn *Connector) { conn.httpClient = c }
}

func New(opts ...Option) *Connector {
	c := &Connector{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Connector) Code() string { return Code }

func (c *Connector) Verify(ctx context.Context, in domain.VerifyInput) error {
	a, baseURL, err := parseAuth(in.Config, in.Secret)
	if err != nil {
		return err
	}
	return NewClient(baseURL, c.httpClient).Ping(ctx, a)
}

func (c *Connector) Health(ctx context.Context, in domain.HealthInput) error {
	_ = ctx
	if stringField(in.Config, "app_id", "appId") == "" || stringField(in.Config, "business_id", "businessId") == "" {
		return ErrMissingCredentials
	}
	return nil
}

func (c *Connector) Invoke(ctx context.Context, in domain.InvokeInput) (domain.InvokeOutput, error) {
	cap := strings.TrimSpace(in.Capability)
	switch {
	case cap == CapabilityInvoiceGet || cap == Code+"."+CapabilityInvoiceGet:
		return c.getInvoice(ctx, in)
	case cap == CapabilityInvoiceList || cap == Code+"."+CapabilityInvoiceList:
		return c.listInvoices(ctx, in)
	case cap == CapabilityInvoiceUpdateStatus || cap == Code+"."+CapabilityInvoiceUpdateStatus:
		return c.updateStatus(ctx, in)
	default:
		return domain.InvokeOutput{}, fmt.Errorf("%w: %s", ErrUnsupportedCap, cap)
	}
}

func (c *Connector) getInvoice(ctx context.Context, in domain.InvokeInput) (domain.InvokeOutput, error) {
	a, baseURL, err := parseAuth(in.Config, in.Secret)
	if err != nil {
		return domain.InvokeOutput{}, err
	}
	billID := int64Field(in.Payload, "id", "bill_id", "invoice_id", "invoice_number")
	if billID == 0 {
		return domain.InvokeOutput{}, fmt.Errorf("%w: id is required", ErrInvalidPayload)
	}
	bill, err := NewClient(baseURL, c.httpClient).GetRetailBill(ctx, a, billID)
	if err != nil {
		return domain.InvokeOutput{}, err
	}
	return domain.InvokeOutput{
		Payload:  billToCanonical(bill),
		Metadata: map[string]any{"connector": Code, "capability": CapabilityInvoiceGet},
	}, nil
}

func (c *Connector) listInvoices(ctx context.Context, in domain.InvokeInput) (domain.InvokeOutput, error) {
	a, baseURL, err := parseAuth(in.Config, in.Secret)
	if err != nil {
		return domain.InvokeOutput{}, err
	}
	cursor, _ := in.Payload["cursor"].(map[string]any)
	if cursor == nil {
		cursor = map[string]any{}
	}
	pageSize := int(int64Field(in.Payload, "page_size", "limit"))
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}
	fromDate := stringField(cursor, "last_updated_at", "from_date", "updated_from")
	bills, err := NewClient(baseURL, c.httpClient).ListRetailBills(ctx, a, listBillsFilter{
		FromDate: fromDate,
		PageSize: pageSize,
	})
	if err != nil {
		return domain.InvokeOutput{}, err
	}
	records := make([]any, 0, len(bills))
	next := map[string]any{}
	for k, v := range cursor {
		next[k] = v
	}
	for i := range bills {
		rec := billToCanonical(&bills[i])
		records = append(records, rec)
		if bills[i].Date != "" {
			next["last_updated_at"] = bills[i].Date
		}
		next["last_record_id"] = fmt.Sprint(bills[i].ID)
	}
	if len(bills) == 0 {
		next["exhausted"] = true
	}
	return domain.InvokeOutput{
		Payload: map[string]any{
			"records":     records,
			"next_cursor": next,
		},
		Metadata: map[string]any{"connector": Code, "capability": CapabilityInvoiceList},
	}, nil
}

func (c *Connector) updateStatus(ctx context.Context, in domain.InvokeInput) (domain.InvokeOutput, error) {
	a, baseURL, err := parseAuth(in.Config, in.Secret)
	if err != nil {
		return domain.InvokeOutput{}, err
	}
	orderID := int64Field(in.Payload, "order_id", "id")
	status := int(int64Field(in.Payload, "status", "status_code"))
	if orderID == 0 || status == 0 {
		return domain.InvokeOutput{}, fmt.Errorf("%w: order_id and status are required", ErrInvalidPayload)
	}
	if err := NewClient(baseURL, c.httpClient).UpdateOrderStatus(ctx, a, orderID, status); err != nil {
		return domain.InvokeOutput{}, err
	}
	out := map[string]any{
		"order_id": orderID,
		"status":   status,
	}
	return domain.InvokeOutput{
		Payload:  out,
		Metadata: map[string]any{"connector": Code, "capability": CapabilityInvoiceUpdateStatus},
	}, nil
}
