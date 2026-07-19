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
