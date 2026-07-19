package infrastructure

import (
	"context"
	"encoding/json"

	"hublio/internal/integration/domain"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/persistence/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SyncRouteRepository struct {
	pool *pgxpool.Pool
}

func NewSyncRouteRepository(pool *pgxpool.Pool) *SyncRouteRepository {
	return &SyncRouteRepository{pool: pool}
}

func (r *SyncRouteRepository) q(ctx context.Context) *sqlc.Queries {
	return sqlc.New(persistence.Conn(ctx, r.pool))
}

func (r *SyncRouteRepository) Save(ctx context.Context, route *domain.SyncRoute) error {
	params, err := syncRouteInsertParams(route)
	if err != nil {
		return err
	}
	return mapUnique(r.q(ctx).InsertSyncRoute(ctx, params))
}

func (r *SyncRouteRepository) Update(ctx context.Context, route *domain.SyncRoute) error {
	params, err := syncRouteUpdateParams(route)
	if err != nil {
		return err
	}
	return mapUnique(r.q(ctx).UpdateSyncRoute(ctx, params))
}

func (r *SyncRouteRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.SyncRoute, error) {
	row, err := r.q(ctx).GetSyncRouteByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return mapSyncRoute(row)
}

func (r *SyncRouteRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*domain.SyncRoute, error) {
	rows, err := r.q(ctx).ListSyncRoutesByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.SyncRoute, 0, len(rows))
	for _, row := range rows {
		route, err := mapSyncRoute(row)
		if err != nil {
			return nil, err
		}
		out = append(out, route)
	}
	return out, nil
}

func (r *SyncRouteRepository) ListEnabledSchedulable(ctx context.Context) ([]*domain.SyncRoute, error) {
	rows, err := r.q(ctx).ListEnabledSchedulableSyncRoutes(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.SyncRoute, 0, len(rows))
	for _, row := range rows {
		route, err := mapSyncRoute(row)
		if err != nil {
			return nil, err
		}
		out = append(out, route)
	}
	return out, nil
}

type SyncRouteWatermarkRepository struct {
	pool *pgxpool.Pool
}

func NewSyncRouteWatermarkRepository(pool *pgxpool.Pool) *SyncRouteWatermarkRepository {
	return &SyncRouteWatermarkRepository{pool: pool}
}

func (r *SyncRouteWatermarkRepository) q(ctx context.Context) *sqlc.Queries {
	return sqlc.New(persistence.Conn(ctx, r.pool))
}

func (r *SyncRouteWatermarkRepository) Upsert(ctx context.Context, wm *domain.SyncRouteWatermark) error {
	cursor, err := marshalJSONMap(wm.Cursor)
	if err != nil {
		return err
	}
	if cursor == nil {
		cursor = []byte("{}")
	}
	return r.q(ctx).UpsertSyncRouteWatermark(ctx, sqlc.UpsertSyncRouteWatermarkParams{
		SyncRouteID:  wm.SyncRouteID,
		ResourceType: wm.ResourceType,
		Cursor:       cursor,
		UpdatedAt:    timestamptz(wm.UpdatedAt),
	})
}

func (r *SyncRouteWatermarkRepository) Find(ctx context.Context, syncRouteID uuid.UUID, resourceType string) (*domain.SyncRouteWatermark, error) {
	row, err := r.q(ctx).GetSyncRouteWatermark(ctx, sqlc.GetSyncRouteWatermarkParams{
		SyncRouteID:  syncRouteID,
		ResourceType: resourceType,
	})
	if err != nil {
		return nil, mapNotFound(err)
	}
	return mapWatermark(row)
}

func (r *SyncRouteWatermarkRepository) ListBySyncRoute(ctx context.Context, syncRouteID uuid.UUID) ([]*domain.SyncRouteWatermark, error) {
	rows, err := r.q(ctx).ListSyncRouteWatermarks(ctx, syncRouteID)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.SyncRouteWatermark, 0, len(rows))
	for _, row := range rows {
		wm, err := mapWatermark(row)
		if err != nil {
			return nil, err
		}
		out = append(out, wm)
	}
	return out, nil
}

