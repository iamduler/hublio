package domain

import (
	"time"

	"github.com/google/uuid"
)

type ExecutionStatus string

const (
	ExecutionStatusCreated    ExecutionStatus = "created"
	ExecutionStatusQueued     ExecutionStatus = "queued"
	ExecutionStatusRunning    ExecutionStatus = "running"
	ExecutionStatusSucceeded  ExecutionStatus = "succeeded"
	ExecutionStatusFailed     ExecutionStatus = "failed"
	ExecutionStatusCancelled  ExecutionStatus = "cancelled"
	ExecutionStatusExpired    ExecutionStatus = "expired"
	ExecutionStatusDeadLetter ExecutionStatus = "dead_letter"
)

type ExecutionResult string

const (
	ExecutionResultSuccess    ExecutionResult = "success"
	ExecutionResultFailed     ExecutionResult = "failed"
	ExecutionResultCancelled  ExecutionResult = "cancelled"
	ExecutionResultExpired    ExecutionResult = "expired"
	ExecutionResultDeadLetter ExecutionResult = "dead_letter"
)

type ExecutionStepType string

const (
	StepTypeValidate          ExecutionStepType = "validate"
	StepTypeTransformRequest  ExecutionStepType = "transform_request"
	StepTypeInvokeConnector   ExecutionStepType = "invoke_connector"
	StepTypeTransformResponse ExecutionStepType = "transform_response"
	StepTypePublishEvent      ExecutionStepType = "publish_event"
)

// DefaultStepTypes returns the frozen, sequential v1 Execution Step pipeline.
func DefaultStepTypes() []ExecutionStepType {
	return []ExecutionStepType{
		StepTypeValidate,
		StepTypeTransformRequest,
		StepTypeInvokeConnector,
		StepTypeTransformResponse,
		StepTypePublishEvent,
	}
}

type ExecutionStepStatus string

const (
	StepStatusPending   ExecutionStepStatus = "pending"
	StepStatusRunning   ExecutionStepStatus = "running"
	StepStatusSucceeded ExecutionStepStatus = "succeeded"
	StepStatusFailed    ExecutionStepStatus = "failed"
)

// ExecutionStep is a sequential child step of an Execution.
type ExecutionStep struct {
	id           uuid.UUID
	executionID  uuid.UUID
	stepNo       int
	stepType     ExecutionStepType
	status       ExecutionStepStatus
	retryAttempt int
	durationMs   *int
	errorMessage *string
	errorCode    *string
	startedAt    *time.Time
	completedAt  *time.Time
}

func ReconstituteExecutionStep(
	id, executionID uuid.UUID,
	stepNo int,
	stepType ExecutionStepType,
	status ExecutionStepStatus,
	retryAttempt int,
	durationMs *int,
	errorMessage, errorCode *string,
	startedAt, completedAt *time.Time,
) *ExecutionStep {
	return &ExecutionStep{
		id:           id,
		executionID:  executionID,
		stepNo:       stepNo,
		stepType:     stepType,
		status:       status,
		retryAttempt: retryAttempt,
		durationMs:   durationMs,
		errorMessage: errorMessage,
		errorCode:    errorCode,
		startedAt:    startedAt,
		completedAt:  completedAt,
	}
}

func (s *ExecutionStep) ID() uuid.UUID               { return s.id }
func (s *ExecutionStep) ExecutionID() uuid.UUID      { return s.executionID }
func (s *ExecutionStep) StepNo() int                 { return s.stepNo }
func (s *ExecutionStep) StepType() ExecutionStepType { return s.stepType }
func (s *ExecutionStep) Status() ExecutionStepStatus { return s.status }
func (s *ExecutionStep) RetryAttempt() int           { return s.retryAttempt }
func (s *ExecutionStep) DurationMs() *int            { return s.durationMs }
func (s *ExecutionStep) ErrorMessage() *string       { return s.errorMessage }
func (s *ExecutionStep) ErrorCode() *string          { return s.errorCode }
func (s *ExecutionStep) StartedAt() *time.Time       { return s.startedAt }
func (s *ExecutionStep) CompletedAt() *time.Time     { return s.completedAt }

