package infrastructure

import (
	"context"

	orchestrationapp "hublio/internal/orchestration/application"
	"hublio/internal/platform/queue"
)

// QueueJobEnqueuer adapts the Platform work queue (Infrastructure) into the Orchestration
// Application's JobEnqueuer port.
type QueueJobEnqueuer struct {
	queue queue.Queue
}

func NewQueueJobEnqueuer(q queue.Queue) *QueueJobEnqueuer {
	return &QueueJobEnqueuer{queue: q}
}

func (e *QueueJobEnqueuer) EnqueueExecution(ctx context.Context, job orchestrationapp.ExecutionJob) error {
	return queue.EnqueueExecution(ctx, e.queue, map[string]any{
		"execution_id":    job.ExecutionID.String(),
		"intent_id":       job.IntentID.String(),
		"organization_id": job.OrganizationID.String(),
		"workspace_id":    job.WorkspaceID.String(),
		"correlation_id":  job.CorrelationID,
	})
}
