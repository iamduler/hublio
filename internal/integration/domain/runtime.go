package domain

import (
	"context"

	"github.com/google/uuid"
)

// VerifyInput carries canonical-ish, provider-agnostic data to a Connector Runtime's Verify call.
// Config/Secret are already-decrypted maps; the Runtime translates them to provider DTOs internally.
type VerifyInput struct {
	ConnectionID uuid.UUID
	Config       map[string]any
	Secret       map[string]any
}

type HealthInput struct {
	ConnectionID uuid.UUID
	Config       map[string]any
}

type InvokeInput struct {
	ConnectionID uuid.UUID
	Capability   string
	Config       map[string]any
	Secret       map[string]any
	Payload      map[string]any
}

type InvokeOutput struct {
	Payload  map[string]any
	Metadata map[string]any
}

// Runtime is the Connector Runtime contract. Implementations live only under
// internal/integration/connectors/<vendor>/ and must never leak provider DTOs across this boundary.
type Runtime interface {
	Code() string
	Verify(ctx context.Context, in VerifyInput) error
	Health(ctx context.Context, in HealthInput) error
	Invoke(ctx context.Context, in InvokeInput) (InvokeOutput, error)
}

// RuntimeRegistry resolves a Runtime by Connector code. Implemented in internal/integration/connectors.
type RuntimeRegistry interface {
	Resolve(code string) (Runtime, error)
}
