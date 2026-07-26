package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"hublio/internal/identity/domain"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/auth"

	"github.com/google/uuid"
)

const (
	mfaChallengeTTL      = 5 * time.Minute
	mfaTrustTTL          = 30 * 24 * time.Hour
	mfaChallengeAttempts = 5
	mfaRecoveryCodeCount = 10
)

// TOTPSecret is a freshly provisioned shared secret plus the otpauth:// URL an authenticator
// app can consume as a QR code.
type TOTPSecret struct {
	Secret     string
	OTPAuthURL string
}

// TOTPVerifier provisions and validates time-based one-time passwords. Infrastructure owns the
// algorithm so the Application layer never imports a crypto/OTP library directly.
type TOTPVerifier interface {
	Generate(accountName string) (TOTPSecret, error)
	Verify(secret, code string, at time.Time) bool
}

// MFASecretCipher protects the TOTP shared secret at rest (AES-GCM in Infrastructure).
type MFASecretCipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// mfaChallengePayload is the short-lived Redis state bridging password login and code entry.
// ExpiresAt is stored explicitly so a failed attempt can be re-persisted without extending the
// original 5 minute window.
type mfaChallengePayload struct {
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	Attempts  int       `json:"attempts"`
}

// MFASetupResult carries the only copy of the plaintext secret and recovery codes the user will
// ever receive. Never log or persist these values.
type MFASetupResult struct {
	Secret        string
	OTPAuthURL    string
	RecoveryCodes []string
}

// MFAStatus is the enrollment state shown in Settings → Security.
type MFAStatus struct {
	Enabled                 bool
	PendingEnrollment       bool
	RemainingRecoveryCodes  int
	CanEnroll               bool
}

// GetMFAStatus returns whether MFA is enabled/pending for the current user.
// Never exposes secrets or recovery codes.
func (s *Services) GetMFAStatus(ctx context.Context, userID uuid.UUID) (*MFAStatus, error) {
	user, err := s.Users.FindByID(ctx, userID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	out := &MFAStatus{CanEnroll: user.HasPassword()}
	cfg, err := s.MFA.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return out, nil
		}
		return nil, mapRepoErr(err)
	}
	out.Enabled = cfg.Enabled()
	out.PendingEnrollment = !cfg.Enabled()
	out.RemainingRecoveryCodes = cfg.RemainingRecoveryCodes()
	return out, nil
}

// SetupMFA starts (or restarts) an enrollment: it provisions a TOTP secret and recovery codes
// and stores them with enabled_at = NULL. MFA only takes effect after EnableMFA.
func (s *Services) SetupMFA(ctx context.Context, userID uuid.UUID) (*MFASetupResult, error) {
	if err := s.requireMFA(); err != nil {
		return nil, err
	}
	user, err := s.Users.FindByID(ctx, userID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if !user.HasPassword() {
		return nil, apperr.New("mfa requires an account with a password", apperr.ErrCodeBadRequest)
	}

	provisioned, err := s.TOTP.Generate(user.Email())
	if err != nil {
		return nil, apperr.Wrap(err, "failed to generate totp secret", apperr.ErrCodeInternal)
	}
	ciphertext, err := s.MFASecrets.Encrypt(provisioned.Secret)
	if err != nil {
		return nil, apperr.Wrap(err, "failed to encrypt totp secret", apperr.ErrCodeInternal)
	}

	codes, err := generateRecoveryCodes(mfaRecoveryCodeCount)
	if err != nil {
		return nil, apperr.Wrap(err, "failed to generate recovery codes", apperr.ErrCodeInternal)
	}
	hashes := make([]string, 0, len(codes))
	for _, code := range codes {
		hash, err := s.Passwords.Hash(code)
		if err != nil {
			return nil, apperr.Wrap(err, "failed to hash recovery code", apperr.ErrCodeInternal)
		}
		hashes = append(hashes, hash)
	}

	now := s.clock().Now()
	existing, err := s.MFA.FindByUserID(ctx, userID)
	switch {
	case err == nil:
		if err := existing.Reenroll(ciphertext, hashes, now); err != nil {
			return nil, mapDomainErr(err)
		}
		if err := s.MFA.Update(ctx, existing); err != nil {
			return nil, mapRepoErr(err)
		}
	case errors.Is(err, domain.ErrNotFound):
		cfg, err := domain.NewMFAConfig(userID, ciphertext, hashes, now)
		if err != nil {
			return nil, mapDomainErr(err)
		}
		if err := s.MFA.Save(ctx, cfg); err != nil {
			return nil, mapRepoErr(err)
		}
	default:
		return nil, mapRepoErr(err)
	}

	return &MFASetupResult{
		Secret:        provisioned.Secret,
		OTPAuthURL:    provisioned.OTPAuthURL,
		RecoveryCodes: codes,
	}, nil
}

// EnableMFA turns on the second factor once the user proves possession of the pending secret.
func (s *Services) EnableMFA(ctx context.Context, userID uuid.UUID, code string) error {
	if err := s.requireMFA(); err != nil {
		return err
	}
	cfg, err := s.MFA.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apperr.New("start mfa enrollment first", apperr.ErrCodeBadRequest)
		}
		return mapRepoErr(err)
	}
	if cfg.Enabled() {
		return apperr.Wrap(domain.ErrMFAAlreadyEnabled, "mfa is already enabled", apperr.ErrCodeConflict)
	}
	if err := s.verifyTOTP(cfg, code); err != nil {
		return err
	}
	if err := cfg.Enable(s.clock().Now()); err != nil {
		return mapDomainErr(err)
	}
	if err := s.MFA.Update(ctx, cfg); err != nil {
		return mapRepoErr(err)
	}
	return nil
}

