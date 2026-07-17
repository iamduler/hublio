package infrastructure

import (
	"context"
	"errors"
	"time"

	"hublio/internal/identity/domain"
	"hublio/internal/platform/apikey"

	"github.com/google/uuid"
)

// DBAuthenticator authenticates Workspace-scoped API keys from PostgreSQL.
type DBAuthenticator struct {
	keys       domain.APIKeyRepository
	workspaces domain.WorkspaceRepository
	orgs       domain.OrganizationRepository
}

func NewDBAuthenticator(
	keys domain.APIKeyRepository,
	workspaces domain.WorkspaceRepository,
	orgs domain.OrganizationRepository,
) *DBAuthenticator {
	return &DBAuthenticator{keys: keys, workspaces: workspaces, orgs: orgs}
}

func (a *DBAuthenticator) Authenticate(ctx context.Context, plaintextKey string) (apikey.Principal, error) {
	prefix, ok := apikey.PrefixFromPlaintext(plaintextKey)
	if !ok {
		return apikey.Principal{}, apikey.ErrUnauthorized
	}

	key, err := a.keys.FindByPrefix(ctx, prefix)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apikey.Principal{}, apikey.ErrUnauthorized
		}
		return apikey.Principal{}, err
	}
	if key.KeyHash() != apikey.Hash(plaintextKey) {
		return apikey.Principal{}, apikey.ErrUnauthorized
	}
	now := time.Now().UTC()
	if err := key.IsUsable(now); err != nil {
		return apikey.Principal{}, apikey.ErrUnauthorized
	}

	ws, err := a.workspaces.FindByID(ctx, key.WorkspaceID())
	if err != nil {
		return apikey.Principal{}, apikey.ErrUnauthorized
	}
	if !ws.CanExecuteIntents() {
		return apikey.Principal{}, apikey.ErrUnauthorized
	}

	org, err := a.orgs.FindByID(ctx, ws.OrganizationID())
	if err != nil || !org.CanSubmitIntents() {
		return apikey.Principal{}, apikey.ErrUnauthorized
	}

	_ = a.keys.TouchLastUsed(ctx, key.ID(), now)

	return apikey.Principal{
		APIKeyID:       key.ID(),
		WorkspaceID:    ws.ID(),
		OrganizationID: org.ID(),
		Name:           key.Name(),
	}, nil
}

func (a *DBAuthenticator) Touch(ctx context.Context, apiKeyID uuid.UUID, at time.Time) error {
	return a.keys.TouchLastUsed(ctx, apiKeyID, at)
}
