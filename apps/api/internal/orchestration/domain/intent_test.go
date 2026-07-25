package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func newTestIntent(t *testing.T) *Intent {
	t.Helper()
	now := time.Now()
	intent, err := NewIntent(uuid.New(), uuid.New(), uuid.New(), uuid.New(), "invoice.create", map[string]any{"amount": 100}, "corr-1", "idem-1", now)
	if err != nil {
		t.Fatalf("NewIntent() unexpected error: %v", err)
	}
	return intent
}

func TestNewIntent(t *testing.T) {
	now := time.Now()
	orgID, wsID, connID := uuid.New(), uuid.New(), uuid.New()

	tests := []struct {
		name       string
		id         uuid.UUID
		orgID      uuid.UUID
		wsID       uuid.UUID
		connID     uuid.UUID
		capability string
		payload    map[string]any
		wantErr    error
	}{
		{
			name:       "valid",
			id:         uuid.New(),
			orgID:      orgID,
			wsID:       wsID,
			connID:     connID,
			capability: "invoice.create",
			payload:    map[string]any{"amount": 1},
		},
		{
			name:    "nil id rejected",
			id:      uuid.Nil,
			orgID:   orgID,
			wsID:    wsID,
			connID:  connID,
			wantErr: ErrInvalidID,
		},
		{
			name:    "nil organization id rejected",
			id:      uuid.New(),
			orgID:   uuid.Nil,
			wsID:    wsID,
			connID:  connID,
			wantErr: ErrInvalidID,
		},
		{
			name:    "nil workspace id rejected",
			id:      uuid.New(),
			orgID:   orgID,
			wsID:    uuid.Nil,
			connID:  connID,
			wantErr: ErrInvalidID,
		},
		{
			name:    "nil connection id rejected",
			id:      uuid.New(),
			orgID:   orgID,
			wsID:    wsID,
			connID:  uuid.Nil,
			wantErr: ErrInvalidID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intent, err := NewIntent(tt.id, tt.orgID, tt.wsID, tt.connID, tt.capability, tt.payload, "", "", now)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if intent.Status() != IntentStatusSubmitted {
				t.Fatalf("expected status submitted, got %s", intent.Status())
			}
			events := intent.PullEvents()
			if len(events) != 1 || events[0].Name != EventIntentSubmitted {
				t.Fatalf("expected one IntentSubmitted event, got %+v", events)
			}
		})
	}
}

func TestIntent_IsValid(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		capability string
		payload    map[string]any
		want       bool
	}{
		{name: "capability and payload present", capability: "invoice.create", payload: map[string]any{"a": 1}, want: true},
		{name: "empty capability", capability: "", payload: map[string]any{"a": 1}, want: false},
		{name: "empty payload", capability: "invoice.create", payload: map[string]any{}, want: false},
		{name: "nil payload", capability: "invoice.create", payload: nil, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intent, err := NewIntent(uuid.New(), uuid.New(), uuid.New(), uuid.New(), tt.capability, tt.payload, "", "", now)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := intent.IsValid(); got != tt.want {
				t.Fatalf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntent_Accept(t *testing.T) {
	tests := []struct {
		name          string
		mutate        func(*Intent)
		wantErr       error
		wantStatus    IntentStatus
		wantImmutable bool
	}{
		{
			name:          "submitted can accept",
			mutate:        func(i *Intent) {},
			wantStatus:    IntentStatusAccepted,
			wantImmutable: true,
		},
		{
			name: "accepted cannot accept again",
			mutate: func(i *Intent) {
				_ = i.Accept(time.Now())
			},
			wantErr:       ErrInvalidTransition,
			wantStatus:    IntentStatusAccepted,
			wantImmutable: true,
		},
		{
			name: "rejected cannot accept",
			mutate: func(i *Intent) {
				_ = i.Reject("bad", time.Now())
			},
			wantErr:    ErrInvalidTransition,
			wantStatus: IntentStatusRejected,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intent := newTestIntent(t)
			tt.mutate(intent)
			err := intent.Accept(time.Now())
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if intent.Status() != tt.wantStatus {
				t.Fatalf("expected status %s, got %s", tt.wantStatus, intent.Status())
			}
			if intent.IsImmutable() != tt.wantImmutable {
				t.Fatalf("expected IsImmutable() = %v, got %v", tt.wantImmutable, intent.IsImmutable())
			}
		})
	}
}

func TestIntent_Reject(t *testing.T) {
	intent := newTestIntent(t)
	if err := intent.Reject("capability unknown", time.Now()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent.Status() != IntentStatusRejected {
		t.Fatalf("expected status rejected, got %s", intent.Status())
	}
	if err := intent.Reject("again", time.Now()); err != ErrInvalidTransition {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}
	if err := intent.Accept(time.Now()); err != ErrInvalidTransition {
		t.Fatalf("expected ErrInvalidTransition on accept-after-reject, got %v", err)
	}
}

func TestIntent_Expire(t *testing.T) {
	intent := newTestIntent(t)
	if err := intent.Expire(time.Now()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent.Status() != IntentStatusExpired {
		t.Fatalf("expected status expired, got %s", intent.Status())
	}
	if err := intent.Accept(time.Now()); err != ErrInvalidTransition {
		t.Fatalf("expected ErrInvalidTransition on accept-after-expire, got %v", err)
	}
}
