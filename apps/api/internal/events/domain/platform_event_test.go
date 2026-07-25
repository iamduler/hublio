package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewPlatformEvent(t *testing.T) {
	now := time.Now()
	validID := uuid.Must(uuid.NewV7())
	validAggregateID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name          string
		id            uuid.UUID
		aggregateType AggregateType
		aggregateID   uuid.UUID
		category      Category
		eventName     string
		wantErr       error
	}{
		{"valid", validID, AggregateTypeExecution, validAggregateID, CategoryRuntime, "ExecutionSucceeded", nil},
		{"nil id", uuid.Nil, AggregateTypeExecution, validAggregateID, CategoryRuntime, "ExecutionSucceeded", ErrInvalidID},
		{"nil aggregate id", validID, AggregateTypeExecution, uuid.Nil, CategoryRuntime, "ExecutionSucceeded", ErrInvalidID},
		{"invalid aggregate type", validID, AggregateType("bogus"), validAggregateID, CategoryRuntime, "ExecutionSucceeded", ErrInvalidAggregateType},
		{"invalid category", validID, AggregateTypeExecution, validAggregateID, Category("bogus"), "ExecutionSucceeded", ErrInvalidCategory},
		{"empty event name", validID, AggregateTypeExecution, validAggregateID, CategoryRuntime, "", ErrInvalidEventName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := NewPlatformEvent(
				tt.id, nil, nil, tt.aggregateType, tt.aggregateID, nil, tt.category, tt.eventName, "",
				nil, nil, "test", now,
			)
			if err != tt.wantErr {
				t.Fatalf("got err %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && event == nil {
				t.Fatal("expected non-nil event on success")
			}
		})
	}
}

func TestNewPlatformEvent_EventNameTooLong(t *testing.T) {
	longName := make([]byte, 151)
	for i := range longName {
		longName[i] = 'a'
	}
	_, err := NewPlatformEvent(
		uuid.Must(uuid.NewV7()), nil, nil, AggregateTypeExecution, uuid.Must(uuid.NewV7()), nil,
		CategoryRuntime, string(longName), "", nil, nil, "test", time.Now(),
	)
	if err != ErrInvalidEventName {
		t.Fatalf("got err %v, want %v", err, ErrInvalidEventName)
	}
}

func TestNewPlatformEvent_PreservesFields(t *testing.T) {
	id := uuid.Must(uuid.NewV7())
	orgID := uuid.Must(uuid.NewV7())
	wsID := uuid.Must(uuid.NewV7())
	execID := uuid.Must(uuid.NewV7())
	aggregateID := uuid.Must(uuid.NewV7())
	now := time.Now()
	payload := map[string]any{"reason": "boom"}
	metadata := map[string]any{"step": 3}

	event, err := NewPlatformEvent(
		id, &orgID, &wsID, AggregateTypeExecution, aggregateID, &execID,
		CategoryRuntime, "ExecutionFailed", "corr-123", payload, metadata, "orchestration", now,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.ID() != id {
		t.Fatalf("ID() = %v, want %v", event.ID(), id)
	}
	if event.OrganizationID() == nil || *event.OrganizationID() != orgID {
		t.Fatalf("OrganizationID() = %v, want %v", event.OrganizationID(), orgID)
	}
	if event.WorkspaceID() == nil || *event.WorkspaceID() != wsID {
		t.Fatalf("WorkspaceID() = %v, want %v", event.WorkspaceID(), wsID)
	}
	if event.ExecutionID() == nil || *event.ExecutionID() != execID {
		t.Fatalf("ExecutionID() = %v, want %v", event.ExecutionID(), execID)
	}
	if event.AggregateType() != AggregateTypeExecution {
		t.Fatalf("AggregateType() = %v, want %v", event.AggregateType(), AggregateTypeExecution)
	}
	if event.Category() != CategoryRuntime {
		t.Fatalf("Category() = %v, want %v", event.Category(), CategoryRuntime)
	}
	if event.EventName() != "ExecutionFailed" {
		t.Fatalf("EventName() = %q, want %q", event.EventName(), "ExecutionFailed")
	}
	if event.CorrelationID() != "corr-123" {
		t.Fatalf("CorrelationID() = %q, want %q", event.CorrelationID(), "corr-123")
	}
	if event.Payload()["reason"] != "boom" {
		t.Fatalf("Payload() = %v", event.Payload())
	}
	if event.Metadata()["step"] != 3 {
		t.Fatalf("Metadata() = %v", event.Metadata())
	}
	if event.PublishedBy() != "orchestration" {
		t.Fatalf("PublishedBy() = %q", event.PublishedBy())
	}
	if !event.CreatedAt().Equal(now.UTC()) {
		t.Fatalf("CreatedAt() = %v, want %v", event.CreatedAt(), now.UTC())
	}
}

func TestReconstitutePlatformEvent_SkipsValidation(t *testing.T) {
	// Reconstitution must not re-validate: rows coming from persistence are already
	// known-good, and validation errors would make hydration impossible for legacy rows.
	event := ReconstitutePlatformEvent(
		uuid.Must(uuid.NewV7()), nil, nil, AggregateType("legacy"), uuid.Must(uuid.NewV7()), nil,
		Category("legacy"), "", "", nil, nil, "", time.Now(),
	)
	if event == nil {
		t.Fatal("expected non-nil event")
	}
}