func (s *ExecutionStep) reset() {
	s.status = StepStatusPending
	s.durationMs = nil
	s.errorMessage = nil
	s.errorCode = nil
	s.startedAt = nil
	s.completedAt = nil
}

func (s *ExecutionStep) start(now time.Time) error {
	if s.status != StepStatusPending {
		return ErrInvalidTransition
	}
	at := now.UTC()
	s.status = StepStatusRunning
	s.startedAt = &at
	return nil
}

func (s *ExecutionStep) succeed(now time.Time) error {
	if s.status != StepStatusRunning {
		return ErrInvalidTransition
	}
	at := now.UTC()
	s.status = StepStatusSucceeded
	s.completedAt = &at
	s.durationMs = durationMillis(s.startedAt, &at)
	return nil
}

func (s *ExecutionStep) fail(errorCode, errorMessage string, now time.Time) error {
	if s.status != StepStatusRunning {
		return ErrInvalidTransition
	}
	at := now.UTC()
	s.status = StepStatusFailed
	s.completedAt = &at
	s.durationMs = durationMillis(s.startedAt, &at)
	if errorCode != "" {
		s.errorCode = &errorCode
	}
	if errorMessage != "" {
		s.errorMessage = &errorMessage
	}
	return nil
}

func durationMillis(start, end *time.Time) *int {
	if start == nil || end == nil {
		return nil
	}
	ms := int(end.Sub(*start).Milliseconds())
	return &ms
}

type SnapshotType string

const (
	SnapshotTypeCanonicalRequest  SnapshotType = "canonical_request"
	SnapshotTypeProviderRequest   SnapshotType = "provider_request"
	SnapshotTypeProviderResponse  SnapshotType = "provider_response"
	SnapshotTypeCanonicalResponse SnapshotType = "canonical_response"
)

// ExecutionSnapshot is an immutable, append-only capture of request/response payloads.
type ExecutionSnapshot struct {
	id           uuid.UUID
	executionID  uuid.UUID
	stepID       *uuid.UUID
	snapshotType SnapshotType
	snapshot     map[string]any
	contentType  string
	createdAt    time.Time
}

func ReconstituteExecutionSnapshot(
	id, executionID uuid.UUID,
	stepID *uuid.UUID,
	snapshotType SnapshotType,
	snapshot map[string]any,
	contentType string,
	createdAt time.Time,
) *ExecutionSnapshot {
	return &ExecutionSnapshot{
		id:           id,
		executionID:  executionID,
		stepID:       stepID,
		snapshotType: snapshotType,
		snapshot:     snapshot,
		contentType:  contentType,
		createdAt:    createdAt,
	}
}

func (s *ExecutionSnapshot) ID() uuid.UUID              { return s.id }
func (s *ExecutionSnapshot) ExecutionID() uuid.UUID     { return s.executionID }
func (s *ExecutionSnapshot) StepID() *uuid.UUID         { return s.stepID }
func (s *ExecutionSnapshot) SnapshotType() SnapshotType { return s.snapshotType }
func (s *ExecutionSnapshot) Snapshot() map[string]any   { return s.snapshot }
func (s *ExecutionSnapshot) ContentType() string        { return s.contentType }
func (s *ExecutionSnapshot) CreatedAt() time.Time       { return s.createdAt }

// TimelineEntry is an immutable, append-only Execution history entry.
type TimelineEntry struct {
	id          uuid.UUID
	executionID uuid.UUID
	event       string
	message     string
	metadata    map[string]any
	createdAt   time.Time
}

func ReconstituteTimelineEntry(
	id, executionID uuid.UUID,
	event, message string,
	metadata map[string]any,
	createdAt time.Time,
) *TimelineEntry {
	return &TimelineEntry{
		id:          id,
		executionID: executionID,
		event:       event,
		message:     message,
		metadata:    metadata,
		createdAt:   createdAt,
	}
}

