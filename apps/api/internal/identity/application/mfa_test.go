package application_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"hublio/internal/identity/application"
	"hublio/internal/identity/domain"
	"hublio/internal/platform/apperr"

	"github.com/google/uuid"
)

// memMFA is an in-memory domain.MFARepository.
type memMFA struct {
	byUser map[uuid.UUID]*domain.MFAConfig
}

func newMemMFA() *memMFA { return &memMFA{byUser: map[uuid.UUID]*domain.MFAConfig{}} }

func (m *memMFA) Save(ctx context.Context, cfg *domain.MFAConfig) error {
	_ = ctx
	m.byUser[cfg.UserID()] = cfg
	return nil
}

func (m *memMFA) Update(ctx context.Context, cfg *domain.MFAConfig) error {
	return m.Save(ctx, cfg)
}

func (m *memMFA) FindByUserID(ctx context.Context, userID uuid.UUID) (*domain.MFAConfig, error) {
	_ = ctx
	cfg, ok := m.byUser[userID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return cfg, nil
}

func (m *memMFA) Delete(ctx context.Context, userID uuid.UUID) error {
	_ = ctx
	delete(m.byUser, userID)
	return nil
}

var _ domain.MFARepository = (*memMFA)(nil)

// stubTOTP accepts one fixed code, so tests never depend on wall-clock time steps.
type stubTOTP struct {
	validCode string
}

func (s stubTOTP) Generate(accountName string) (application.TOTPSecret, error) {
	return application.TOTPSecret{
		Secret:     "SECRET-" + accountName,
		OTPAuthURL: "otpauth://totp/Hublio:" + accountName + "?secret=SECRET",
	}, nil
}

func (s stubTOTP) Verify(secret, code string, at time.Time) bool {
	_, _ = secret, at
	return code == s.validCode
}

var _ application.TOTPVerifier = stubTOTP{}

// reversibleCipher stands in for AES-GCM: encryption must round-trip, nothing more.
type reversibleCipher struct{}

func (reversibleCipher) Encrypt(plaintext string) (string, error) {
	return "enc:" + plaintext, nil
}

func (reversibleCipher) Decrypt(ciphertext string) (string, error) {
	if !strings.HasPrefix(ciphertext, "enc:") {
		return "", errors.New("bad ciphertext")
	}
	return strings.TrimPrefix(ciphertext, "enc:"), nil
}

var _ application.MFASecretCipher = reversibleCipher{}

type mfaFixture struct {
	svc   *application.Services
	users *memUsers
	mfa   *memMFA
	cache *memCache
	user  *domain.User
}

func newMFAFixture(t *testing.T, validCode string) *mfaFixture {
	t.Helper()
	users := newMemUsers()
	user := newPasswordUser(t, "mfa@example.com")
	_ = users.Save(context.Background(), user)

	mfa := newMemMFA()
	cache := newMemCache()
	return &mfaFixture{
		svc: &application.Services{
			Users:      users,
			MFA:        mfa,
			Passwords:  plainHasher{},
			TOTP:       stubTOTP{validCode: validCode},
			MFASecrets: reversibleCipher{},
			Cache:      cache,
		},
		users: users,
		mfa:   mfa,
		cache: cache,
		user:  user,
	}
}

// enrol runs setup + enable and returns the plaintext recovery codes.
func (f *mfaFixture) enrol(t *testing.T, code string) []string {
	t.Helper()
	setup, err := f.svc.SetupMFA(context.Background(), f.user.ID())
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := f.svc.EnableMFA(context.Background(), f.user.ID(), code); err != nil {
		t.Fatalf("enable: %v", err)
	}
	return setup.RecoveryCodes
}

func (f *mfaFixture) login(t *testing.T) *application.LoginResult {
	t.Helper()
	result, err := f.svc.Login(context.Background(), tokenAdapter{}, application.LoginInput{
		Email:    "mfa@example.com",
		Password: "original",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	return result
}

func TestSetupMFA_StoresPendingEnrollment(t *testing.T) {
	f := newMFAFixture(t, "123456")

	setup, err := f.svc.SetupMFA(context.Background(), f.user.ID())
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if setup.Secret == "" || setup.OTPAuthURL == "" {
		t.Fatalf("expected secret and otpauth url, got %+v", setup)
	}
	if len(setup.RecoveryCodes) != 10 {
		t.Fatalf("recovery codes = %d, want 10", len(setup.RecoveryCodes))
	}

	cfg, err := f.mfa.FindByUserID(context.Background(), f.user.ID())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Enabled() {
		t.Fatal("enrollment must stay pending until EnableMFA")
	}
	if cfg.TOTPSecretEncrypted() == setup.Secret {
		t.Fatal("totp secret must be stored encrypted")
	}
	for _, hash := range cfg.RecoveryCodeHashes() {
		for _, code := range setup.RecoveryCodes {
			if hash == code {
				t.Fatal("recovery codes must be stored hashed")
			}
		}
	}
}

func TestEnableMFA_RequiresValidCode(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		wantCode apperr.ErrorCode
	}{
		{name: "valid code enables", code: "123456"},
		{name: "wrong code rejected", code: "000000", wantCode: apperr.ErrCodeUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newMFAFixture(t, "123456")
			if _, err := f.svc.SetupMFA(context.Background(), f.user.ID()); err != nil {
				t.Fatal(err)
			}

			err := f.svc.EnableMFA(context.Background(), f.user.ID(), tc.code)
			cfg, findErr := f.mfa.FindByUserID(context.Background(), f.user.ID())
			if findErr != nil {
				t.Fatal(findErr)
			}

			if tc.wantCode == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !cfg.Enabled() {
					t.Fatal("expected mfa enabled")
				}
				return
			}
			var ae *apperr.AppError
			if !errors.As(err, &ae) || ae.Code != tc.wantCode {
				t.Fatalf("error = %v, want code %s", err, tc.wantCode)
			}
			if cfg.Enabled() {
				t.Fatal("mfa must stay disabled after a failed enable")
			}
		})
	}
}

