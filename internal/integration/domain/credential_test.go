package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func mustCredential(t *testing.T) *Credential {
	t.Helper()
	cred, err := NewCredential(
		uuid.Must(uuid.NewV7()),
		uuid.Must(uuid.NewV7()),
		CredentialTypeAPIKey,
		1,
		[]byte("ciphertext"),
		nil,
		uuid.Must(uuid.NewV7()),
		time.Now(),
	)
	if err != nil {
		t.Fatalf("NewCredential() unexpected error: %v", err)
	}
	return cred
}

func TestNewCredential(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		credType CredentialType
		secret   []byte
		wantErr  error
	}{
		{"valid", CredentialTypeAPIKey, []byte("cipher"), nil},
		{"invalid type", CredentialType("bogus"), []byte("cipher"), ErrInvalidCredentialType},
		{"empty secret", CredentialTypeAPIKey, nil, ErrEmptySecret},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCredential(uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), tt.credType, 1, tt.secret, nil, uuid.Must(uuid.NewV7()), now)
			if err != tt.wantErr {
				t.Fatalf("got err %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestCredential_IsUsable(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		status    CredentialStatus
		expiresAt *time.Time
		wantErr   error
	}{
		{"active no expiry", CredentialStatusActive, nil, nil},
		{"active not yet expired", CredentialStatusActive, timePtr(now.Add(time.Hour)), nil},
		{"active expired", CredentialStatusActive, timePtr(now.Add(-time.Hour)), ErrCredentialNotActive},
		{"revoked", CredentialStatusRevoked, nil, ErrCredentialNotActive},
		{"expired status", CredentialStatusExpired, nil, ErrCredentialNotActive},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred := mustCredential(t)
			cred.status = tt.status
			cred.expiresAt = tt.expiresAt
			if err := cred.IsUsable(now); err != tt.wantErr {
				t.Fatalf("got err %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestCredential_Revoke(t *testing.T) {
	cred := mustCredential(t)
	if err := cred.Revoke(time.Now()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cred.Status() != CredentialStatusRevoked {
		t.Fatalf("expected revoked status, got %v", cred.Status())
	}
	if cred.RotatedAt() == nil {
		t.Fatalf("expected rotated_at to be set")
	}
	if err := cred.Revoke(time.Now()); err != ErrInvalidTransition {
		t.Fatalf("expected ErrInvalidTransition revoking twice, got %v", err)
	}
}

func TestCredential_MarkExpired(t *testing.T) {
	cred := mustCredential(t)
	if err := cred.MarkExpired(time.Now()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cred.Status() != CredentialStatusExpired {
		t.Fatalf("expected expired status, got %v", cred.Status())
	}
	if err := cred.MarkExpired(time.Now()); err != ErrInvalidTransition {
		t.Fatalf("expected ErrInvalidTransition marking expired twice, got %v", err)
	}
}

func TestRotateCredential_NoPrevious(t *testing.T) {
	connID := uuid.Must(uuid.NewV7())
	createdBy := uuid.Must(uuid.NewV7())
	next, err := RotateCredential(uuid.Must(uuid.NewV7()), nil, connID, CredentialTypeAPIKey, []byte("cipher"), nil, createdBy, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if next.Version() != 1 {
		t.Fatalf("expected version 1, got %d", next.Version())
	}
	events := next.PullEvents()
	if len(events) != 2 || events[0].Name != EventCredentialCreated || events[1].Name != EventCredentialRotated {
		t.Fatalf("unexpected events: %+v", events)
	}
}

func TestRotateCredential_WithPrevious(t *testing.T) {
	previous := mustCredential(t)
	previous.version = 3
	previous.PullEvents()

	connID := previous.ConnectionID()
	createdBy := uuid.Must(uuid.NewV7())
	next, err := RotateCredential(uuid.Must(uuid.NewV7()), previous, connID, CredentialTypeAPIKey, []byte("new-cipher"), nil, createdBy, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if next.Version() != 4 {
		t.Fatalf("expected version 4, got %d", next.Version())
	}
	if previous.Status() != CredentialStatusRevoked {
		t.Fatalf("expected previous credential revoked, got %v", previous.Status())
	}
	prevEvents := previous.PullEvents()
	if len(prevEvents) != 1 || prevEvents[0].Name != EventCredentialRevoked {
		t.Fatalf("expected previous CredentialRevoked event, got %+v", prevEvents)
	}
}

func TestRotateCredential_PreviousAlreadyRevoked(t *testing.T) {
	previous := mustCredential(t)
	if err := previous.Revoke(time.Now()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := RotateCredential(uuid.Must(uuid.NewV7()), previous, previous.ConnectionID(), CredentialTypeAPIKey, []byte("cipher"), nil, uuid.Must(uuid.NewV7()), time.Now())
	if err != ErrInvalidTransition {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestCredential_PullEvents(t *testing.T) {
	cred := mustCredential(t)
	events := cred.PullEvents()
	if len(events) != 1 || events[0].Name != EventCredentialCreated {
		t.Fatalf("expected 1 CredentialCreated event, got %+v", events)
	}
}