func (t *TimelineEntry) ID() uuid.UUID            { return t.id }
func (t *TimelineEntry) ExecutionID() uuid.UUID   { return t.executionID }
func (t *TimelineEntry) Event() string            { return t.event }
func (t *TimelineEntry) Message() string          { return t.message }
func (t *TimelineEntry) Metadata() map[string]any { return t.metadata }
func (t *TimelineEntry) CreatedAt() time.Time     { return t.createdAt }

const defaultMaxRetries = 3

// Execution is the Runtime aggregate driving one Intent through its sequential Steps.
// States: Created -> Queued -> Running -> Succeeded | Failed | Cancelled | Expired.
// Failed -> Queued (retry) | DeadLetter. One Execution per Intent in v1 (executions.intent_id UNIQUE).
type Execution struct {
	eventRecorder

	id            uuid.UUID
	intentID      uuid.UUID
	status        ExecutionStatus
	result        *ExecutionResult
	retryAttempt  int
	currentStepNo int
	context       map[string]any
	failureReason *string
	startedAt     *time.Time
	completedAt   *time.Time
	createdAt     time.Time

	steps     []*ExecutionStep
	snapshots []*ExecutionSnapshot
	timeline  []*TimelineEntry
}

// NewExecution creates a Created Execution for intentID with the frozen 5-step pipeline.
// stepIDs must contain exactly 5 application-generated UUIDs (one per DefaultStepTypes()).
func NewExecution(id, intentID uuid.UUID, stepIDs []uuid.UUID, now time.Time) (*Execution, error) {
	if id == uuid.Nil || intentID == uuid.Nil {
		return nil, ErrInvalidID
	}
	stepTypes := DefaultStepTypes()
	if len(stepIDs) != len(stepTypes) {
		return nil, ErrInvalidStepCount
	}

	exec := &Execution{
		id:            id,
		intentID:      intentID,
		status:        ExecutionStatusCreated,
		currentStepNo: 0,
		context:       map[string]any{},
		createdAt:     now.UTC(),
	}
	for i, stepType := range stepTypes {
		exec.steps = append(exec.steps, &ExecutionStep{
			id:          stepIDs[i],
			executionID: id,
			stepNo:      i + 1,
			stepType:    stepType,
			status:      StepStatusPending,
		})
	}
	exec.record(EventExecutionCreated, id, now.UTC(), map[string]any{"intent_id": intentID.String()})
	return exec, nil
}

// ReconstituteExecution hydrates an Execution aggregate (with children) from persistence.
func ReconstituteExecution(
	id, intentID uuid.UUID,
	status ExecutionStatus,
	result *ExecutionResult,
	retryAttempt, currentStepNo int,
	execContext map[string]any,
	failureReason *string,
	startedAt, completedAt *time.Time,
	createdAt time.Time,
	steps []*ExecutionStep,
	snapshots []*ExecutionSnapshot,
	timeline []*TimelineEntry,
) *Execution {
	return &Execution{
		id:            id,
		intentID:      intentID,
		status:        status,
		result:        result,
		retryAttempt:  retryAttempt,
		currentStepNo: currentStepNo,
		context:       execContext,
		failureReason: failureReason,
		startedAt:     startedAt,
		completedAt:   completedAt,
		createdAt:     createdAt,
		steps:         steps,
		snapshots:     snapshots,
		timeline:      timeline,
	}
}

func (e *Execution) ID() uuid.UUID                   { return e.id }
func (e *Execution) IntentID() uuid.UUID             { return e.intentID }
func (e *Execution) Status() ExecutionStatus         { return e.status }
func (e *Execution) Result() *ExecutionResult        { return e.result }
func (e *Execution) RetryAttempt() int               { return e.retryAttempt }
func (e *Execution) CurrentStepNo() int              { return e.currentStepNo }
func (e *Execution) Context() map[string]any         { return e.context }
func (e *Execution) FailureReason() *string          { return e.failureReason }
func (e *Execution) StartedAt() *time.Time           { return e.startedAt }
func (e *Execution) CompletedAt() *time.Time         { return e.completedAt }
func (e *Execution) CreatedAt() time.Time            { return e.createdAt }
func (e *Execution) Steps() []*ExecutionStep         { return e.steps }
func (e *Execution) Snapshots() []*ExecutionSnapshot { return e.snapshots }
func (e *Execution) Timeline() []*TimelineEntry      { return e.timeline }

