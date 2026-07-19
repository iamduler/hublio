package infrastructure

import (
	"context"
	"fmt"

	"hublio/internal/events/domain"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/persistence/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EventRepository is the Postgres implementation of domain.EventRepository. It only ever
// inserts: the `events` table is append-only and immutable.
type EventRepository struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{pool: pool}
}

func (r *EventRepository) q(ctx context.Context) *sqlc.Queries {
	return sqlc.New(persistence.Conn(ctx, r.pool))
}

func (r *EventRepository) Save(ctx context.Context, event *domain.PlatformEvent) error {
	payload, err := marshalJSONMap(event.Payload())
	if err != nil {
		return fmt.Errorf("events repo: marshal payload: %w", err)
	}
	metadata, err := marshalJSONMap(event.Metadata())
	if err != nil {
		return fmt.Errorf("events repo: marshal metadata: %w", err)
	}

	if err := r.q(ctx).InsertEvent(ctx, sqlc.InsertEventParams{
		ID:             event.ID(),
		OrganizationID: uuidPtrToPgtype(event.OrganizationID()),
		WorkspaceID:    uuidPtrToPgtype(event.WorkspaceID()),
		AggregateType:  string(event.AggregateType()),
		AggregateID:    event.AggregateID(),
		ExecutionID:    uuidPtrToPgtype(event.ExecutionID()),
		Category:       string(event.Category()),
		EventName:      event.EventName(),
		CorrelationID:  strPtr(event.CorrelationID()),
		Payload:        payload,
		Metadata:       metadata,
		PublishedBy:    strPtr(event.PublishedBy()),
		CreatedAt:      timestamptz(event.CreatedAt()),
	}); err != nil {
		return fmt.Errorf("events repo: insert event: %w", err)
	}
	return nil
}

// ListByWorkspace returns the most recent PlatformEvents for a Workspace (optionally
// filtered by executionID), newest first, bounded by limit. Used by the Platform Events API
// (GET /api/v1/events).
func (r *EventRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, executionID *uuid.UUID, limit int32) ([]*domain.PlatformEvent, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	if executionID != nil {
		rows, err := r.q(ctx).ListEventsByWorkspaceAndExecution(ctx, sqlc.ListEventsByWorkspaceAndExecutionParams{
			WorkspaceID: uuidPtrToPgtype(&workspaceID),
			ExecutionID: uuidPtrToPgtype(executionID),
			Limit:       limit,
		})
		if err != nil {
			return nil, fmt.Errorf("events repo: list by workspace+execution: %w", err)
		}
		return hydrateEvents(rows)
	}

	rows, err := r.q(ctx).ListEventsByWorkspace(ctx, sqlc.ListEventsByWorkspaceParams{
		WorkspaceID: uuidPtrToPgtype(&workspaceID),
		Limit:       limit,
	})
	if err != nil {
		return nil, fmt.Errorf("events repo: list by workspace: %w", err)
	}
	return hydrateEvents(rows)
}

func hydrateEvents(rows []sqlc.Event) ([]*domain.PlatformEvent, error) {
	out := make([]*domain.PlatformEvent, 0, len(rows))
	for _, row := range rows {
		payload, err := unmarshalJSONMap(row.Payload)
		if err != nil {
			return nil, fmt.Errorf("events repo: unmarshal payload: %w", err)
		}
		metadata, err := unmarshalJSONMap(row.Metadata)
		if err != nil {
			return nil, fmt.Errorf("events repo: unmarshal metadata: %w", err)
		}
		out = append(out, domain.ReconstitutePlatformEvent(
			row.ID,
			pgtypeToUUIDPtr(row.OrganizationID),
			pgtypeToUUIDPtr(row.WorkspaceID),
			domain.AggregateType(row.AggregateType),
			row.AggregateID,
			pgtypeToUUIDPtr(row.ExecutionID),
			domain.Category(row.Category),
			row.EventName,
			strFromPtr(row.CorrelationID),
			payload,
			metadata,
			strFromPtr(row.PublishedBy),
			timeFrom(row.CreatedAt),
		))
	}
	return out, nil
}
