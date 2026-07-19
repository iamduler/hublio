package infrastructure

import (
	"context"
	"encoding/json"
	"strings"

	identitydomain "hublio/internal/identity/domain"
	integrationdomain "hublio/internal/integration/domain"
	orchestrationapp "hublio/internal/orchestration/application"
	"hublio/internal/platform/apperr"
)

// SyncRouteGateway adapts Integration SyncRoute + Identity Workspace into Orchestration's
// SyncRouteGateway port for webhook ingress. Provider payloads stay out of Orchestration.
type SyncRouteGateway struct {
	routes     integrationdomain.SyncRouteRepository
	workspaces identitydomain.WorkspaceRepository
	secrets    SecretDecryptor
}

func NewSyncRouteGateway(
	routes integrationdomain.SyncRouteRepository,
	workspaces identitydomain.WorkspaceRepository,
	secrets SecretDecryptor,
) *SyncRouteGateway {
	return &SyncRouteGateway{routes: routes, workspaces: workspaces, secrets: secrets}
}

func (g *SyncRouteGateway) ResolveWebhook(ctx context.Context, in orchestrationapp.ResolveWebhookInput) (orchestrationapp.ResolvedWebhookRoute, error) {
	if g.routes == nil || g.secrets == nil {
		return orchestrationapp.ResolvedWebhookRoute{}, apperr.New("sync route gateway not configured", apperr.ErrCodeInternal)
	}

	route, err := g.routes.FindByID(ctx, in.SyncRouteID)
	if err != nil {
		if err == integrationdomain.ErrNotFound {
			return orchestrationapp.ResolvedWebhookRoute{}, apperr.New("sync route not found", apperr.ErrCodeNotFound)
		}
		return orchestrationapp.ResolvedWebhookRoute{}, apperr.Wrap(err, "failed to load sync route", apperr.ErrCodeInternal)
	}
	if err := route.CanAcceptWebhook(); err != nil {
		return orchestrationapp.ResolvedWebhookRoute{}, mapSyncRouteErr(err)
	}

	expected, err := g.decryptWebhookSecret(route.WebhookSecretCiphertext())
	if err != nil {
		return orchestrationapp.ResolvedWebhookRoute{}, err
	}
	if !integrationdomain.MatchWebhookSecret(expected, in.SecretHeader) {
		return orchestrationapp.ResolvedWebhookRoute{}, apperr.New("unauthorized", apperr.ErrCodeUnauthorized)
	}

	resourceType := strings.TrimSpace(strings.ToLower(in.ResourceType))
	if !route.AllowsResourceType(resourceType) {
		return orchestrationapp.ResolvedWebhookRoute{}, apperr.Wrap(integrationdomain.ErrResourceTypeNotAllowed, "resource_type not allowed", apperr.ErrCodeBadRequest)
	}

	ok, err := integrationdomain.MatchFilter(route.Filter(), in.Payload)
	if err != nil {
		return orchestrationapp.ResolvedWebhookRoute{}, apperr.Wrap(err, "invalid filter", apperr.ErrCodeBadRequest)
	}
	if !ok {
		return orchestrationapp.ResolvedWebhookRoute{}, apperr.Wrap(integrationdomain.ErrFilterRejected, "payload rejected by filter", apperr.ErrCodeBadRequest)
	}

	step, err := route.PrimaryActivityStep()
	if err != nil {
		return orchestrationapp.ResolvedWebhookRoute{}, apperr.Wrap(err, "sync route has no activity step", apperr.ErrCodeConflict)
	}

	ws, err := g.workspaces.FindByID(ctx, route.WorkspaceID())
	if err != nil {
		return orchestrationapp.ResolvedWebhookRoute{}, apperr.New("workspace not found", apperr.ErrCodeNotFound)
	}
	if !ws.CanExecuteIntents() {
		return orchestrationapp.ResolvedWebhookRoute{}, apperr.New("workspace is disabled", apperr.ErrCodeConflict)
	}

	return orchestrationapp.ResolvedWebhookRoute{
		SyncRouteID:        route.ID(),
		OrganizationID:     ws.OrganizationID(),
		WorkspaceID:        route.WorkspaceID(),
		ConnectionID:       step.DestinationConnectionID,
		Capability:         step.Capability,
		IdempotencyRule:    route.IdempotencyRule(),
		SourceConnectionID: route.SourceConnectionID(),
	}, nil
}

func (g *SyncRouteGateway) decryptWebhookSecret(ciphertext []byte) (string, error) {
	if len(ciphertext) == 0 {
		return "", apperr.Wrap(integrationdomain.ErrWebhookSecretRequired, "webhook secret missing", apperr.ErrCodeConflict)
	}
	plaintext, err := g.secrets.Decrypt(ciphertext)
	if err != nil {
		return "", apperr.Wrap(err, "failed to decrypt webhook secret", apperr.ErrCodeInternal)
	}
	secret := map[string]any{}
	if len(plaintext) > 0 {
		if err := json.Unmarshal(plaintext, &secret); err != nil {
			return "", apperr.Wrap(err, "failed to decode webhook secret", apperr.ErrCodeInternal)
		}
	}
	raw, _ := secret["webhook_secret"].(string)
	if strings.TrimSpace(raw) == "" {
		return "", apperr.Wrap(integrationdomain.ErrWebhookSecretRequired, "webhook secret missing", apperr.ErrCodeConflict)
	}
	return raw, nil
}

func mapSyncRouteErr(err error) error {
	switch err {
	case integrationdomain.ErrSyncRouteNotEnabled:
		return apperr.Wrap(err, "sync route is not enabled", apperr.ErrCodeConflict)
	case integrationdomain.ErrWebhookNotConfigured:
		return apperr.Wrap(err, "sync route does not accept webhooks", apperr.ErrCodeBadRequest)
	case integrationdomain.ErrWebhookSecretRequired:
		return apperr.Wrap(err, "webhook secret required", apperr.ErrCodeConflict)
	case integrationdomain.ErrSyncRouteRemoved:
		return apperr.New("sync route not found", apperr.ErrCodeNotFound)
	default:
		return apperr.Wrap(err, err.Error(), apperr.ErrCodeBadRequest)
	}
}
