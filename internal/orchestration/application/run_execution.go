package application

import (
	"context"
	"fmt"
	"time"

	"hublio/internal/orchestration/domain"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/id"

	"github.com/google/uuid"
)

// RunExecutionResult is returned by RunExecution. When RequeueJob is non-nil the caller
// must enqueue it AFTER the surrounding transaction commits.
type RunExecutionResult struct {
	Execution  *domain.Execution
	RequeueJob *ExecutionJob
}

// RunExecution drives one Execution through its sequential Steps (validate ->
// transform_request -> invoke_connector -> transform_response -> publish_event) using the
// Fake/real Connector Runtime behind ConnectorGateway. Called by the worker for the
// orchestration.execution job. Transaction boundary is owned by the caller (worker handler).
func (s *Services) RunExecution(ctx context.Context, executionID uuid.UUID) (*RunExecutionResult, error) {
	execution, err := s.Executions.FindByID(ctx, executionID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	intent, err := s.Intents.FindByID(ctx, execution.IntentID())
	if err != nil {
		return nil, mapRepoErr(err)
	}

	now := s.clock().Now()
	switch execution.Status() {
	case domain.ExecutionStatusQueued:
		if err := execution.Start(now); err != nil {
			return nil, mapDomainErr(err)
		}
		if err := s.appendTimeline(execution, "execution_started", "execution started", nil, now); err != nil {
			return nil, err
		}
	case domain.ExecutionStatusRunning:
		// Re-delivered job on an already-running Execution: resume from the next pending step.
	default:
		// Terminal (or not yet queued) Execution: idempotent no-op for the worker.
		return &RunExecutionResult{Execution: execution}, nil
	}

	resolved, resolveErr := s.Connections.ResolveForIntent(ctx, intent.WorkspaceID(), intent.ConnectionID())

	for {
		step, ok := execution.NextPendingStep()
		if !ok {
			break
		}
		now = s.clock().Now()
		if err := execution.StartStep(step.StepNo(), now); err != nil {
			return nil, mapDomainErr(err)
		}

		stepErr := s.runStep(ctx, execution, intent, step, resolved, resolveErr)
		now = s.clock().Now()
		if stepErr != nil {
			_ = execution.FailStep(step.StepNo(), stepErrorCode(stepErr), stepErr.Error(), now)
			if err := s.appendTimeline(execution, "step_failed", stepErr.Error(), map[string]any{
				"step_no": step.StepNo(), "step_type": string(step.StepType()),
			}, now); err != nil {
				return nil, err
			}
			break
		}
		if err := execution.SucceedStep(step.StepNo(), now); err != nil {
			return nil, mapDomainErr(err)
		}
		if err := s.appendTimeline(execution, "step_succeeded", "", map[string]any{
			"step_no": step.StepNo(), "step_type": string(step.StepType()),
		}, now); err != nil {
			return nil, err
		}
	}

	now = s.clock().Now()
	if execution.AllStepsSucceeded() {
		if err := execution.Succeed(now); err != nil {
			return nil, mapDomainErr(err)
		}
		if err := s.appendTimeline(execution, "execution_succeeded", "", nil, now); err != nil {
			return nil, err
		}
		if err := s.Executions.Update(ctx, execution); err != nil {
			return nil, mapRepoErr(err)
		}
		return &RunExecutionResult{Execution: execution}, nil
	}

	reason := failureReason(execution)
	if err := execution.Fail(reason, now); err != nil {
		return nil, mapDomainErr(err)
	}

	if execution.CanRetry(s.maxRetries()) {
		if err := execution.ScheduleRetry(now); err != nil {
			return nil, mapDomainErr(err)
		}
		if err := s.appendTimeline(execution, "execution_retry_scheduled", reason, map[string]any{
			"retry_attempt": execution.RetryAttempt(),
		}, now); err != nil {
			return nil, err
		}
		if err := s.Executions.Update(ctx, execution); err != nil {
			return nil, mapRepoErr(err)
		}
		return &RunExecutionResult{
			Execution: execution,
			RequeueJob: &ExecutionJob{
				ExecutionID:    execution.ID(),
				IntentID:       intent.ID(),
				OrganizationID: intent.OrganizationID(),
				WorkspaceID:    intent.WorkspaceID(),
				CorrelationID:  intent.CorrelationID(),
			},
		}, nil
	}

	if err := execution.DeadLetter(now); err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.appendTimeline(execution, "execution_dead_lettered", reason, nil, now); err != nil {
		return nil, err
	}
	if err := s.Executions.Update(ctx, execution); err != nil {
		return nil, mapRepoErr(err)
	}
	return &RunExecutionResult{Execution: execution}, nil
}

func (s *Services) runStep(ctx context.Context, execution *domain.Execution, intent *domain.Intent, step *domain.ExecutionStep, resolved ResolvedConnection, resolveErr error) error {
	switch step.StepType() {
	case domain.StepTypeValidate:
		if resolveErr != nil {
			return resolveErr
		}
		return nil

	case domain.StepTypeTransformRequest:
		// Phase E replaces this passthrough with real canonical transform capabilities.
		execution.MergeContext(map[string]any{"request": intent.Payload()})
		return nil

	case domain.StepTypeInvokeConnector:
		return s.invokeConnectorStep(ctx, execution, intent, step, resolved)

	case domain.StepTypeTransformResponse:
		// Phase E replaces this passthrough with real canonical transform capabilities.
		if resp, ok := execution.Context()["invoke_response"]; ok {
			execution.MergeContext(map[string]any{"response": resp})
		}
		return nil

	case domain.StepTypePublishEvent:
		// Events BC (Phase F) will publish a durable runtime event here; for now we only
		// record the intent on the Execution timeline.
		return s.appendTimeline(execution, "publish_event", "execution runtime event recorded (Events BC pending)", map[string]any{
			"capability": intent.Capability(),
		}, s.clock().Now())

	default:
		return fmt.Errorf("orchestration: unknown step type %q", step.StepType())
	}
}

func (s *Services) invokeConnectorStep(ctx context.Context, execution *domain.Execution, intent *domain.Intent, step *domain.ExecutionStep, resolved ResolvedConnection) error {
	requestPayload, ok := execution.Context()["request"].(map[string]any)
	if !ok || requestPayload == nil {
		requestPayload = intent.Payload()
	}
	stepID := step.ID()
	now := s.clock().Now()

	if err := s.addSnapshot(execution, &stepID, domain.SnapshotTypeCanonicalRequest, requestPayload, "application/json", now); err != nil {
		return err
	}

	resp, err := s.Connectors.Invoke(ctx, resolved.ConnectorCode, InvokeRequest{
		ConnectionID: resolved.ConnectionID,
		Capability:   intent.Capability(),
		Config:       resolved.Config,
		Secret:       resolved.Secret,
		Payload:      requestPayload,
	})
	if err != nil {
		return err
	}

	if err := s.addSnapshot(execution, &stepID, domain.SnapshotTypeCanonicalResponse, resp.Payload, "application/json", s.clock().Now()); err != nil {
		return err
	}
	execution.MergeContext(map[string]any{"invoke_response": resp.Payload, "invoke_metadata": resp.Metadata})
	return nil
}

func (s *Services) appendTimeline(execution *domain.Execution, event, message string, metadata map[string]any, now time.Time) error {
	timelineID, err := id.NewV7()
	if err != nil {
		return apperr.Wrap(err, "failed to generate timeline id", apperr.ErrCodeInternal)
	}
	execution.AppendTimeline(timelineID, event, message, metadata, now)
	return nil
}

func (s *Services) addSnapshot(execution *domain.Execution, stepID *uuid.UUID, snapshotType domain.SnapshotType, snapshot map[string]any, contentType string, now time.Time) error {
	snapshotID, err := id.NewV7()
	if err != nil {
		return apperr.Wrap(err, "failed to generate snapshot id", apperr.ErrCodeInternal)
	}
	execution.AddSnapshot(snapshotID, stepID, snapshotType, snapshot, contentType, now)
	return nil
}

func stepErrorCode(err error) string {
	if ae, ok := err.(*apperr.AppError); ok {
		return string(ae.Code)
	}
	return "STEP_FAILED"
}

func failureReason(execution *domain.Execution) string {
	for _, step := range execution.Steps() {
		if step.Status() == domain.StepStatusFailed {
			if step.ErrorMessage() != nil {
				return fmt.Sprintf("step %d (%s) failed: %s", step.StepNo(), step.StepType(), *step.ErrorMessage())
			}
			return fmt.Sprintf("step %d (%s) failed", step.StepNo(), step.StepType())
		}
	}
	return "execution failed"
}