// DisableMFA clears the second factor after a password confirmation.
func (s *Services) DisableMFA(ctx context.Context, userID uuid.UUID, password string) error {
	if s.MFA == nil {
		return apperr.New("mfa is not configured", apperr.ErrCodeInternal)
	}
	user, err := s.Users.FindByID(ctx, userID)
	if err != nil {
		return mapRepoErr(err)
	}
	if !user.HasPassword() {
		return apperr.New("mfa requires an account with a password", apperr.ErrCodeBadRequest)
	}
	if err := s.Passwords.Compare(user.PasswordHash(), password); err != nil {
		return apperr.New("invalid password", apperr.ErrCodeUnauthorized)
	}
	if _, err := s.MFA.FindByUserID(ctx, userID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apperr.Wrap(domain.ErrMFANotEnabled, "mfa is not enabled", apperr.ErrCodeConflict)
		}
		return mapRepoErr(err)
	}
	if err := s.MFA.Delete(ctx, userID); err != nil {
		return mapRepoErr(err)
	}
	return nil
}

// VerifyMFAInput completes a login challenge with either a TOTP code or a recovery code.
type VerifyMFAInput struct {
	Token        string
	Code         string
	RecoveryCode string
	DeviceID     string
	TrustDevice  bool
}

// VerifyMFA consumes a login challenge, validates the second factor and issues login tokens.
func (s *Services) VerifyMFA(ctx context.Context, tokens auth.TokenService, in VerifyMFAInput) (*LoginResult, error) {
	if err := s.requireMFA(); err != nil {
		return nil, err
	}
	code := strings.TrimSpace(in.Code)
	recoveryCode := normalizeRecoveryCode(in.RecoveryCode)
	if (code == "") == (recoveryCode == "") {
		return nil, apperr.New("provide either code or recovery_code", apperr.ErrCodeBadRequest)
	}

	key, payload, err := s.peekMFAChallenge(in.Token)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		_ = s.Cache.Delete(key)
		return nil, apperr.New("invalid or expired mfa challenge", apperr.ErrCodeUnauthorized)
	}
	user, err := s.Users.FindByID(ctx, userID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if !user.CanLogin() {
		return nil, apperr.New("invalid email or password", apperr.ErrCodeUnauthorized)
	}
	cfg, err := s.MFA.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apperr.Wrap(domain.ErrMFANotEnabled, "mfa is not enabled", apperr.ErrCodeConflict)
		}
		return nil, mapRepoErr(err)
	}
	if !cfg.Enabled() {
		return nil, apperr.Wrap(domain.ErrMFANotEnabled, "mfa is not enabled", apperr.ErrCodeConflict)
	}

	if code != "" {
		if err := s.verifyTOTP(cfg, code); err != nil {
			s.failMFAChallenge(key, payload)
			return nil, err
		}
	} else {
		if err := s.consumeRecoveryCode(ctx, cfg, recoveryCode); err != nil {
			s.failMFAChallenge(key, payload)
			return nil, err
		}
	}

	_ = s.Cache.Delete(key)
	if in.TrustDevice {
		s.trustDevice(userID, in.DeviceID)
	}
	return s.issueLoginTokens(ctx, tokens, user)
}

// verifyTOTP decrypts the stored secret and validates a submitted code (±1 30s step).
func (s *Services) verifyTOTP(cfg *domain.MFAConfig, code string) error {
	secret, err := s.MFASecrets.Decrypt(cfg.TOTPSecretEncrypted())
	if err != nil {
		return apperr.Wrap(err, "failed to read totp secret", apperr.ErrCodeInternal)
	}
	if !s.TOTP.Verify(secret, strings.TrimSpace(code), s.clock().Now()) {
		return apperr.Wrap(domain.ErrInvalidMFACode, "invalid mfa code", apperr.ErrCodeUnauthorized)
	}
	return nil
}

// consumeRecoveryCode burns the matching single-use code. The stored hashes are scanned with
// the password hasher, so a wrong code is indistinguishable from an already used one.
func (s *Services) consumeRecoveryCode(ctx context.Context, cfg *domain.MFAConfig, recoveryCode string) error {
	matched := ""
	for _, hash := range cfg.RecoveryCodeHashes() {
		if err := s.Passwords.Compare(hash, recoveryCode); err == nil {
			matched = hash
			break
		}
	}
	if matched == "" {
		return apperr.Wrap(domain.ErrInvalidMFACode, "invalid mfa code", apperr.ErrCodeUnauthorized)
	}
	if err := cfg.ConsumeRecoveryCode(matched, s.clock().Now()); err != nil {
		return mapDomainErr(err)
	}
	if err := s.MFA.Update(ctx, cfg); err != nil {
		return mapRepoErr(err)
	}
	return nil
}

