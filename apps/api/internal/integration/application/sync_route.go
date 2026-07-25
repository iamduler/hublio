package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"hublio/internal/integration/domain"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/id"

	"github.com/google/uuid"
)

type ActivityStepInput struct {
	DestinationConnectionID uuid.UUID
	Capability              string
	MappingKey              string
}

type ActivityGroupInput struct {
	Mode  string
	Steps []ActivityStepInput
}

type ReverseInput struct {
	Capability string
	On         string
}

type CreateSyncRouteInput struct {
	WorkspaceID        uuid.UUID
	SourceConnectionID uuid.UUID
	Name               string
	Trigger            string
	ResourceTypes      []string
	Schedule           map[string]any
	Filter             map[string]any
	IdempotencyRule    map[string]any
	Activities         []ActivityGroupInput
	Reverse            *ReverseInput
	RetryPolicy        map[string]any
}

type CreateSyncRouteResult struct {
	Route                  *domain.SyncRoute
	WebhookSecretPlaintext string // only on create/rotate when trigger needs webhook; never persisted plaintext
}

type UpdateSyncRouteInput struct {
	WorkspaceID        uuid.UUID
	SyncRouteID        uuid.UUID
	Name               *string
	SourceConnectionID *uuid.UUID
	Trigger            *string
	ResourceTypes      []string
	Schedule           map[string]any
	Filter             map[string]any
	IdempotencyRule    map[string]any
	Activities         []ActivityGroupInput
	Reverse            *ReverseInput
	ClearReverse       bool
	RetryPolicy        map[string]any
}

// CreateSyncRoute creates a Draft SyncRoute. When trigger is webhook|both, generates and
// encrypts a webhook secret; plaintext is returned once and never stored.
func (s *Services) CreateSyncRoute(ctx context.Context, in CreateSyncRouteInput) (*CreateSyncRouteResult, error) {
	if _, err := s.findWorkspaceConnection(ctx, in.WorkspaceID, in.SourceConnectionID); err != nil {
		return nil, err
	}
	if err := s.validateActivityConnections(ctx, in.WorkspaceID, in.Activities); err != nil {
		return nil, err
	}

	trigger, err := domain.ParseSyncRouteTrigger(in.Trigger)
	if err != nil {
		return nil, mapDomainErr(err)
	}
	activities, err := mapActivityInputs(in.Activities)
	if err != nil {
		return nil, mapDomainErr(err)
	}
	reverse, err := mapReverseInput(in.Reverse)
	if err != nil {
		return nil, mapDomainErr(err)
	}

	routeID, err := id.NewV7()
	if err != nil {
		return nil, apperr.Wrap(err, "failed to generate sync route id", apperr.ErrCodeInternal)
	}
	now := s.clock().Now()
	route, err := domain.NewSyncRoute(domain.NewSyncRouteParams{
		ID:                 routeID,
		WorkspaceID:        in.WorkspaceID,
		SourceConnectionID: in.SourceConnectionID,
		Name:               in.Name,
		Trigger:            trigger,
		ResourceTypes:      in.ResourceTypes,
		Schedule:           in.Schedule,
		Filter:             in.Filter,
		IdempotencyRule:    in.IdempotencyRule,
		Activities:         activities,
		Reverse:            reverse,
		RetryPolicy:        in.RetryPolicy,
		Now:                now,
	})
	if err != nil {
		return nil, mapDomainErr(err)
	}

	var plaintext string
	if route.NeedsWebhookSecret() {
		plaintext, err = generateWebhookSecret()
		if err != nil {
			return nil, err
		}
		ciphertext, err := s.encryptWebhookSecret(plaintext)
		if err != nil {
			return nil, err
		}
		if err := route.AttachWebhookSecretCiphertext(ciphertext, now); err != nil {
			return nil, mapDomainErr(err)
		}
	}

	if err := s.SyncRoutes.Save(ctx, route); err != nil {
		return nil, mapRepoErr(err)
	}

	return &CreateSyncRouteResult{Route: route, WebhookSecretPlaintext: plaintext}, nil
}

