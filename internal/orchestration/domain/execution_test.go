package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func fiveStepIDs() []uuid.UUID {
	return []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()}
}

func newTestExecution(t *testing.T) *Execution {
	t.Helper()
	exec, err := NewExecution(uuid.New(), uuid.New(), fiveStepIDs(), time.Now())
	if err != nil {
		t.Fatalf("NewExecution() unexpected error: %v", err)
	}
	return exec
}

func TestNewExecution(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		id       uuid.UUID
		intentID uuid.UUID
		stepIDs  []uuid.UUID
		wantErr  error
	}{
		{name: "valid", id: uuid.New(), intentID: uuid.New(), stepIDs: fiveStepIDs()},
		{name: "nil id", id: uuid.Nil, intentID: uuid.New(), stepIDs: fiveStepIDs(), wantErr: ErrInvalidID},
		{name: "nil intent id", id: uuid.New(), intentID: uuid.Nil, stepIDs: fiveStepIDs(), wantErr: ErrInvalidID},
		{name: "wrong step count", id: uuid.New(), intentID: uuid.New(), stepIDs: []uuid.UUID{uuid.New()}, wantErr: ErrInvalidStepCount},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exec, err := NewExecution(tt.id, tt.intentID, tt.stepIDs, now)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if exec.Status() != ExecutionStatusCreated {
				t.Fatalf("expected status created, got %s", exec.Status())
			}
			if len(exec.Steps()) != 5 {
				t.Fatalf("expected 5 steps, got %d", len(exec.Steps()))
			}
			for i, step := range exec.Steps() {
				if step.StepNo() != i+1 {
					t.Fatalf("expected step_no %d, got %d", i+1, step.StepNo())
				}
				if step.Status() != StepStatusPending {
					t.Fatalf("expected pending step, got %s", step.Status())
				}
			}
			events := exec.PullEvents()
			if len(events) != 1 || events[0].Name != EventExecutionCreated {
				t.Fatalf("expected one ExecutionCreated event, got %+v", events)
			}
		})
	}
}

func TestExecution_LifecycleHappyPath(t *testing.T) {
	exec := newTestExecution(t)
	now := time.Now()

	if err := exec.Queue(now); err != nil {
		t.Fatalf("Queue() unexpected error: %v", err)
	}
	if exec.Status() != ExecutionStatusQueued {
		t.Fatalf("expected queued, got %s", exec.Status())
	}

	if err := exec.Start(now); err != nil {
		t.Fatalf("Start() unexpected error: %v", err)
	}
	if exec.Status() != ExecutionStatusRunning {
		t.Fatalf("expected running, got %s", exec.Status())
	}
	if exec.StartedAt() == nil {
		t.Fatalf("expected StartedAt to be set")
	}

	if err := exec.Succeed(now); err == nil {
		t.Fatalf("expected error succeeding before steps complete")
	} else if err != ErrStepsIncomplete {
		t.Fatalf("expected ErrStepsIncomplete, got %v", err)
	}

	for _, step := range exec.Steps() {
		if err := exec.StartStep(step.StepNo(), now); err != nil {
			t.Fatalf("StartStep(%d) unexpected error: %v", step.StepNo(), err)
		}
		if err := exec.SucceedStep(step.StepNo(), now); err != nil {
			t.Fatalf("SucceedStep(%d) unexpected error: %v", step.StepNo(), err)
		}
	}

	if !exec.AllStepsSucceeded() {
		t.Fatalf("expected all steps succeeded")
	}

	if err := exec.Succeed(now); err != nil {
		t.Fatalf("Succeed() unexpected error: %v", err)
	}
	if exec.Status() != ExecutionStatusSucceeded {
		t.Fatalf("expected succeeded, got %s", exec.Status())
	}
	if exec.Result() == nil || *exec.Result() != ExecutionResultSuccess {
		t.Fatalf("expected result success, got %v", exec.Result())
	}
	if exec.CompletedAt() == nil {
		t.Fatalf("expected CompletedAt to be set")
	}
}

func TestExecution_InvalidTransitions(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		run     func(*Execution) error
		prepare func(*Execution)
	}{
		{
			name: "start before queue",
			run:  func(e *Execution) error { return e.Start(now) },
		},
		{
			name: "succeed before running",
			run:  func(e *Execution) error { return e.Succeed(now) },
		},
		{
			name: "fail before running",
			run:  func(e *Execution) error { return e.Fail("boom", now) },
		},
		{
			name: "queue twice",
			prepare: func(e *Execution) {
				_ = e.Queue(now)
			},
			run: func(e *Execution) error { return e.Queue(now) },
		},
		{
			name: "schedule retry when not failed",
			run:  func(e *Execution) error { return e.ScheduleRetry(now) },
		},
		{
			name: "dead letter when not failed",
			run:  func(e *Execution) error { return e.DeadLetter(now) },
		},
		{
			name: "cancel after succeeded",
			prepare: func(e *Execution) {
				_ = e.Queue(now)
				_ = e.Start(now)
				for _, s := range e.Steps() {
					_ = e.StartStep(s.StepNo(), now)
					_ = e.SucceedStep(s.StepNo(), now)
				}
				_ = e.Succeed(now)
			},
			run: func(e *Execution) error { return e.Cancel(now) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exec := newTestExecution(t)
			if tt.prepare != nil {
				tt.prepare(exec)
			}
			if err := tt.run(exec); err != ErrInvalidTransition {
				t.Fatalf("expected ErrInvalidTransition, got %v", err)
			}
		})
	}
}

