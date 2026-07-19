package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewAuditEntry(t *testing.T) {
	now := time.Now()
	validID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name         string
		id           uuid.UUID
		actorType    ActorType
		action       string
		resourceType string
		wantErr      error
	}{
		{"valid", validID, ActorTypeUser, "api_key.create", "api_key", nil},
		{"nil id", uuid.Nil, ActorTypeUser, "api_key.create", "api_key", ErrInvalidID},
		{"invalid actor type", validID, ActorType("bogus"), "api_key.create", "api_key", ErrInvalidActorType},
		{"empty action", validID, ActorTypeUser, "", "api_key", ErrInvalidAction},
		{"empty resource type", validID, ActorTypeUser, "api_key.create", "", ErrInvalidResourceType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := NewAuditEntry(
				tt.id, nil, nil, tt.actorType, nil, tt.action, tt.resourceType, nil,
				"", "", "", "", nil, now,
			)
			if err != tt.wantErr {
				t.Fatalf("got err %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && entry == nil {
				t.Fatal("expected non-nil entry on success")
			}
		})
	}
}

func TestNewAuditEntry_PreservesFieldsAndNeverStoresRawSecrets(t *testing.T) {
	id := uuid.Must(uuid.NewV7())
	orgID := uuid.Must(uuid.NewV7())
	wsID := uuid.Must(uuid.NewV7())
	actorID := uuid.Must(uuid.NewV7())
	resourceID := uuid.Must(uuid.NewV7())
	now := time.Now()

	// Domain never redacts (that's an Application concern, see events/application.redactMetadata);
	// this test only asserts the entity preserves whatever Metadata it is given verbatim.
	metadata := map[string]any{"name": "ci-key"}

	entry, err := NewAuditEntry(
		id, &orgID, &wsID, ActorTypeAPIKey, &actorID, "connection.create", "connection", &resourceID,
		"req-1", "corr-1", "127.0.0.1", "curl/8.0", metadata, now,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.ID() != id {
		t.Fatalf("ID() = %v, want %v", entry.ID(), id)
	}
	if entry.OrganizationID() == nil || *entry.OrganizationID() != orgID {
		t.Fatalf("OrganizationID() = %v, want %v", entry.OrganizationID(), orgID)
	}
	if entry.WorkspaceID() == nil || *entry.WorkspaceID() != wsID {
		t.Fatalf("WorkspaceID() = %v, want %v", entry.WorkspaceID(), wsID)
	}
	if entry.ActorType() != ActorTypeAPIKey {
		t.Fatalf("ActorType() = %v, want %v", entry.ActorType(), ActorTypeAPIKey)
	}
	if entry.ActorID() == nil || *entry.ActorID() != actorID {
		t.Fatalf("ActorID() = %v, want %v", entry.ActorID(), actorID)
	}
	if entry.Action() != "connection.create" {
		t.Fatalf("Action() = %q", entry.Action())
	}
	if entry.ResourceType() != "connection" {
		t.Fatalf("ResourceType() = %q", entry.ResourceType())
	}
	if entry.ResourceID() == nil || *entry.ResourceID() != resourceID {
		t.Fatalf("ResourceID() = %v, want %v", entry.ResourceID(), resourceID)
	}
	if entry.RequestID() != "req-1" || entry.CorrelationID() != "corr-1" {
		t.Fatalf("RequestID/CorrelationID = %q/%q", entry.RequestID(), entry.CorrelationID())
	}
	if entry.IP() != "127.0.0.1" || entry.UserAgent() != "curl/8.0" {
		t.Fatalf("IP/UserAgent = %q/%q", entry.IP(), entry.UserAgent())
	}
	if entry.Metadata()["name"] != "ci-key" {
		t.Fatalf("Metadata() = %v", entry.Metadata())
	}
	if !entry.CreatedAt().Equal(now.UTC()) {
		t.Fatalf("CreatedAt() = %v, want %v", entry.CreatedAt(), now.UTC())
	}
}

func TestNewAuditEntry_ActionTooLong(t *testing.T) {
	longAction := make([]byte, 151)
	for i := range longAction {
		longAction[i] = 'a'
	}
	_, err := NewAuditEntry(
		uuid.Must(uuid.NewV7()), nil, nil, ActorTypeSystem, nil, string(longAction), "resource", nil,
		"", "", "", "", nil, time.Now(),
	)
	if err != ErrInvalidAction {
		t.Fatalf("got err %v, want %v", err, ErrInvalidAction)
	}
}

func TestNewAuditEntry_ResourceTypeTooLong(t *testing.T) {
	longType := make([]byte, 101)
	for i := range longType {
		longType[i] = 'a'
	}
	_, err := NewAuditEntry(
		uuid.Must(uuid.NewV7()), nil, nil, ActorTypeSystem, nil, "action", string(longType), nil,
		"", "", "", "", nil, time.Now(),
	)
	if err != ErrInvalidResourceType {
		t.Fatalf("got err %v, want %v", err, ErrInvalidResourceType)
	}
}
