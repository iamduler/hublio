package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// MFAConfig holds a user's TOTP second factor (Identity configuration entity, like
// OAuthIdentity — not a Runtime Aggregate).
//
// The Domain never sees the plaintext TOTP secret nor a plaintext recovery code: the secret
// arrives already encrypted and recovery codes already hashed, so all cryptography stays in
// Infrastructure while the enrollment rules stay here.
type MFAConfig struct {
	userID              uuid.UUID
	totpSecretEncrypted string
	enabledAt           *time.Time
	recoveryCodeHashes  []string
	createdAt           time.Time
	updatedAt           time.Time
}

// NewMFAConfig starts an enrollment: the secret is stored but MFA stays off until Enable.
func NewMFAConfig(
	userID uuid.UUID,
	totpSecretEncrypted string,
	recoveryCodeHashes []string,
	now time.Time,
) (*MFAConfig, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidMFAConfig
	}
	if strings.TrimSpace(totpSecretEncrypted) == "" {
		return nil, ErrInvalidMFAConfig
	}
	if len(recoveryCodeHashes) == 0 {
		return nil, ErrInvalidMFAConfig
	}
	at := now.UTC()
	return &MFAConfig{
		userID:              userID,
		totpSecretEncrypted: totpSecretEncrypted,
		recoveryCodeHashes:  cloneHashes(recoveryCodeHashes),
		createdAt:           at,
		updatedAt:           at,
	}, nil
}

func ReconstituteMFAConfig(
	userID uuid.UUID,
	totpSecretEncrypted string,
	enabledAt *time.Time,
	recoveryCodeHashes []string,
	createdAt, updatedAt time.Time,
) *MFAConfig {
	return &MFAConfig{
		userID:              userID,
		totpSecretEncrypted: totpSecretEncrypted,
		enabledAt:           enabledAt,
		recoveryCodeHashes:  cloneHashes(recoveryCodeHashes),
		createdAt:           createdAt,
		updatedAt:           updatedAt,
	}
}

func (m *MFAConfig) UserID() uuid.UUID            { return m.userID }
func (m *MFAConfig) TOTPSecretEncrypted() string  { return m.totpSecretEncrypted }
func (m *MFAConfig) EnabledAt() *time.Time        { return m.enabledAt }
func (m *MFAConfig) CreatedAt() time.Time         { return m.createdAt }
func (m *MFAConfig) UpdatedAt() time.Time         { return m.updatedAt }
func (m *MFAConfig) RecoveryCodeHashes() []string { return cloneHashes(m.recoveryCodeHashes) }
func (m *MFAConfig) RemainingRecoveryCodes() int  { return len(m.recoveryCodeHashes) }
func (m *MFAConfig) Enabled() bool                { return m.enabledAt != nil }

// Enable turns the second factor on once the user proved possession of the secret.
func (m *MFAConfig) Enable(now time.Time) error {
	if m.Enabled() {
		return ErrMFAAlreadyEnabled
	}
	at := now.UTC()
	m.enabledAt = &at
	m.updatedAt = at
	return nil
}

// Reenroll replaces a pending enrollment with a freshly generated secret and recovery codes.
// An enabled config must be disabled first, so a stolen session cannot silently swap factors.
func (m *MFAConfig) Reenroll(totpSecretEncrypted string, recoveryCodeHashes []string, now time.Time) error {
	if m.Enabled() {
		return ErrMFAAlreadyEnabled
	}
	if strings.TrimSpace(totpSecretEncrypted) == "" || len(recoveryCodeHashes) == 0 {
		return ErrInvalidMFAConfig
	}
	at := now.UTC()
	m.totpSecretEncrypted = totpSecretEncrypted
	m.recoveryCodeHashes = cloneHashes(recoveryCodeHashes)
	m.updatedAt = at
	return nil
}

// ConsumeRecoveryCode burns a single-use recovery code, identified by the hash the caller
// matched. Recovery codes only exist as a fallback for an enabled factor.
func (m *MFAConfig) ConsumeRecoveryCode(hash string, now time.Time) error {
	if !m.Enabled() {
		return ErrMFANotEnabled
	}
	for i, h := range m.recoveryCodeHashes {
		if h != hash {
			continue
		}
		m.recoveryCodeHashes = append(m.recoveryCodeHashes[:i:i], m.recoveryCodeHashes[i+1:]...)
		m.updatedAt = now.UTC()
		return nil
	}
	return ErrInvalidMFACode
}

func cloneHashes(in []string) []string {
	out := make([]string, len(in))
	copy(out, in)
	return out
}
