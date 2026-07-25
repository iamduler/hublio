package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	identitydomain "hublio/internal/identity/domain"
	integrationdomain "hublio/internal/integration/domain"
	orchestrationapp "hublio/internal/orchestration/application"
	"hublio/internal/platform/apperr"

	"github.com/google/uuid"
)

// SyncRouteGateway adapts Integration SyncRoute + watermarks + Identity Workspace into
// Orchestration's SyncRouteGateway port for webhook and poll ingress.
type SyncRouteGateway struct {
	routes     integrationdomain.SyncRouteRepository
	watermarks integrationdomain.SyncRouteWatermarkRepository
	workspaces identitydomain.WorkspaceRepository
	secrets    SecretDecryptor
}

func NewSyncRouteGateway(
	routes integrationdomain.SyncRouteRepository,
	watermarks integrationdomain.SyncRouteWatermarkRepository,
	workspaces identitydomain.WorkspaceRepository,
	secrets SecretDecryptor,
) *SyncRouteGateway {
	return &SyncRouteGateway{
		routes:     routes,
		watermarks: watermarks,
		workspaces: workspaces,
		secrets:    secrets,
	}
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

	groups := mapFanOutGroups(route.Activities())
	if len(groups) == 0 {
		return orchestrationapp.ResolvedWebhookRoute{}, apperr.Wrap(
			integrationdomain.ErrInvalidActivityGroup,
			"sync route has no activity step",
			apperr.ErrCodeConflict,
		)
	}

	ws, err := g.loadExecutableWorkspace(ctx, route.WorkspaceID())
	if err != nil {
		return orchestrationapp.ResolvedWebhookRoute{}, err
	}

	capability := ""
	if len(groups[0].Steps) > 0 {
		capability = groups[0].Steps[0].Capability
	}

	return orchestrationapp.ResolvedWebhookRoute{
		SyncRouteID:        route.ID(),
		OrganizationID:     ws.OrganizationID(),
		WorkspaceID:        route.WorkspaceID(),
		Capability:         capability,
		FanOutGroups:       groups,
		FanOutReverse:      mapFanOutReverse(route.SourceConnectionID(), route.Reverse()),
		IdempotencyRule:    route.IdempotencyRule(),
		SourceConnectionID: route.SourceConnectionID(),
	}, nil
}

func (g *SyncRouteGateway) ResolvePoll(ctx context.Context, in orchestrationapp.ResolvePollInput) (orchestrationapp.ResolvedPollRoute, error) {
	if g.routes == nil {
		return orchestrationapp.ResolvedPollRoute{}, apperr.New("sync route gateway not configured", apperr.ErrCodeInternal)
	}

	route, err := g.routes.FindByID(ctx, in.SyncRouteID)
	if err != nil {
		if err == integrationdomain.ErrNotFound {
			return orchestrationapp.ResolvedPollRoute{}, apperr.New("sync route not found", apperr.ErrCodeNotFound)
		}
		return orchestrationapp.ResolvedPollRoute{}, apperr.Wrap(err, "failed to load sync route", apperr.ErrCodeInternal)
	}
	if err := route.CanAcceptPoll(); err != nil {
		return orchestrationapp.ResolvedPollRoute{}, mapSyncRouteErr(err)
	}

	resourceType := strings.TrimSpace(strings.ToLower(in.ResourceType))
	if !route.AllowsResourceType(resourceType) {
		return orchestrationapp.ResolvedPollRoute{}, apperr.Wrap(integrationdomain.ErrResourceTypeNotAllowed, "resource_type not allowed", apperr.ErrCodeBadRequest)
	}

	groups := mapFanOutGroups(route.Activities())
	if len(groups) == 0 {
		return orchestrationapp.ResolvedPollRoute{}, apperr.Wrap(
			integrationdomain.ErrInvalidActivityGroup,
			"sync route has no activity step",
			apperr.ErrCodeConflict,
		)
	}

	interval, err := integrationdomain.ScheduleIntervalSeconds(route.Schedule())
	if err != nil {
		return orchestrationapp.ResolvedPollRoute{}, mapSyncRouteErr(err)
	}

	ws, err := g.loadExecutableWorkspace(ctx, route.WorkspaceID())
	if err != nil {
		return orchestrationapp.ResolvedPollRoute{}, err
	}

	capability := ""
	if len(groups[0].Steps) > 0 {
		capability = groups[0].Steps[0].Capability
	}

	return orchestrationapp.ResolvedPollRoute{
		SyncRouteID:        route.ID(),
		OrganizationID:     ws.OrganizationID(),
		WorkspaceID:        route.WorkspaceID(),
		Capability:         capability,
		ListCapability:     integrationdomain.ScheduleListCapability(route.Schedule(), resourceType),
		FanOutGroups:       groups,
		FanOutReverse:      mapFanOutReverse(route.SourceConnectionID(), route.Reverse()),
		IdempotencyRule:    route.IdempotencyRule(),
		Filter:             route.Filter(),
		SourceConnectionID: route.SourceConnectionID(),
		IntervalSeconds:    int(interval / time.Second),
	}, nil
}