func TestLogin_IssuesChallengeOnlyWhenMFAEnabled(t *testing.T) {
	tests := []struct {
		name       string
		enrol      bool
		wantMFA    bool
		wantTokens bool
	}{
		{name: "no mfa issues tokens", enrol: false, wantMFA: false, wantTokens: true},
		{name: "mfa enabled issues challenge", enrol: true, wantMFA: true, wantTokens: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newMFAFixture(t, "123456")
			if tc.enrol {
				f.enrol(t, "123456")
			}

			result := f.login(t)

			if result.MFARequired != tc.wantMFA {
				t.Fatalf("mfa_required = %v, want %v", result.MFARequired, tc.wantMFA)
			}
			if (result.AccessToken != "") != tc.wantTokens {
				t.Fatalf("access token issued = %v, want %v", result.AccessToken != "", tc.wantTokens)
			}
			if tc.wantMFA {
				if result.MFAToken == "" {
					t.Fatal("expected an mfa token")
				}
				if result.RefreshToken != "" {
					t.Fatal("refresh token must not be issued before verification")
				}
				if f.cache.countPrefix("mfa_challenge:") != 1 {
					t.Fatal("expected a stored mfa challenge")
				}
			}
		})
	}
}

func TestVerifyMFA_TOTPAndRecoveryCodes(t *testing.T) {
	t.Run("valid totp issues tokens and consumes the challenge", func(t *testing.T) {
		f := newMFAFixture(t, "123456")
		f.enrol(t, "123456")
		challenge := f.login(t)

		result, err := f.svc.VerifyMFA(context.Background(), tokenAdapter{}, application.VerifyMFAInput{
			Token: challenge.MFAToken,
			Code:  "123456",
		})
		if err != nil {
			t.Fatalf("verify: %v", err)
		}
		if result.AccessToken == "" || result.RefreshToken == "" || result.MFARequired {
			t.Fatalf("expected login tokens, got %+v", result)
		}
		if f.cache.countPrefix("mfa_challenge:") != 0 {
			t.Fatal("challenge must be consumed")
		}
	})

	t.Run("invalid totp is rejected and issues no tokens", func(t *testing.T) {
		f := newMFAFixture(t, "123456")
		f.enrol(t, "123456")
		challenge := f.login(t)

		_, err := f.svc.VerifyMFA(context.Background(), tokenAdapter{}, application.VerifyMFAInput{
			Token: challenge.MFAToken,
			Code:  "999999",
		})
		var ae *apperr.AppError
		if !errors.As(err, &ae) || ae.Code != apperr.ErrCodeUnauthorized {
			t.Fatalf("error = %v, want unauthorized", err)
		}
	})

	t.Run("expired or unknown challenge is rejected", func(t *testing.T) {
		f := newMFAFixture(t, "123456")
		f.enrol(t, "123456")

		_, err := f.svc.VerifyMFA(context.Background(), tokenAdapter{}, application.VerifyMFAInput{
			Token: "not-a-real-token",
			Code:  "123456",
		})
		var ae *apperr.AppError
		if !errors.As(err, &ae) || ae.Code != apperr.ErrCodeUnauthorized {
			t.Fatalf("error = %v, want unauthorized", err)
		}
	})

	t.Run("recovery code works exactly once", func(t *testing.T) {
		f := newMFAFixture(t, "123456")
		codes := f.enrol(t, "123456")
		recovery := codes[0]

		challenge := f.login(t)
		result, err := f.svc.VerifyMFA(context.Background(), tokenAdapter{}, application.VerifyMFAInput{
			Token:        challenge.MFAToken,
			RecoveryCode: strings.ToUpper(recovery),
		})
		if err != nil {
			t.Fatalf("verify with recovery code: %v", err)
		}
		if result.AccessToken == "" {
			t.Fatal("expected login tokens")
		}

		cfg, err := f.mfa.FindByUserID(context.Background(), f.user.ID())
		if err != nil {
			t.Fatal(err)
		}
		if cfg.RemainingRecoveryCodes() != len(codes)-1 {
			t.Fatalf("remaining codes = %d, want %d", cfg.RemainingRecoveryCodes(), len(codes)-1)
		}

		challenge = f.login(t)
		if _, err := f.svc.VerifyMFA(context.Background(), tokenAdapter{}, application.VerifyMFAInput{
			Token:        challenge.MFAToken,
			RecoveryCode: recovery,
		}); err == nil {
			t.Fatal("expected a reused recovery code to be rejected")
		}
	})

	t.Run("code and recovery_code are mutually exclusive", func(t *testing.T) {
		f := newMFAFixture(t, "123456")
		codes := f.enrol(t, "123456")
		challenge := f.login(t)

		_, err := f.svc.VerifyMFA(context.Background(), tokenAdapter{}, application.VerifyMFAInput{
			Token:        challenge.MFAToken,
			Code:         "123456",
			RecoveryCode: codes[0],
		})
		var ae *apperr.AppError
		if !errors.As(err, &ae) || ae.Code != apperr.ErrCodeBadRequest {
			t.Fatalf("error = %v, want bad request", err)
		}
	})
}

