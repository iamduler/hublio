// Package misa implements the MISA meInvoice Connector Runtime (destination e-invoice).
// Provider DTOs stay inside this package; Runtime methods only exchange Canonical map payloads.
package misa

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"hublio/internal/integration/domain"
)

// Code is the Connector code registered in the platform catalog.
const Code = "misa"

// CapabilityInvoiceCreate publishes an electronic invoice via meInvoice Open API.
const CapabilityInvoiceCreate = "invoice.create"

// Connector is the MISA meInvoice Runtime.
type Connector struct {
	httpClient *http.Client
}

// Option configures the Connector (tests inject a custom HTTP client / transport).
type Option func(*Connector)

func WithHTTPClient(c *http.Client) Option {
	return func(conn *Connector) {
		conn.httpClient = c
	}
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
	cred, err := parseCredentials(in.Config, in.Secret)
	if err != nil {
		return err
	}
	settings := parseSettings(in.Config)
	client := NewClient(settings.BaseURL, c.httpClient)
	token, err := client.GetToken(ctx, cred.AppID, cred.TaxCode, cred.Username, cred.Password)
	if err != nil {
		return err
	}
	// Authenticated templates call proves the token works beyond login (stronger than token-only).
	return client.ListTemplates(ctx, token, cred.TaxCode)
}

func (c *Connector) Health(ctx context.Context, in domain.HealthInput) error {
	_ = ctx
	// HealthInput has no Secret; auth proof belongs to Verify. Here we only assert
	// Connection config carries the seller tax code required for meInvoice calls.
	if stringField(in.Config, "tax_code", "taxcode", "taxCode") == "" {
		return ErrMissingConfig
	}
	return nil
}

func (c *Connector) Invoke(ctx context.Context, in domain.InvokeInput) (domain.InvokeOutput, error) {
	cap := strings.TrimSpace(in.Capability)
	// Accept both bare and connector-prefixed capability codes.
	switch {
	case cap == CapabilityInvoiceCreate || cap == Code+"."+CapabilityInvoiceCreate || strings.EqualFold(cap, "CreateInvoice"):
		return c.createInvoice(ctx, in)
	default:
		return domain.InvokeOutput{}, fmt.Errorf("%w: %s", ErrUnsupportedCap, cap)
	}
}

func (c *Connector) createInvoice(ctx context.Context, in domain.InvokeInput) (domain.InvokeOutput, error) {
	cred, err := parseCredentials(in.Config, in.Secret)
	if err != nil {
		return domain.InvokeOutput{}, err
	}
	settings := parseSettings(in.Config)
	invoice, err := toProviderInvoice(in.Payload, settings)
	if err != nil {
		return domain.InvokeOutput{}, err
	}

	client := NewClient(settings.BaseURL, c.httpClient)
	token, err := client.GetToken(ctx, cred.AppID, cred.TaxCode, cred.Username, cred.Password)
	if err != nil {
		return domain.InvokeOutput{}, err
	}

	result, err := client.CreateInvoice(ctx, token, cred.TaxCode, createInvoiceRequest{
		SignType:           settings.SignType,
		InvoiceData:        []invoiceData{invoice},
		PublishInvoiceData: nil,
	})
	if err != nil {
		return domain.InvokeOutput{}, err
	}

	return domain.InvokeOutput{
		Payload:  toCanonicalResponse(in.Payload, result),
		Metadata: responseMetadata(result),
	}, nil
}