func syncRouteInsertParams(route *domain.SyncRoute) (sqlc.InsertSyncRouteParams, error) {
	resourceTypes, schedule, filter, idempotency, activities, reverse, retry, webhook, err := marshalSyncRouteJSON(route)
	if err != nil {
		return sqlc.InsertSyncRouteParams{}, err
	}
	return sqlc.InsertSyncRouteParams{
		ID:                 route.ID(),
		WorkspaceID:        route.WorkspaceID(),
		SourceConnectionID: route.SourceConnectionID(),
		Name:               route.Name(),
		Status:             string(route.Status()),
		TriggerType:        string(route.Trigger()),
		ResourceTypes:      resourceTypes,
		Schedule:           schedule,
		Filter:             filter,
		IdempotencyRule:    idempotency,
		Activities:         activities,
		Reverse:            reverse,
		RetryPolicy:        retry,
		WebhookSecret:      webhook,
		CreatedAt:          timestamptz(route.CreatedAt()),
		UpdatedAt:          timestamptz(route.UpdatedAt()),
		DeletedAt:          timestamptzPtr(route.DeletedAt()),
	}, nil
}

func syncRouteUpdateParams(route *domain.SyncRoute) (sqlc.UpdateSyncRouteParams, error) {
	resourceTypes, schedule, filter, idempotency, activities, reverse, retry, webhook, err := marshalSyncRouteJSON(route)
	if err != nil {
		return sqlc.UpdateSyncRouteParams{}, err
	}
	return sqlc.UpdateSyncRouteParams{
		ID:                 route.ID(),
		SourceConnectionID: route.SourceConnectionID(),
		Name:               route.Name(),
		Status:             string(route.Status()),
		TriggerType:        string(route.Trigger()),
		ResourceTypes:      resourceTypes,
		Schedule:           schedule,
		Filter:             filter,
		IdempotencyRule:    idempotency,
		Activities:         activities,
		Reverse:            reverse,
		RetryPolicy:        retry,
		WebhookSecret:      webhook,
		UpdatedAt:          timestamptz(route.UpdatedAt()),
		DeletedAt:          timestamptzPtr(route.DeletedAt()),
	}, nil
}

func marshalSyncRouteJSON(route *domain.SyncRoute) (resourceTypes, schedule, filter, idempotency, activities, reverse, retry, webhook []byte, err error) {
	resourceTypes, err = json.Marshal(route.ResourceTypes())
	if err != nil {
		return
	}
	schedule, err = marshalJSONMap(route.Schedule())
	if err != nil {
		return
	}
	filter, err = marshalJSONMap(route.Filter())
	if err != nil {
		return
	}
	idempotency, err = marshalJSONMap(route.IdempotencyRule())
	if err != nil {
		return
	}
	activities, err = marshalActivities(route.Activities())
	if err != nil {
		return
	}
	reverse, err = marshalReverse(route.Reverse())
	if err != nil {
		return
	}
	retry, err = marshalJSONMap(route.RetryPolicy())
	if err != nil {
		return
	}
	if len(route.WebhookSecretCiphertext()) > 0 {
		webhook, err = marshalEncryptedSecret(route.WebhookSecretCiphertext())
	}
	return
}

type activityStepDTO struct {
	DestinationConnectionID string `json:"destination_connection_id"`
	Capability              string `json:"capability"`
	MappingKey              string `json:"mapping_key,omitempty"`
}

type activityGroupDTO struct {
	Mode  string            `json:"group_mode"`
	Steps []activityStepDTO `json:"steps"`
}

type reverseDTO struct {
	Capability string `json:"capability"`
	On         string `json:"on"`
}

