package infrastructure

import (
	"context"

	"hublio/internal/orchestration/domain"
	"hublio/internal/platform/persistence"
	"hublio/internal/platform/persistence/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ExecutionRepository persists the Execution aggregate (execution row) together with its
// Steps, Snapshots and Timeline children. Snapshots/Timeline are append-only: Update()
// re-submits every in-memory entry with an idempotent ON CONFLICT (id) DO NOTHING insert so
// already-persisted rows are safely skipped.
type ExecutionRepository struct {
	pool *pgxpool.Pool
}

func NewExecutionRepository(pool *pgxpool.Pool) *ExecutionRepository {
	return &ExecutionRepository{pool: pool}
}

func (r *ExecutionRepository) q(ctx context.Context) *sqlc.Queries {
	return sqlc.New(persistence.Conn(ctx, r.pool))
}

func (r *ExecutionRepository) Save(ctx context.Context, execution *domain.Execution) error {
	q := r.q(ctx)
	execCtx, err := marshalJSONMap(execution.Context())
	if err != nil {
		return err
	}
	if err := mapUnique(q.InsertExecution(ctx, sqlc.InsertExecutionParams{
		ID:            execution.ID(),
		IntentID:      execution.IntentID(),
		Status:        string(execution.Status()),
		Result:        resultParam(execution.Result()),
		RetryAttempt:  int32(execution.RetryAttempt()),
		CurrentStepNo: int32(execution.CurrentStepNo()),
		Context:       execCtx,
		FailureReason: execution.FailureReason(),
		StartedAt:     timestamptzPtr(execution.StartedAt()),
		CompletedAt:   timestamptzPtr(execution.CompletedAt()),
		CreatedAt:     timestamptz(execution.CreatedAt()),
	})); err != nil {
		return err
	}

	for _, step := range execution.Steps() {
		if err := q.InsertExecutionStep(ctx, sqlc.InsertExecutionStepParams{
			ID:           step.ID(),
			ExecutionID:  execution.ID(),
			StepNo:       int32(step.StepNo()),
			StepType:     string(step.StepType()),
			Status:       string(step.Status()),
			RetryAttempt: int32(step.RetryAttempt()),
			DurationMs:   int32PtrFromIntPtr(step.DurationMs()),
			ErrorMessage: step.ErrorMessage(),
			ErrorCode:    step.ErrorCode(),
			StartedAt:    timestamptzPtr(step.StartedAt()),
			CompletedAt:  timestamptzPtr(step.CompletedAt()),
		}); err != nil {
			return mapUnique(err)
		}
	}
	return nil
}

func (r *ExecutionRepository) Update(ctx context.Context, execution *domain.Execution) error {
	q := r.q(ctx)
	execCtx, err := marshalJSONMap(execution.Context())
	if err != nil {
		return err
	}
	if err := mapUnique(q.UpdateExecution(ctx, sqlc.UpdateExecutionParams{
		ID:            execution.ID(),
		Status:        string(execution.Status()),
		Result:        resultParam(execution.Result()),
		RetryAttempt:  int32(execution.RetryAttempt()),
		CurrentStepNo: int32(execution.CurrentStepNo()),
		Context:       execCtx,
		FailureReason: execution.FailureReason(),
		StartedAt:     timestamptzPtr(execution.StartedAt()),
		CompletedAt:   timestamptzPtr(execution.CompletedAt()),
	})); err != nil {
		return err
	}

	for _, step := range execution.Steps() {
		if err := q.UpdateExecutionStep(ctx, sqlc.UpdateExecutionStepParams{
			ID:           step.ID(),
			Status:       string(step.Status()),
			RetryAttempt: int32(step.RetryAttempt()),
			DurationMs:   int32PtrFromIntPtr(step.DurationMs()),
			ErrorMessage: step.ErrorMessage(),
			ErrorCode:    step.ErrorCode(),
			StartedAt:    timestamptzPtr(step.StartedAt()),
			CompletedAt:  timestamptzPtr(step.CompletedAt()),
		}); err != nil {
			return mapUnique(err)
		}
	}

	for _, snap := range execution.Snapshots() {
		snapshot, err := marshalJSONMap(snap.Snapshot())
		if err != nil {
			return err
		}
		if err := q.InsertExecutionSnapshot(ctx, sqlc.InsertExecutionSnapshotParams{
			ID:           snap.ID(),
			ExecutionID:  execution.ID(),
			StepID:       uuidPtrToPgtype(snap.StepID()),
			SnapshotType: string(snap.SnapshotType()),
			Snapshot:     snapshot,
			ContentType:  strPtr(snap.ContentType()),
			CreatedAt:    timestamptz(snap.CreatedAt()),
		}); err != nil {
			return mapUnique(err)
		}
	}

	for _, entry := range execution.Timeline() {
		metadata, err := marshalJSONMap(entry.Metadata())
		if err != nil {
			return err
		}
		if err := q.InsertExecutionTimeline(ctx, sqlc.InsertExecutionTimelineParams{
			ID:          entry.ID(),
			ExecutionID: execution.ID(),
			Event:       entry.Event(),
			Message:     strPtr(entry.Message()),
			Metadata:    metadata,
			CreatedAt:   timestamptz(entry.CreatedAt()),
		}); err != nil {
			return mapUnique(err)
		}
	}

	return nil
}

func (r *ExecutionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Execution, error) {
	row, err := r.q(ctx).GetExecutionByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return r.hydrate(ctx, row)
}

func (r *ExecutionRepository) FindByIntentID(ctx context.Context, intentID uuid.UUID) (*domain.Execution, error) {
	row, err := r.q(ctx).GetExecutionByIntentID(ctx, intentID)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return r.hydrate(ctx, row)
}

func (r *ExecutionRepository) ListByIntentID(ctx context.Context, intentID uuid.UUID) ([]*domain.Execution, error) {
	rows, err := r.q(ctx).ListExecutionsByIntentID(ctx, intentID)
	if err != nil {
		return nil, mapNotFound(err)
	}
	out := make([]*domain.Execution, 0, len(rows))
	for _, row := range rows {
		exec, err := r.hydrate(ctx, row)
		if err != nil {
			return nil, err
		}
		out = append(out, exec)
	}
	return out, nil
}

func (r *ExecutionRepository) hydrate(ctx context.Context, row sqlc.Execution) (*domain.Execution, error) {
	q := r.q(ctx)

	execCtx, err := unmarshalJSONMap(row.Context)
	if err != nil {
		return nil, err
	}

	stepRows, err := q.ListExecutionStepsByExecution(ctx, row.ID)
	if err != nil {
		return nil, mapNotFound(err)
	}
	steps := make([]*domain.ExecutionStep, 0, len(stepRows))
	for _, sr := range stepRows {
		steps = append(steps, domain.ReconstituteExecutionStep(
			sr.ID,
			sr.ExecutionID,
			int(sr.StepNo),
			domain.ExecutionStepType(sr.StepType),
			domain.ExecutionStepStatus(sr.Status),
			int(sr.RetryAttempt),
			intPtrFromInt32Ptr(sr.DurationMs),
			sr.ErrorMessage,
			sr.ErrorCode,
			timePtrFrom(sr.StartedAt),
			timePtrFrom(sr.CompletedAt),
		))
	}

	snapshotRows, err := q.ListExecutionSnapshotsByExecution(ctx, row.ID)
	if err != nil {
		return nil, mapNotFound(err)
	}
	snapshots := make([]*domain.ExecutionSnapshot, 0, len(snapshotRows))
	for _, snr := range snapshotRows {
		snapshot, err := unmarshalJSONMap(snr.Snapshot)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, domain.ReconstituteExecutionSnapshot(
			snr.ID,
			snr.ExecutionID,
			pgtypeToUUIDPtr(snr.StepID),
			domain.SnapshotType(snr.SnapshotType),
			snapshot,
			strFromPtr(snr.ContentType),
			timeFrom(snr.CreatedAt),
		))
	}

	timelineRows, err := q.ListExecutionTimelinesByExecution(ctx, row.ID)
	if err != nil {
		return nil, mapNotFound(err)
	}
	timeline := make([]*domain.TimelineEntry, 0, len(timelineRows))
	for _, tr := range timelineRows {
		metadata, err := unmarshalJSONMap(tr.Metadata)
		if err != nil {
			return nil, err
		}
		timeline = append(timeline, domain.ReconstituteTimelineEntry(
			tr.ID,
			tr.ExecutionID,
			tr.Event,
			strFromPtr(tr.Message),
			metadata,
			timeFrom(tr.CreatedAt),
		))
	}

	var result *domain.ExecutionResult
	if s := nullableEnumPtr(row.Result); s != nil {
		r := domain.ExecutionResult(*s)
		result = &r
	}

	return domain.ReconstituteExecution(
		row.ID,
		row.IntentID,
		domain.ExecutionStatus(row.Status),
		result,
		int(row.RetryAttempt),
		int(row.CurrentStepNo),
		execCtx,
		row.FailureReason,
		timePtrFrom(row.StartedAt),
		timePtrFrom(row.CompletedAt),
		timeFrom(row.CreatedAt),
		steps,
		snapshots,
		timeline,
	), nil
}

func resultParam(result *domain.ExecutionResult) any {
	if result == nil {
		return nil
	}
	return string(*result)
}

func int32PtrFromIntPtr(v *int) *int32 {
	if v == nil {
		return nil
	}
	i := int32(*v)
	return &i
}

func intPtrFromInt32Ptr(v *int32) *int {
	if v == nil {
		return nil
	}
	i := int(*v)
	return &i
}
