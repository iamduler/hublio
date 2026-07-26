package domain_test

import (
	"errors"
	"testing"
	"time"

	"hublio/internal/identity/domain"

	"github.com/google/uuid"
)

func newPendingMFA(t *testing.T, now time.Time) *domain.MFAConfig {
	t.Helper()
	cfg, err := domain.NewMFAConfig(uuid.Must(uuid.NewV7()), "cipher", []string{"hash-a", "hash-b"}, now)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestNewMFAConfig_Validation(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name    string
		userID  uuid.UUID
		secret  string
		hashes  []string
		wantErr error
	}{
		{name: "valid enrollment", userID: uuid.Must(uuid.NewV7()), secret: "cipher", hashes: []string{"hash-a"}},
		{name: "missing user", userID: uuid.Nil, secret: "cipher", hashes: []string{"hash-a"}, wantErr: domain.ErrInvalidMFAConfig},
		{name: "missing secret", userID: uuid.Must(uuid.NewV7()), secret: "  ", hashes: []string{"hash-a"}, wantErr: domain.ErrInvalidMFAConfig},
		{name: "no recovery codes", userID: uuid.Must(uuid.NewV7()), secret: "cipher", wantErr: domain.ErrInvalidMFAConfig},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := domain.NewMFAConfig(tc.userID, tc.secret, tc.hashes, now)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("error = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Enabled() {
				t.Fatal("a new enrollment must start disabled")
			}
		})
	}
}

func TestMFAConfig_EnableIsOneWay(t *testing.T) {
	now := time.Now().UTC()
	cfg := newPendingMFA(t, now)

	if err := cfg.Enable(now); err != nil {
		t.Fatalf("enable: %v", err)
	}
	if !cfg.Enabled() || cfg.EnabledAt() == nil {
		t.Fatal("expected mfa to be enabled")
	}
	if err := cfg.Enable(now); !errors.Is(err, domain.ErrMFAAlreadyEnabled) {
		t.Fatalf("second enable = %v, want ErrMFAAlreadyEnabled", err)
	}
	if err := cfg.Reenroll("other", []string{"hash-c"}, now); !errors.Is(err, domain.ErrMFAAlreadyEnabled) {
		t.Fatalf("reenroll while enabled = %v, want ErrMFAAlreadyEnabled", err)
	}
}

func TestMFAConfig_ConsumeRecoveryCode(t *testing.T) {
	now := time.Now().UTC()

	pending := newPendingMFA(t, now)
	if err := pending.ConsumeRecoveryCode("hash-a", now); !errors.Is(err, domain.ErrMFANotEnabled) {
		t.Fatalf("consume while pending = %v, want ErrMFANotEnabled", err)
	}

	cfg := newPendingMFA(t, now)
	if err := cfg.Enable(now); err != nil {
		t.Fatal(err)
	}
	if err := cfg.ConsumeRecoveryCode("hash-a", now); err != nil {
		t.Fatalf("consume: %v", err)
	}
	if got := cfg.RemainingRecoveryCodes(); got != 1 {
		t.Fatalf("remaining = %d, want 1", got)
	}
	if err := cfg.ConsumeRecoveryCode("hash-a", now); !errors.Is(err, domain.ErrInvalidMFACode) {
		t.Fatalf("reuse = %v, want ErrInvalidMFACode", err)
	}
}

func TestMFAConfig_RecoveryCodeHashesAreCopied(t *testing.T) {
	now := time.Now().UTC()
	cfg := newPendingMFA(t, now)

	hashes := cfg.RecoveryCodeHashes()
	hashes[0] = "tampered"

	if cfg.RecoveryCodeHashes()[0] == "tampered" {
		t.Fatal("RecoveryCodeHashes must return a copy")
	}
}
