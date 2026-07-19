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
// Only Canonical Documents cross this boundary; capability is used solely to pick which
// built-in normalization spec applies. Capabilities that do not look invoice-like get an
// empty spec (identity transform), so e.g. the Fake connector's "fake.echo" capability is
// always passed through unchanged.
type TransformerAdapter struct {
	transform *transformationapp.Services
}

func NewTransformerAdapter(transform *transformationapp.Services) *TransformerAdapter {
	return &TransformerAdapter{transform: transform}
}

func (a *TransformerAdapter) TransformRequest(ctx context.Context, capability string, doc map[string]any) (map[string]any, error) {
	return a.run(ctx, "request", capability, doc, transformationdomain.DefaultRequestPipelineSpec())
}

func (a *TransformerAdapter) TransformResponse(ctx context.Context, capability string, doc map[string]any) (map[string]any, error) {
	return a.run(ctx, "response", capability, doc, transformationdomain.DefaultResponsePipelineSpec())
}

func (a *TransformerAdapter) run(ctx context.Context, direction, capability string, doc map[string]any, defaultSpec []transformationdomain.OperationSpec) (map[string]any, error) {
	result, err := a.transform.Transform(ctx, transformationapp.TransformInput{
		Direction: direction,
		Document:  doc,
		Spec:      specForCapability(capability, defaultSpec),
	})
	if err != nil {
		return nil, err
	}
	return result.Document, nil
}

// specForCapability applies the built-in invoice normalization only to invoice-like
// capabilities (e.g. "misa.invoice.create", "nhanh.CreateInvoice"). Every other capability
// (including the Fake connector's "fake.echo") gets a nil spec, which Transformation.Services
// runs as an identity transform.
func specForCapability(capability string, defaultSpec []transformationdomain.OperationSpec) []transformationdomain.OperationSpec {
	if strings.Contains(strings.ToLower(capability), "invoice") {
		return defaultSpec
	}
	return nil
}
