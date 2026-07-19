package fake

import (
	"context"
	"fmt"
	"strings"

	"hublio/internal/integration/domain"
)

// Code is the Connector code this Runtime answers to.
const Code = "fake"

const (
	CapabilityEcho        = "echo"
	CapabilityInvoiceList = "invoice.list"
)

// Connector is a Noop Connector Runtime: it never calls any external system.
type Connector struct{}

func New() *Connector { return &Connector{} }

func (c *Connector) Code() string { return Code }

func (c *Connector) Verify(ctx context.Context, in domain.VerifyInput) error {
	_ = ctx
	_ = in
	return nil
}

func (c *Connector) Health(ctx context.Context, in domain.HealthInput) error {
	_ = ctx
	_ = in
	return nil
}

func (c *Connector) Invoke(ctx context.Context, in domain.InvokeInput) (domain.InvokeOutput, error) {
	_ = ctx
	cap := strings.TrimSpace(in.Capability)
	switch {
	case cap == CapabilityInvoiceList || cap == Code+"."+CapabilityInvoiceList || strings.HasSuffix(cap, ".list"):
		return listSinceCursor(in)
	default:
		return domain.InvokeOutput{
			Payload: in.Payload,
			Metadata: map[string]any{
				"connector":  Code,
				"capability": in.Capability,
				"echo":       true,
			},
		}, nil
	}
}

// listSinceCursor returns a small synthetic page for poll e2e. Cursor keys:
// last_record_id / exhausted. After id "2" the page is empty.
func listSinceCursor(in domain.InvokeInput) (domain.InvokeOutput, error) {
	cursor, _ := in.Payload["cursor"].(map[string]any)
	if cursor == nil {
		cursor = map[string]any{}
	}
	if exhausted, _ := cursor["exhausted"].(bool); exhausted {
		return domain.InvokeOutput{
			Payload: map[string]any{
				"records":     []any{},
				"next_cursor": cursor,
			},
			Metadata: map[string]any{"connector": Code, "capability": CapabilityInvoiceList},
		}, nil
	}

	lastID := fmt.Sprint(cursor["last_record_id"])
	var records []any
	switch lastID {
	case "", "<nil>":
		records = []any{
			map[string]any{
				"id":             "1",
				"invoice_number": "INV-FAKE-1",
				"total":          100,
				"currency":       "VND",
				"status":         "confirmed",
				"updated_at":     "2026-01-01T00:00:00Z",
			},
			map[string]any{
				"id":             "2",
				"invoice_number": "INV-FAKE-2",
				"total":          200,
				"currency":       "VND",
				"status":         "confirmed",
				"updated_at":     "2026-01-01T01:00:00Z",
			},
		}
	default:
		records = []any{}
	}

	next := map[string]any{
		"last_record_id": "2",
		"last_updated_at": "2026-01-01T01:00:00Z",
	}
	if len(records) == 0 {
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