// mfaEnabledFor reports whether the user must pass a second factor. A missing repository (MFA
// not wired) or a pending enrollment both mean "no second factor".
func (s *Services) mfaEnabledFor(ctx context.Context, userID uuid.UUID) (bool, error) {
	if s.MFA == nil {
		return false, nil
	}
	cfg, err := s.MFA.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return false, nil
		}
		return false, mapRepoErr(err)
	}
	return cfg.Enabled(), nil
}

func (s *Services) requireMFA() error {
	if s.MFA == nil || s.TOTP == nil || s.MFASecrets == nil {
		return apperr.New("mfa is not configured", apperr.ErrCodeInternal)
	}
	return nil
}

func (s *Services) storeMFAChallenge(userID uuid.UUID) (string, error) {
	if s.Cache == nil {
		return "", apperr.New("mfa challenge store unavailable", apperr.ErrCodeInternal)
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", apperr.Wrap(err, "failed to generate mfa token", apperr.ErrCodeInternal)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	payload := mfaChallengePayload{
		UserID:    userID.String(),
		ExpiresAt: s.clock().Now().Add(mfaChallengeTTL),
	}
	if err := s.Cache.Set(mfaChallengeCacheKey(token), payload, mfaChallengeTTL); err != nil {
		return "", apperr.Wrap(err, "failed to store mfa challenge", apperr.ErrCodeInternal)
	}
	return token, nil
}

func (s *Services) peekMFAChallenge(token string) (string, *mfaChallengePayload, error) {
	if s.Cache == nil {
		return "", nil, apperr.New("mfa challenge store unavailable", apperr.ErrCodeInternal)
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", nil, apperr.New("invalid or expired mfa challenge", apperr.ErrCodeUnauthorized)
	}
	key := mfaChallengeCacheKey(token)
	var payload mfaChallengePayload
	if err := s.Cache.Get(key, &payload); err != nil {
		return "", nil, apperr.New("invalid or expired mfa challenge", apperr.ErrCodeUnauthorized)
	}
	if !payload.ExpiresAt.IsZero() && !s.clock().Now().Before(payload.ExpiresAt) {
		_ = s.Cache.Delete(key)
		return "", nil, apperr.New("invalid or expired mfa challenge", apperr.ErrCodeUnauthorized)
	}
	return key, &payload, nil
}

// failMFAChallenge keeps the challenge usable for a few retries without extending its original
// deadline, and drops it once the attempt budget is spent.
func (s *Services) failMFAChallenge(key string, payload *mfaChallengePayload) {
	payload.Attempts++
	remaining := payload.ExpiresAt.Sub(s.clock().Now())
	if payload.Attempts >= mfaChallengeAttempts || remaining <= 0 {
		_ = s.Cache.Delete(key)
		return
	}
	_ = s.Cache.Set(key, payload, remaining)
}

// trustDevice remembers a verified device so subsequent logins skip the challenge.
func (s *Services) trustDevice(userID uuid.UUID, deviceID string) {
	key, ok := mfaTrustCacheKey(userID, deviceID)
	if !ok || s.Cache == nil {
		return
	}
	_ = s.Cache.Set(key, true, mfaTrustTTL)
}

func (s *Services) deviceTrusted(userID uuid.UUID, deviceID string) bool {
	key, ok := mfaTrustCacheKey(userID, deviceID)
	if !ok || s.Cache == nil {
		return false
	}
	trusted, err := s.Cache.Exists(key)
	return err == nil && trusted
}

func mfaChallengeCacheKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return "mfa_challenge:" + hex.EncodeToString(sum[:])
}

func mfaTrustCacheKey(userID uuid.UUID, deviceID string) (string, bool) {
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return "", false
	}
	sum := sha256.Sum256([]byte(deviceID))
	return "mfa_trust:" + userID.String() + ":" + hex.EncodeToString(sum[:]), true
}

// generateRecoveryCodes returns n single-use codes formatted xxxx-xxxx-xxxx. The alphabet drops
// visually ambiguous characters (i, l, o, 0, 1) so codes survive being written down.
func generateRecoveryCodes(n int) ([]string, error) {
	const alphabet = "abcdefghjkmnpqrstuvwxyz23456789"
	const groups, groupSize = 3, 4

	codes := make([]string, 0, n)
	buf := make([]byte, groups*groupSize)
	for i := 0; i < n; i++ {
		if _, err := rand.Read(buf); err != nil {
			return nil, err
		}
		var sb strings.Builder
		for j, b := range buf {
			if j > 0 && j%groupSize == 0 {
				sb.WriteByte('-')
			}
			sb.WriteByte(alphabet[int(b)%len(alphabet)])
		}
		codes = append(codes, sb.String())
	}
	return codes, nil
}

func normalizeRecoveryCode(code string) string {
	return strings.ToLower(strings.TrimSpace(code))
}
