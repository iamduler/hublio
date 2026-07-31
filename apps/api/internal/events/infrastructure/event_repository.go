package infrastructure

import (
	"context"
	"fmt"

	"hublio/internal/events/domain"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/persistence/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

// ListByWorkspace returns a keyset page of PlatformEvents for a Workspace.
func (r *EventRepository) ListByWorkspace(
	ctx context.Context,
	workspaceID uuid.UUID,
	filter domain.EventListFilter,
) (*domain.EventListPage, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	params := sqlc.ListEventsFilteredParams{
		WorkspaceID: uuidPtrToPgtype(&workspaceID),
		ExecutionID: uuidPtrToPgtype(filter.ExecutionID),
		FetchLimit:  limit + 1,
	}
	if filter.Category != nil {
		cat := string(*filter.Category)
		params.Category = &cat
	}
	if filter.Cursor != nil {
		params.CursorCreatedAt = timestamptz(filter.Cursor.CreatedAt)
		params.CursorID = uuidPtrToPgtype(&filter.Cursor.ID)
	} else {
		params.CursorCreatedAt = pgtype.Timestamptz{Valid: false}
		params.CursorID = pgtype.UUID{Valid: false}
	}

	rows, err := r.q(ctx).ListEventsFiltered(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("events repo: list filtered: %w", err)
	}

	events, err := hydrateEvents(rows)
	if err != nil {
		return nil, err
	}

	page := &domain.EventListPage{Events: events, HasNext: false}
	if int32(len(events)) > limit {
		page.HasNext = true
		page.Events = events[:limit]
	}
	if page.HasNext && len(page.Events) > 0 {
		last := page.Events[len(page.Events)-1]
		page.Next = &domain.EventListCursor{
			CreatedAt: last.CreatedAt(),
			ID:        last.ID(),
		}
	}
	return page, nil
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
