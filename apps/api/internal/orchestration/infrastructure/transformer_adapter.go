package infrastructure

import (
	"context"
	"strings"

	transformationapp "hublio/internal/transformation/application"
	transformationdomain "hublio/internal/transformation/domain"
)

// TransformerAdapter adapts the Transformation Application into the Orchestration
// Application's Transformer port. It lives in Infrastructure so that Orchestration's own
// Domain/Application never import the Transformation BC directly.
//
// Only Canonical Documents cross this boundary; capability selects which built-in spec runs:
//   - invoice create/publish → full request pipeline (incl. ValidateRequired)
//   - other invoice capabilities (get / update_status) → identity on request; light normalize on response
//   - everything else (e.g. fake.echo) → identity
type TransformerAdapter struct {
	transform *transformationapp.Services
}

func NewTransformerAdapter(transform *transformationapp.Services) *TransformerAdapter {
	return &TransformerAdapter{transform: transform}
}

func (a *TransformerAdapter) TransformRequest(ctx context.Context, capability string, doc map[string]any) (map[string]any, error) {
	return a.run(ctx, "request", doc, requestSpecForCapability(capability))
}

func (a *TransformerAdapter) TransformResponse(ctx context.Context, capability string, doc map[string]any) (map[string]any, error) {
	return a.run(ctx, "response", doc, responseSpecForCapability(capability))
}

func (a *TransformerAdapter) run(ctx context.Context, direction string, doc map[string]any, spec []transformationdomain.OperationSpec) (map[string]any, error) {
	result, err := a.transform.Transform(ctx, transformationapp.TransformInput{
		Direction: direction,
		Document:  doc,
		Spec:      spec,
	})
	if err != nil {
		return nil, err
	}
	return result.Document, nil
}

func requestSpecForCapability(capability string) []transformationdomain.OperationSpec {
	if isInvoiceCreateCapability(capability) {
		return transformationdomain.DefaultRequestPipelineSpec()
	}
	return nil
}

func responseSpecForCapability(capability string) []transformationdomain.OperationSpec {
	if isInvoiceCapability(capability) {
		return transformationdomain.DefaultResponsePipelineSpec()
	}
	return nil
}

func isInvoiceCapability(capability string) bool {
	return strings.Contains(strings.ToLower(capability), "invoice")
}

// isInvoiceCreateCapability is true for create/publish invoice Intents that carry a full
// Canonical Invoice document. Read/update capabilities (invoice.get, invoice.update_status)
// must not run ValidateRequired(invoice_number, issue_date).
func isInvoiceCreateCapability(capability string) bool {
	c := strings.ToLower(strings.TrimSpace(capability))
	if !strings.Contains(c, "invoice") {
		return false
	}
	return strings.Contains(c, "create") || strings.Contains(c, "publish")
}