func (g *SyncRouteGateway) LoadWatermark(ctx context.Context, syncRouteID uuid.UUID, resourceType string) (orchestrationapp.WatermarkSnapshot, error) {
	if g.watermarks == nil {
		return orchestrationapp.WatermarkSnapshot{}, apperr.New("watermark store not configured", apperr.ErrCodeInternal)
	}
	wm, err := g.watermarks.Find(ctx, syncRouteID, strings.TrimSpace(strings.ToLower(resourceType)))
	if err != nil {
		if errors.Is(err, integrationdomain.ErrNotFound) {
			return orchestrationapp.WatermarkSnapshot{Cursor: map[string]any{}, Found: false}, nil
		}
		return orchestrationapp.WatermarkSnapshot{}, apperr.Wrap(err, "failed to load watermark", apperr.ErrCodeInternal)
	}
	cursor := wm.Cursor
	if cursor == nil {
		cursor = map[string]any{}
	}
	return orchestrationapp.WatermarkSnapshot{
		Cursor:    cursor,
		UpdatedAt: wm.UpdatedAt,
		Found:     true,
	}, nil
}

func (g *SyncRouteGateway) AdvanceWatermark(ctx context.Context, syncRouteID uuid.UUID, resourceType string, cursor map[string]any) error {
	if g.watermarks == nil {
		return apperr.New("watermark store not configured", apperr.ErrCodeInternal)
	}
	if cursor == nil {
		cursor = map[string]any{}
	}
	wm, err := integrationdomain.NewSyncRouteWatermark(syncRouteID, resourceType, cursor, time.Now().UTC())
	if err != nil {
		return apperr.Wrap(err, "invalid watermark", apperr.ErrCodeBadRequest)
	}
	if err := g.watermarks.Upsert(ctx, wm); err != nil {
		return apperr.Wrap(err, "failed to advance watermark", apperr.ErrCodeInternal)
	}
	return nil
}

func (g *SyncRouteGateway) ListDuePollTargets(ctx context.Context, now time.Time) ([]orchestrationapp.DuePollTarget, error) {
	if g.routes == nil {
		return nil, apperr.New("sync route gateway not configured", apperr.ErrCodeInternal)
	}
	routes, err := g.routes.ListEnabledSchedulable(ctx)
	if err != nil {
		return nil, apperr.Wrap(err, "failed to list schedulable sync routes", apperr.ErrCodeInternal)
	}
	var out []orchestrationapp.DuePollTarget
	for _, route := range routes {
		if err := route.CanAcceptPoll(); err != nil {
			continue
		}
		for _, rt := range route.ResourceTypes() {
			last := time.Time{}
			if g.watermarks != nil {
				wm, err := g.watermarks.Find(ctx, route.ID(), rt)
				if err == nil {
					last = wm.UpdatedAt
				} else if !errors.Is(err, integrationdomain.ErrNotFound) {
					return nil, apperr.Wrap(err, "failed to load watermark", apperr.ErrCodeInternal)
				}
			}
			if !route.PollIsDue(last, now) {
				continue
			}
			out = append(out, orchestrationapp.DuePollTarget{
				SyncRouteID:  route.ID(),
				WorkspaceID:  route.WorkspaceID(),
				ResourceType: rt,
			})
		}
	}
	return out, nil
}

func (g *SyncRouteGateway) loadExecutableWorkspace(ctx context.Context, workspaceID uuid.UUID) (*identitydomain.Workspace, error) {
	ws, err := g.workspaces.FindByID(ctx, workspaceID)
	if err != nil {
		return nil, apperr.New("workspace not found", apperr.ErrCodeNotFound)
	}
	if !ws.CanExecuteIntents() {
		return nil, apperr.New("workspace is disabled", apperr.ErrCodeConflict)
	}
	return ws, nil
}

func (g *SyncRouteGateway) MatchRouteFilter(filter map[string]any, payload map[string]any) (bool, error) {
	return integrationdomain.MatchFilter(filter, payload)
}

func mapFanOutGroups(in []integrationdomain.ActivityGroup) []orchestrationapp.FanOutGroup {
	out := make([]orchestrationapp.FanOutGroup, 0, len(in))
	for _, g := range in {
		steps := make([]orchestrationapp.FanOutStep, 0, len(g.Steps))
		for _, s := range g.Steps {
			steps = append(steps, orchestrationapp.FanOutStep{
				ConnectionID: s.DestinationConnectionID,
				Capability:   s.Capability,
				MappingKey:   s.MappingKey,
			})
		}
		out = append(out, orchestrationapp.FanOutGroup{
			Mode:  string(g.Mode),
			Steps: steps,
		})
	}
	return out
}

func mapFanOutReverse(sourceConnID uuid.UUID, rev *integrationdomain.ReverseConfig) *orchestrationapp.FanOutReverse {
	if rev == nil {
		return nil
	}
	return &orchestrationapp.FanOutReverse{
		ConnectionID: sourceConnID,
		Capability:   rev.Capability,
		On:           rev.On,
	}
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
	case integrationdomain.ErrPollNotConfigured:
		return apperr.Wrap(err, "sync route does not accept poll", apperr.ErrCodeBadRequest)
	case integrationdomain.ErrInvalidSchedule:
		return apperr.Wrap(err, "invalid sync route schedule", apperr.ErrCodeConflict)
	case integrationdomain.ErrWebhookSecretRequired:
		return apperr.Wrap(err, "webhook secret required", apperr.ErrCodeConflict)
	case integrationdomain.ErrSyncRouteRemoved:
		return apperr.New("sync route not found", apperr.ErrCodeNotFound)
	default:
		return apperr.Wrap(err, err.Error(), apperr.ErrCodeBadRequest)
	}
}