// MergeContext shallow-merges canonical (non-secret) fields into the Execution context.
func (e *Execution) MergeContext(patch map[string]any) {
	if e.context == nil {
		e.context = map[string]any{}
	}
	for k, v := range patch {
		e.context[k] = v
	}
}

// Queue transitions a freshly Created Execution to Queued for the worker to pick up.
func (e *Execution) Queue(now time.Time) error {
	if e.status != ExecutionStatusCreated {
		return ErrInvalidTransition
	}
	e.status = ExecutionStatusQueued
	e.record(EventExecutionQueued, e.id, now.UTC(), nil)
	return nil
}

// Start transitions Queued -> Running.
func (e *Execution) Start(now time.Time) error {
	if e.status != ExecutionStatusQueued {
		return ErrInvalidTransition
	}
	at := now.UTC()
	e.status = ExecutionStatusRunning
	if e.startedAt == nil {
		e.startedAt = &at
	}
	e.record(EventExecutionStarted, e.id, at, nil)
	return nil
}

// AllStepsSucceeded reports whether every Step reached Succeeded.
func (e *Execution) AllStepsSucceeded() bool {
	if len(e.steps) == 0 {
		return false
	}
	for _, s := range e.steps {
		if s.status != StepStatusSucceeded {
			return false
		}
	}
	return true
}

// Succeed transitions Running -> Succeeded. Requires every Step to have Succeeded.
func (e *Execution) Succeed(now time.Time) error {
	if e.status != ExecutionStatusRunning {
		return ErrInvalidTransition
	}
	if !e.AllStepsSucceeded() {
		return ErrStepsIncomplete
	}
	at := now.UTC()
	result := ExecutionResultSuccess
	e.status = ExecutionStatusSucceeded
	e.result = &result
	e.completedAt = &at
	e.failureReason = nil
	e.record(EventExecutionSucceeded, e.id, at, nil)
	return nil
}

// Fail transitions Running -> Failed. Failed is not terminal: ScheduleRetry or DeadLetter follow.
func (e *Execution) Fail(reason string, now time.Time) error {
	if e.status != ExecutionStatusRunning {
		return ErrInvalidTransition
	}
	result := ExecutionResultFailed
	e.status = ExecutionStatusFailed
	e.result = &result
	e.failureReason = &reason
	e.record(EventExecutionFailed, e.id, now.UTC(), map[string]any{"reason": reason})
	return nil
}

// Cancel transitions Created/Queued/Running -> Cancelled. Terminal.
func (e *Execution) Cancel(now time.Time) error {
	switch e.status {
	case ExecutionStatusCreated, ExecutionStatusQueued, ExecutionStatusRunning:
	default:
		return ErrInvalidTransition
	}
	at := now.UTC()
	result := ExecutionResultCancelled
	e.status = ExecutionStatusCancelled
	e.result = &result
	e.completedAt = &at
	e.record(EventExecutionCancelled, e.id, at, nil)
	return nil
}

// Expire transitions Created/Queued/Running -> Expired. Terminal.
func (e *Execution) Expire(now time.Time) error {
	switch e.status {
	case ExecutionStatusCreated, ExecutionStatusQueued, ExecutionStatusRunning:
	default:
		return ErrInvalidTransition
	}
	at := now.UTC()
	result := ExecutionResultExpired
	e.status = ExecutionStatusExpired
	e.result = &result
	e.completedAt = &at
	e.record(EventExecutionExpired, e.id, at, nil)
	return nil
}