func marshalActivities(groups []domain.ActivityGroup) ([]byte, error) {
	dtos := make([]activityGroupDTO, 0, len(groups))
	for _, g := range groups {
		steps := make([]activityStepDTO, 0, len(g.Steps))
		for _, s := range g.Steps {
			steps = append(steps, activityStepDTO{
				DestinationConnectionID: s.DestinationConnectionID.String(),
				Capability:              s.Capability,
				MappingKey:              s.MappingKey,
			})
		}
		dtos = append(dtos, activityGroupDTO{Mode: string(g.Mode), Steps: steps})
	}
	return json.Marshal(dtos)
}

func marshalReverse(r *domain.ReverseConfig) ([]byte, error) {
	if r == nil {
		return nil, nil
	}
	return json.Marshal(reverseDTO{Capability: r.Capability, On: r.On})
}

func mapSyncRoute(row sqlc.SyncRoute) (*domain.SyncRoute, error) {
	resourceTypes, err := unmarshalStringSlice(row.ResourceTypes)
	if err != nil {
		return nil, err
	}
	schedule, err := unmarshalJSONMap(row.Schedule)
	if err != nil {
		return nil, err
	}
	filter, err := unmarshalJSONMap(row.Filter)
	if err != nil {
		return nil, err
	}
	idempotency, err := unmarshalJSONMap(row.IdempotencyRule)
	if err != nil {
		return nil, err
	}
	activities, err := unmarshalActivities(row.Activities)
	if err != nil {
		return nil, err
	}
	reverse, err := unmarshalReverse(row.Reverse)
	if err != nil {
		return nil, err
	}
	retry, err := unmarshalJSONMap(row.RetryPolicy)
	if err != nil {
		return nil, err
	}
	var webhook []byte
	if len(row.WebhookSecret) > 0 {
		webhook, err = unmarshalEncryptedSecret(row.WebhookSecret)
		if err != nil {
			return nil, err
		}
	}
	return domain.ReconstituteSyncRoute(
		row.ID,
		row.WorkspaceID,
		row.SourceConnectionID,
		row.Name,
		domain.SyncRouteStatus(row.Status),
		domain.SyncRouteTrigger(row.TriggerType),
		resourceTypes,
		schedule,
		filter,
		idempotency,
		activities,
		reverse,
		retry,
		webhook,
		row.CreatedAt.Time,
		row.UpdatedAt.Time,
		timePtrFrom(row.DeletedAt),
	), nil
}

func mapWatermark(row sqlc.SyncRouteWatermark) (*domain.SyncRouteWatermark, error) {
	cursor, err := unmarshalJSONMap(row.Cursor)
	if err != nil {
		return nil, err
	}
	return &domain.SyncRouteWatermark{
		SyncRouteID:  row.SyncRouteID,
		ResourceType: row.ResourceType,
		Cursor:       cursor,
		UpdatedAt:    row.UpdatedAt.Time,
	}, nil
}

func unmarshalStringSlice(data []byte) ([]string, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var out []string
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func unmarshalActivities(data []byte) ([]domain.ActivityGroup, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var dtos []activityGroupDTO
	if err := json.Unmarshal(data, &dtos); err != nil {
		return nil, err
	}
	out := make([]domain.ActivityGroup, 0, len(dtos))
	for _, g := range dtos {
		steps := make([]domain.ActivityStep, 0, len(g.Steps))
		for _, s := range g.Steps {
			id, err := uuid.Parse(s.DestinationConnectionID)
			if err != nil {
				return nil, err
			}
			steps = append(steps, domain.ActivityStep{
				DestinationConnectionID: id,
				Capability:              s.Capability,
				MappingKey:              s.MappingKey,
			})
		}
		out = append(out, domain.ActivityGroup{
			Mode:  domain.ActivityGroupMode(g.Mode),
			Steps: steps,
		})
	}
	return out, nil
}

func unmarshalReverse(data []byte) (*domain.ReverseConfig, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var dto reverseDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, err
	}
	return &domain.ReverseConfig{Capability: dto.Capability, On: dto.On}, nil
}
