// Package fake provides a Noop Connector Runtime used for local development and
// for exercising the Orchestration BC before any real provider Connector exists.
// It always verifies successfully and echoes Invoke payloads back to the caller.
package fake

import (
	"context"

	"hublio/internal/integration/domain"
)

// Code is the Connector code this Runtime answers to.
const Code = "fake"

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
	return domain.InvokeOutput{
		Payload: in.Payload,
		Metadata: map[string]any{
			"connector":  Code,
			"capability": in.Capability,
			"echo":       true,
		},
	}, nil
}