func TestVerifyMFA_TrustedDeviceSkipsNextChallenge(t *testing.T) {
	f := newMFAFixture(t, "123456")
	f.enrol(t, "123456")

	challenge := f.login(t)
	if _, err := f.svc.VerifyMFA(context.Background(), tokenAdapter{}, application.VerifyMFAInput{
		Token:       challenge.MFAToken,
		Code:        "123456",
		DeviceID:    "device-1",
		TrustDevice: true,
	}); err != nil {
		t.Fatalf("verify: %v", err)
	}

	trusted, err := f.svc.Login(context.Background(), tokenAdapter{}, application.LoginInput{
		Email:    "mfa@example.com",
		Password: "original",
		DeviceID: "device-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if trusted.MFARequired || trusted.AccessToken == "" {
		t.Fatalf("trusted device should log in directly, got %+v", trusted)
	}

	unknown, err := f.svc.Login(context.Background(), tokenAdapter{}, application.LoginInput{
		Email:    "mfa@example.com",
		Password: "original",
		DeviceID: "device-2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !unknown.MFARequired {
		t.Fatal("unknown device must still be challenged")
	}
}

func TestDisableMFA_RequiresPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		wantCode    apperr.ErrorCode
		wantCleared bool
	}{
		{name: "correct password clears mfa", password: "original", wantCleared: true},
		{name: "wrong password rejected", password: "nope", wantCode: apperr.ErrCodeUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newMFAFixture(t, "123456")
			f.enrol(t, "123456")

			err := f.svc.DisableMFA(context.Background(), f.user.ID(), tc.password)
			_, findErr := f.mfa.FindByUserID(context.Background(), f.user.ID())
			cleared := errors.Is(findErr, domain.ErrNotFound)

			if tc.wantCode != "" {
				var ae *apperr.AppError
				if !errors.As(err, &ae) || ae.Code != tc.wantCode {
					t.Fatalf("error = %v, want code %s", err, tc.wantCode)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cleared != tc.wantCleared {
				t.Fatalf("mfa cleared = %v, want %v", cleared, tc.wantCleared)
			}
		})
	}
}
