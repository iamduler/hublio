package application

import (
	"context"
	"errors"
	"testing"

	"hublio/internal/events/domain"
	"hublio/internal/platform/metrics"

	"github.com/google/uuid"
)

type fakeAuditRepository struct {
	saved   []*domain.AuditEntry
	saveErr error
}

func (f *fakeAuditRepository) Save(ctx context.Context, entry *domain.AuditEntry) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = append(f.saved, entry)
	return nil
}

func TestRecord_PersistsViaAuditRepository(t *testing.T) {
	repo := &fakeAuditRepository{}
	counters := metrics.New()
	svc := &Services{Audit: repo, Metrics: counters}

	actorID := uuid.Must(uuid.NewV7())
	resourceID := uuid.Must(uuid.NewV7())
	err := svc.Record(context.Background(), AuditRecord{
		ActorType:    "user",
		ActorID:      &actorID,
		Action:       "api_key.create",
		ResourceType: "api_key",
		ResourceID:   &resourceID,
	})
	if err != nil {
		t.Fatalf("Record() unexpected error: %v", err)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("expected 1 saved audit entry, got %d", len(repo.saved))
	}
	if repo.saved[0].Action() != "api_key.create" {
		t.Fatalf("Action() = %q", repo.saved[0].Action())
	}
	if counters.Snapshot().AuditRecords != 1 {
		t.Fatalf("audit_records_total = %d, want 1", counters.Snapshot().AuditRecords)
	}
}

func TestRecord_RedactsSecretLikeMetadataKeys(t *testing.T) {
	repo := &fakeAuditRepository{}
	svc := &Services{Audit: repo}

	err := svc.Record(context.Background(), AuditRecord{
		ActorType:    "system",
		Action:       "credential.rotate",
		ResourceType: "credential",
		Metadata: map[string]any{
			"connection_id":    "keep-me",
			"secret":           "must-not-be-stored",
			"plaintext_secret": "must-not-be-stored",
			"password":         "must-not-be-stored",
			"token":            "must-not-be-stored",
			"key_hash":         "must-not-be-stored",
		},
	})
	if err != nil {
		t.Fatalf("Record() unexpected error: %v", err)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("expected 1 saved audit entry, got %d", len(repo.saved))
	}

	metadata := repo.saved[0].Metadata()
	if metadata["connection_id"] != "keep-me" {
		t.Fatalf("expected non-secret metadata to be preserved, got %v", metadata)
	}
	for _, redactedKey := range []string{"secret", "plaintext_secret", "password", "token", "key_hash"} {
		if _, present := metadata[redactedKey]; present {
			t.Fatalf("expected %q to be redacted, got %v", redactedKey, metadata)
		}
	}
}

func TestRecord_RepositoryFailureIsSurfaced(t *testing.T) {
	repo := &fakeAuditRepository{saveErr: errors.New("db down")}
	svc := &Services{Audit: repo}

	err := svc.Record(context.Background(), AuditRecord{
		ActorType:    "user",
		Action:       "user.login",
		ResourceType: "user",
	})
	if err == nil {
		t.Fatal("expected Record() to surface the repository error")
	}
}

func TestRecord_InvalidInputRejected(t *testing.T) {
	repo := &fakeAuditRepository{}
	svc := &Services{Audit: repo}

	err := svc.Record(context.Background(), AuditRecord{
		ActorType:    "bogus", // invalid: NewAuditEntry requires a known ActorType
		Action:       "user.login",
		ResourceType: "user",
	})
	if err == nil {
		t.Fatal("expected Record() to reject an invalid AuditRecord")
	}
	if len(repo.saved) != 0 {
		t.Fatalf("expected nothing persisted on validation failure, got %d", len(repo.saved))
	}
}