// ScheduleRetry transitions Failed -> Queued, increments RetryAttempt, and resets Steps to
// Pending so the worker replays the pipeline from the beginning. Used for both automatic
// (RunExecution) and manual (RetryExecution use case) retries.
func (e *Execution) ScheduleRetry(now time.Time) error {
	if e.status != ExecutionStatusFailed {
		return ErrInvalidTransition
	}
	e.retryAttempt++
	e.currentStepNo = 0
	e.completedAt = nil
	for _, s := range e.steps {
		s.reset()
	}
	e.status = ExecutionStatusQueued
	e.record(EventExecutionRetryScheduled, e.id, now.UTC(), map[string]any{"retry_attempt": e.retryAttempt})
	return nil
}

// CanRetry reports whether an automatic retry is still allowed under maxRetries.
func (e *Execution) CanRetry(maxRetries int) bool {
	if maxRetries <= 0 {
		maxRetries = defaultMaxRetries
	}
	return e.retryAttempt < maxRetries
}

// DeadLetter transitions Failed -> DeadLetter. Terminal.
func (e *Execution) DeadLetter(now time.Time) error {
	if e.status != ExecutionStatusFailed {
		return ErrInvalidTransition
	}
	at := now.UTC()
	result := ExecutionResultDeadLetter
	e.status = ExecutionStatusDeadLetter
	e.result = &result
	e.completedAt = &at
	e.record(EventExecutionDeadLettered, e.id, at, nil)
	return nil
}

// NextPendingStep returns the earliest Pending Step in step_no order, if any.
func (e *Execution) NextPendingStep() (*ExecutionStep, bool) {
	for _, s := range e.steps {
		if s.status == StepStatusPending {
			return s, true
		}
	}
	return nil, false
}

// StartStep marks stepNo Running and advances CurrentStepNo.
func (e *Execution) StartStep(stepNo int, now time.Time) error {
	step, err := e.stepByNo(stepNo)
	if err != nil {
		return err
	}
	if err := step.start(now); err != nil {
		return err
	}
	e.currentStepNo = stepNo
	return nil
}

// SucceedStep marks stepNo Succeeded.
func (e *Execution) SucceedStep(stepNo int, now time.Time) error {
	step, err := e.stepByNo(stepNo)
	if err != nil {
		return err
	}
	return step.succeed(now)
}

// FailStep marks stepNo Failed with an error code/message.
func (e *Execution) FailStep(stepNo int, errorCode, errorMessage string, now time.Time) error {
	step, err := e.stepByNo(stepNo)
	if err != nil {
		return err
	}
	return step.fail(errorCode, errorMessage, now)
}

func (e *Execution) stepByNo(stepNo int) (*ExecutionStep, error) {
	for _, s := range e.steps {
		if s.stepNo == stepNo {
			return s, nil
		}
	}
	return nil, ErrStepNotFound
}

// AppendTimeline records a new, immutable Timeline entry. id is generated by the Application
// layer (UUID v7); Domain never generates ids.
func (e *Execution) AppendTimeline(id uuid.UUID, event, message string, metadata map[string]any, now time.Time) *TimelineEntry {
	entry := &TimelineEntry{
		id:          id,
		executionID: e.id,
		event:       event,
		message:     message,
		metadata:    metadata,
		createdAt:   now.UTC(),
	}
	e.timeline = append(e.timeline, entry)
	return entry
}

// AddSnapshot records a new, immutable request/response snapshot. id is generated by the
// Application layer (UUID v7); Domain never generates ids.
func (e *Execution) AddSnapshot(id uuid.UUID, stepID *uuid.UUID, snapshotType SnapshotType, snapshot map[string]any, contentType string, now time.Time) *ExecutionSnapshot {
	s := &ExecutionSnapshot{
		id:           id,
		executionID:  e.id,
		stepID:       stepID,
		snapshotType: snapshotType,
		snapshot:     snapshot,
		contentType:  contentType,
		createdAt:    now.UTC(),
	}
	e.snapshots = append(e.snapshots, s)
	return s
}
