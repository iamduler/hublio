package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func mustConnection(t *testing.T) *Connection {
	t.Helper()
	conn, err := NewConnection(
		uuid.Must(uuid.NewV7()),
		uuid.Must(uuid.NewV7()),
		uuid.Must(uuid.NewV7()),
		"My Connection",
		false,
		"desc",
		"production",
		map[string]any{"base_url": "https://api.example.com"},
		nil,
		30,
		time.Now(),
	)
	if err != nil {
		t.Fatalf("NewConnection() unexpected error: %v", err)
	}
	return conn
}

func TestNewConnection(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		connName    string
		environment string
		wantErr     error
	}{
		{"valid", "conn", "production", nil},
		{"empty name", "", "production", ErrInvalidName},
		{"empty environment", "conn", "", ErrInvalidEnvironment},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewConnection(uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), tt.connName, false, "", tt.environment, nil, nil, 0, now)
			if err != tt.wantErr {
				t.Fatalf("got err %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewConnection_DefaultsStatusToDraft(t *testing.T) {
	conn := mustConnection(t)
	if conn.Status() != ConnectionStatusDraft {
		t.Fatalf("expected draft status, got %v", conn.Status())
	}
	if conn.CanExecuteIntents() {
		t.Fatalf("draft connection must not execute intents")
	}
}

func TestConnection_StatusTransitions(t *testing.T) {
	tests := []struct {
		name      string
		fromState ConnectionStatus
		action    func(*Connection) error
		wantErr   error
		wantState ConnectionStatus
	}{
		{"draft to verifying", ConnectionStatusDraft, func(c *Connection) error { return c.StartVerify(time.Now()) }, nil, ConnectionStatusVerifying},
		{"verifying to active", ConnectionStatusVerifying, func(c *Connection) error { return c.MarkVerified(time.Now()) }, nil, ConnectionStatusActive},
		{"verifying to verification_failed", ConnectionStatusVerifying, func(c *Connection) error { return c.MarkVerificationFailed("timeout", time.Now()) }, nil, ConnectionStatusVerificationFailed},
		{"verification_failed to verifying", ConnectionStatusVerificationFailed, func(c *Connection) error { return c.StartVerify(time.Now()) }, nil, ConnectionStatusVerifying},
		{"active to disabled", ConnectionStatusActive, func(c *Connection) error { return c.Disable(time.Now()) }, nil, ConnectionStatusDisabled},
		{"disabled to active", ConnectionStatusDisabled, func(c *Connection) error { return c.Enable(time.Now()) }, nil, ConnectionStatusActive},
		{"draft cannot mark verified", ConnectionStatusDraft, func(c *Connection) error { return c.MarkVerified(time.Now()) }, ErrInvalidTransition, ConnectionStatusDraft},
		{"active cannot start verify", ConnectionStatusActive, func(c *Connection) error { return c.StartVerify(time.Now()) }, ErrInvalidTransition, ConnectionStatusActive},
		{"draft cannot disable", ConnectionStatusDraft, func(c *Connection) error { return c.Disable(time.Now()) }, ErrInvalidTransition, ConnectionStatusDraft},
		{"draft cannot enable", ConnectionStatusDraft, func(c *Connection) error { return c.Enable(time.Now()) }, ErrInvalidTransition, ConnectionStatusDraft},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := mustConnection(t)
			conn.status = tt.fromState
			err := tt.action(conn)
			if err != tt.wantErr {
				t.Fatalf("got err %v, want %v", err, tt.wantErr)
			}
			if conn.Status() != tt.wantState {
				t.Fatalf("got status %v, want %v", conn.Status(), tt.wantState)
			}
		})
	}
}

func TestConnection_CanExecuteIntents(t *testing.T) {
	conn := mustConnection(t)
	conn.status = ConnectionStatusActive
	if !conn.CanExecuteIntents() {
		t.Fatalf("active connection should execute intents")
	}
	conn.deletedAt = timePtr(time.Now())
	if conn.CanExecuteIntents() {
		t.Fatalf("soft-deleted connection must not execute intents")
	}
}

func TestConnection_SetActiveCredential(t *testing.T) {
	conn := mustConnection(t)
	credID := uuid.Must(uuid.NewV7())
	conn.SetActiveCredential(credID, time.Now())
	if conn.ActiveCredentialID() == nil || *conn.ActiveCredentialID() != credID {
		t.Fatalf("expected active credential id %v, got %v", credID, conn.ActiveCredentialID())
	}
}

func TestConnection_PullEvents(t *testing.T) {
	conn := mustConnection(t)
	events := conn.PullEvents()
	if len(events) != 1 || events[0].Name != EventConnectionCreated {
		t.Fatalf("expected 1 ConnectionCreated event, got %+v", events)
	}
}

func timePtr(t time.Time) *time.Time { return &t }
