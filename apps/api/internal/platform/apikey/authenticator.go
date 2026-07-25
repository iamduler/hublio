package apikey

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Principal is the authenticated machine identity for a Workspace-scoped API key.
type Principal struct {
	APIKeyID       uuid.UUID
	WorkspaceID    uuid.UUID
	OrganizationID uuid.UUID
	Name           string
}

// Authenticator verifies a plaintext API key.
// Phase A ships a fail-closed stub; Phase B provides a DB-backed implementation.
type Authenticator interface {
	Authenticate(ctx context.Context, plaintextKey string) (Principal, error)
}

// StubAuthenticator rejects all keys. Use for wiring until Identity API keys exist.
type StubAuthenticator struct{}

func NewStubAuthenticator() *StubAuthenticator {
	return &StubAuthenticator{}
}

func (s *StubAuthenticator) Authenticate(ctx context.Context, plaintextKey string) (Principal, error) {
	_ = ctx
	_ = plaintextKey
	return Principal{}, ErrUnauthorized
}

// StaticAuthenticator accepts a single bootstrap key from configuration (dev/ops only).
type StaticAuthenticator struct {
	plain string
	hash  string
}

func NewStaticAuthenticator(plaintext string) *StaticAuthenticator {
	return &StaticAuthenticator{
		plain: plaintext,
		hash:  Hash(plaintext),
	}
}

func (s *StaticAuthenticator) Authenticate(ctx context.Context, plaintextKey string) (Principal, error) {
	_ = ctx
	if plaintextKey == "" || Hash(plaintextKey) != s.hash {
		return Principal{}, ErrUnauthorized
	}
	return Principal{
		APIKeyID:       uuid.Nil,
		WorkspaceID:    uuid.Nil,
		OrganizationID: uuid.Nil,
		Name:           "bootstrap",
	}, nil
}

// TouchLastUsed may be implemented by DB authenticators later.
type TouchLastUsed interface {
	Touch(ctx context.Context, apiKeyID uuid.UUID, at time.Time) error
}