func (s *Services) UpdateSyncRoute(ctx context.Context, in UpdateSyncRouteInput) (*domain.SyncRoute, error) {
	route, err := s.findWorkspaceSyncRoute(ctx, in.WorkspaceID, in.SyncRouteID)
	if err != nil {
		return nil, err
	}
	if in.SourceConnectionID != nil {
		if _, err := s.findWorkspaceConnection(ctx, in.WorkspaceID, *in.SourceConnectionID); err != nil {
			return nil, err
		}
	}
	if in.Activities != nil {
		if err := s.validateActivityConnections(ctx, in.WorkspaceID, in.Activities); err != nil {
			return nil, err
		}
	}

	params := domain.UpdateSyncRouteParams{
		Name:               in.Name,
		SourceConnectionID: in.SourceConnectionID,
		ResourceTypes:      in.ResourceTypes,
		Schedule:           in.Schedule,
		Filter:             in.Filter,
		IdempotencyRule:    in.IdempotencyRule,
		ClearReverse:       in.ClearReverse,
		RetryPolicy:        in.RetryPolicy,
		Now:                s.clock().Now(),
	}
	if in.Trigger != nil {
		tr, err := domain.ParseSyncRouteTrigger(*in.Trigger)
		if err != nil {
			return nil, mapDomainErr(err)
		}
		params.Trigger = &tr
	}
	if in.Activities != nil {
		activities, err := mapActivityInputs(in.Activities)
		if err != nil {
			return nil, mapDomainErr(err)
		}
		params.Activities = activities
	}
	if in.Reverse != nil {
		reverse, err := mapReverseInput(in.Reverse)
		if err != nil {
			return nil, mapDomainErr(err)
		}
		params.Reverse = reverse
	}

	if err := route.Update(params); err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.SyncRoutes.Update(ctx, route); err != nil {
		return nil, mapRepoErr(err)
	}
	return route, nil
}

func (s *Services) EnableSyncRoute(ctx context.Context, workspaceID, syncRouteID uuid.UUID) (*domain.SyncRoute, error) {
	return s.changeSyncRoute(ctx, workspaceID, syncRouteID, func(r *domain.SyncRoute) error {
		return r.Enable(s.clock().Now())
	})
}

func (s *Services) DisableSyncRoute(ctx context.Context, workspaceID, syncRouteID uuid.UUID) (*domain.SyncRoute, error) {
	return s.changeSyncRoute(ctx, workspaceID, syncRouteID, func(r *domain.SyncRoute) error {
		return r.Disable(s.clock().Now())
	})
}

func (s *Services) DeleteSyncRoute(ctx context.Context, workspaceID, syncRouteID uuid.UUID) (*domain.SyncRoute, error) {
	return s.changeSyncRoute(ctx, workspaceID, syncRouteID, func(r *domain.SyncRoute) error {
		return r.SoftDelete(s.clock().Now())
	})
}

func (s *Services) GetSyncRoute(ctx context.Context, workspaceID, syncRouteID uuid.UUID) (*domain.SyncRoute, error) {
	return s.findWorkspaceSyncRoute(ctx, workspaceID, syncRouteID)
}

func (s *Services) ListSyncRoutes(ctx context.Context, workspaceID uuid.UUID) ([]*domain.SyncRoute, error) {
	list, err := s.SyncRoutes.ListByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return list, nil
}

type RotateSyncRouteWebhookSecretResult struct {
	Route                  *domain.SyncRoute
	WebhookSecretPlaintext string
}

// RotateSyncRouteWebhookSecret generates a new webhook secret, encrypts it, and returns plaintext once.
func (s *Services) RotateSyncRouteWebhookSecret(ctx context.Context, workspaceID, syncRouteID uuid.UUID) (*RotateSyncRouteWebhookSecretResult, error) {
	route, err := s.findWorkspaceSyncRoute(ctx, workspaceID, syncRouteID)
	if err != nil {
		return nil, err
	}
	if !route.NeedsWebhookSecret() {
		return nil, apperr.New("sync route trigger does not use webhook secrets", apperr.ErrCodeBadRequest)
	}
	plaintext, err := generateWebhookSecret()
	if err != nil {
		return nil, err
	}
	ciphertext, err := s.encryptWebhookSecret(plaintext)
	if err != nil {
		return nil, err
	}
	if err := route.RotateWebhookSecretCiphertext(ciphertext, s.clock().Now()); err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.SyncRoutes.Update(ctx, route); err != nil {
		return nil, mapRepoErr(err)
	}
	return &RotateSyncRouteWebhookSecretResult{Route: route, WebhookSecretPlaintext: plaintext}, nil
}