func TestExecution_FailAndRetry(t *testing.T) {
	exec := newTestExecution(t)
	now := time.Now()
	_ = exec.Queue(now)
	_ = exec.Start(now)

	step := exec.Steps()[0]
	if err := exec.StartStep(step.StepNo(), now); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := exec.FailStep(step.StepNo(), "CONNECTOR_ERROR", "boom", now); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := exec.Fail("step 1 failed: boom", now); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.Status() != ExecutionStatusFailed {
		t.Fatalf("expected failed, got %s", exec.Status())
	}
	if !exec.CanRetry(3) {
		t.Fatalf("expected CanRetry(3) true at attempt 0")
	}

	if err := exec.ScheduleRetry(now); err != nil {
		t.Fatalf("ScheduleRetry() unexpected error: %v", err)
	}
	if exec.Status() != ExecutionStatusQueued {
		t.Fatalf("expected queued after retry, got %s", exec.Status())
	}
	if exec.RetryAttempt() != 1 {
		t.Fatalf("expected retry_attempt 1, got %d", exec.RetryAttempt())
	}
	for _, s := range exec.Steps() {
		if s.Status() != StepStatusPending {
			t.Fatalf("expected step reset to pending, got %s", s.Status())
		}
	}
}

func TestExecution_MaxRetriesThenDeadLetter(t *testing.T) {
	exec := newTestExecution(t)
	now := time.Now()
	_ = exec.Queue(now)

	for attempt := 0; attempt < 10; attempt++ {
		_ = exec.Start(now)
		step := exec.Steps()[0]
		_ = exec.StartStep(step.StepNo(), now)
		_ = exec.FailStep(step.StepNo(), "ERR", "boom", now)
		_ = exec.Fail("boom", now)
		if !exec.CanRetry(3) {
			break
		}
		_ = exec.ScheduleRetry(now)
	}

	if exec.CanRetry(3) {
		t.Fatalf("expected retries exhausted")
	}
	if err := exec.DeadLetter(now); err != nil {
		t.Fatalf("DeadLetter() unexpected error: %v", err)
	}
	if exec.Status() != ExecutionStatusDeadLetter {
		t.Fatalf("expected dead_letter, got %s", exec.Status())
	}
	if exec.Result() == nil || *exec.Result() != ExecutionResultDeadLetter {
		t.Fatalf("expected result dead_letter, got %v", exec.Result())
	}
}

func TestExecution_CancelFromQueuedAndRunning(t *testing.T) {
	now := time.Now()

	t.Run("cancel queued", func(t *testing.T) {
		exec := newTestExecution(t)
		_ = exec.Queue(now)
		if err := exec.Cancel(now); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if exec.Status() != ExecutionStatusCancelled {
			t.Fatalf("expected cancelled, got %s", exec.Status())
		}
	})

	t.Run("cancel running", func(t *testing.T) {
		exec := newTestExecution(t)
		_ = exec.Queue(now)
		_ = exec.Start(now)
		if err := exec.Cancel(now); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if exec.Status() != ExecutionStatusCancelled {
			t.Fatalf("expected cancelled, got %s", exec.Status())
		}
	})
}

func TestExecution_AppendTimelineAndSnapshot(t *testing.T) {
	exec := newTestExecution(t)
	now := time.Now()

	entry := exec.AppendTimeline(uuid.New(), "execution_created", "created", map[string]any{"foo": "bar"}, now)
	if entry == nil || len(exec.Timeline()) != 1 {
		t.Fatalf("expected one timeline entry")
	}

	stepID := exec.Steps()[2].ID()
	snap := exec.AddSnapshot(uuid.New(), &stepID, SnapshotTypeCanonicalRequest, map[string]any{"amount": 1}, "application/json", now)
	if snap == nil || len(exec.Snapshots()) != 1 {
		t.Fatalf("expected one snapshot")
	}
	if exec.Snapshots()[0].StepID() == nil || *exec.Snapshots()[0].StepID() != stepID {
		t.Fatalf("expected snapshot linked to step")
	}
}

func TestExecution_StepNotFound(t *testing.T) {
	exec := newTestExecution(t)
	now := time.Now()
	if err := exec.StartStep(99, now); err != ErrStepNotFound {
		t.Fatalf("expected ErrStepNotFound, got %v", err)
	}
}
