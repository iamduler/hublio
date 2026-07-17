package connectors

import (
	"fmt"

	"hublio/internal/integration/domain"
)

// Registry resolves a Connector Runtime by Connector code. It implements domain.RuntimeRegistry.
type Registry struct {
	runtimes map[string]domain.Runtime
}

func NewRegistry(runtimes ...domain.Runtime) *Registry {
	r := &Registry{runtimes: make(map[string]domain.Runtime, len(runtimes))}
	for _, rt := range runtimes {
		r.Register(rt)
	}
	return r
}

func (r *Registry) Register(rt domain.Runtime) {
	r.runtimes[rt.Code()] = rt
}

func (r *Registry) Resolve(code string) (domain.Runtime, error) {
	rt, ok := r.runtimes[code]
	if !ok {
		return nil, fmt.Errorf("connectors: no runtime registered for code %q", code)
	}
	return rt, nil
}
