package application

import (
	"context"

	"hublio/internal/platform/apperr"
	"hublio/internal/transformation/domain"
)

// TransformInput carries a Canonical Document plus the Operation specs to apply. Direction is
// informational only ("request" | "response") for logging/telemetry — the Use Case never
// branches business logic on it; the caller decides which Spec to pass.
type TransformInput struct {
	Direction string
	Document  map[string]any
	Spec      []domain.OperationSpec
}

// TransformResult carries back the transformed Canonical Document.
type TransformResult struct {
	Document map[string]any
}

// Services is the Transformation BC's Application layer. It has no repositories: the
// Transformation Engine never persists business data (docs/06 §2) — it only runs an
// in-memory Canonical → Canonical Pipeline built from a spec.
type Services struct{}

func NewServices() *Services {
	return &Services{}
}

// Transform runs in.Spec against in.Document and returns the normalized Canonical Document.
// A nil/empty Spec is an identity transform (the Document is returned unchanged), which is
// what lets callers pass "no rules" for capabilities that need no Canonical normalization.
func (s *Services) Transform(ctx context.Context, in TransformInput) (*TransformResult, error) {
	_ = ctx
	pipeline, err := domain.BuildPipeline(in.Spec)
	if err != nil {
		return nil, apperr.Wrap(err, "invalid transform spec", apperr.ErrCodeBadRequest)
	}
	result, err := pipeline.Run(domain.Document(in.Document))
	if err != nil {
		return nil, mapDomainErr(err)
	}
	return &TransformResult{Document: map[string]any(result)}, nil
}