func (s *Services) UpsertSyncRouteWatermark(ctx context.Context, workspaceID, syncRouteID uuid.UUID, resourceType string, cursor map[string]any) (*domain.SyncRouteWatermark, error) {
	if _, err := s.findWorkspaceSyncRoute(ctx, workspaceID, syncRouteID); err != nil {
		return nil, err
	}
	wm, err := domain.NewSyncRouteWatermark(syncRouteID, resourceType, cursor, s.clock().Now())
	if err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.Watermarks.Upsert(ctx, wm); err != nil {
		return nil, mapRepoErr(err)
	}
	return wm, nil
}

func (s *Services) GetSyncRouteWatermark(ctx context.Context, workspaceID, syncRouteID uuid.UUID, resourceType string) (*domain.SyncRouteWatermark, error) {
	if _, err := s.findWorkspaceSyncRoute(ctx, workspaceID, syncRouteID); err != nil {
		return nil, err
	}
	wm, err := s.Watermarks.Find(ctx, syncRouteID, resourceType)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return wm, nil
}

func (s *Services) ListSyncRouteWatermarks(ctx context.Context, workspaceID, syncRouteID uuid.UUID) ([]*domain.SyncRouteWatermark, error) {
	if _, err := s.findWorkspaceSyncRoute(ctx, workspaceID, syncRouteID); err != nil {
		return nil, err
	}
	list, err := s.Watermarks.ListBySyncRoute(ctx, syncRouteID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return list, nil
}

func (s *Services) changeSyncRoute(ctx context.Context, workspaceID, syncRouteID uuid.UUID, fn func(*domain.SyncRoute) error) (*domain.SyncRoute, error) {
	route, err := s.findWorkspaceSyncRoute(ctx, workspaceID, syncRouteID)
	if err != nil {
		return nil, err
	}
	if err := fn(route); err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.SyncRoutes.Update(ctx, route); err != nil {
		return nil, mapRepoErr(err)
	}
	return route, nil
}

func (s *Services) findWorkspaceSyncRoute(ctx context.Context, workspaceID, syncRouteID uuid.UUID) (*domain.SyncRoute, error) {
	if s.SyncRoutes == nil {
		return nil, apperr.New("sync route repository not configured", apperr.ErrCodeInternal)
	}
	route, err := s.SyncRoutes.FindByID(ctx, syncRouteID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if route.WorkspaceID() != workspaceID {
		return nil, apperr.New("sync route not found", apperr.ErrCodeNotFound)
	}
	return route, nil
}

func (s *Services) validateActivityConnections(ctx context.Context, workspaceID uuid.UUID, groups []ActivityGroupInput) error {
	for _, g := range groups {
		for _, step := range g.Steps {
			if _, err := s.findWorkspaceConnection(ctx, workspaceID, step.DestinationConnectionID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Services) encryptWebhookSecret(plaintext string) ([]byte, error) {
	return s.encryptSecret(map[string]any{"webhook_secret": plaintext})
}

func generateWebhookSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", apperr.Wrap(err, "failed to generate webhook secret", apperr.ErrCodeInternal)
	}
	return hex.EncodeToString(buf), nil
}

func mapActivityInputs(in []ActivityGroupInput) ([]domain.ActivityGroup, error) {
	out := make([]domain.ActivityGroup, 0, len(in))
	for _, g := range in {
		mode, err := domain.ParseActivityGroupMode(g.Mode)
		if err != nil {
			return nil, err
		}
		steps := make([]domain.ActivityStep, 0, len(g.Steps))
		for _, s := range g.Steps {
			steps = append(steps, domain.ActivityStep{
				DestinationConnectionID: s.DestinationConnectionID,
				Capability:              s.Capability,
				MappingKey:              s.MappingKey,
			})
		}
		out = append(out, domain.ActivityGroup{Mode: mode, Steps: steps})
	}
	return out, nil
}

func mapReverseInput(in *ReverseInput) (*domain.ReverseConfig, error) {
	if in == nil {
		return nil, nil
	}
	return &domain.ReverseConfig{Capability: in.Capability, On: in.On}, nil
}
