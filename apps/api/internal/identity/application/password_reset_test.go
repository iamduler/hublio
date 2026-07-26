package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"hublio/internal/identity/application"
	"hublio/internal/identity/domain"
	"hublio/internal/platform/apperr"

	"github.com/google/uuid"
)

// plainHasher is a deterministic, dependency-free PasswordHasher for tests.
type plainHasher struct{}

func (plainHasher) Hash(password string) (string, error) { return "hashed:" + password, nil }
func (plainHasher) Compare(hash, password string) error {
	if hash == "hashed:"+password {
		return nil
	}
	return errors.New("mismatch")
}

var _ domain.PasswordHasher = plainHasher{}

func newPasswordUser(t *testing.T, email string) *domain.User {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	orgID, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	user, err := domain.NewUser(id, orgID, email, "Test User", "hashed:original", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	return user
}

func TestRequestPasswordReset_AntiEnumeration(t *testing.T) {
	oauthOnly, err := domain.NewOAuthUser(
		uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), "oauth@example.com", "OAuth User", time.Now().UTC(),
	)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		email      string
		seed       *domain.User
		wantStored bool
		wantMail   bool
	}{
		{name: "unknown email stays silent", email: "missing@example.com", wantStored: false, wantMail: false},
		{name: "password user gets token", email: "user@example.com", seed: newPasswordUser(t, "user@example.com"), wantStored: true, wantMail: true},
		{name: "oauth-only user stays silent", email: "oauth@example.com", seed: oauthOnly, wantStored: false, wantMail: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			users := newMemUsers()
			if tc.seed != nil {
				_ = users.Save(context.Background(), tc.seed)
			}
			mailer := &recordingMailer{}
			cache := newMemCache()
			svc := &application.Services{
				Users:     users,
				Passwords: plainHasher{},
				Cache:     cache,
				Mail:      mailer,
			}

			if err := svc.RequestPasswordReset(context.Background(), tc.email); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			gotStored := cache.countPrefix("password_reset:") > 0
			if gotStored != tc.wantStored {
				t.Fatalf("stored token = %v, want %v", gotStored, tc.wantStored)
			}
			if (len(mailer.sent) > 0) != tc.wantMail {
				t.Fatalf("mail sent = %v, want %v", len(mailer.sent) > 0, tc.wantMail)
			}
		})
	}
}

func TestResetPassword_ConsumesTokenAndSetsHash(t *testing.T) {
	users := newMemUsers()
	user := newPasswordUser(t, "reset@example.com")
	_ = users.Save(context.Background(), user)

	cache := newMemCache()
	svc := &application.Services{
		Users:     users,
		Passwords: plainHasher{},
		Cache:     cache,
		Mail:      &recordingMailer{},
	}

	if err := svc.RequestPasswordReset(context.Background(), "reset@example.com"); err != nil {
		t.Fatal(err)
	}
	token := extractResetToken(t, svc)

	// First reset succeeds and updates the hash + password_changed_at.
	if err := svc.ResetPassword(context.Background(), application.ResetPasswordInput{
		Token:    token,
		Password: "new-strong-pass",
	}); err != nil {
		t.Fatalf("reset failed: %v", err)
	}
	updated, _ := users.FindByID(context.Background(), user.ID())
	if updated.PasswordHash() != "hashed:new-strong-pass" {
		t.Fatalf("password hash not updated: %s", updated.PasswordHash())
	}
	if updated.PasswordChangedAt() == nil {
		t.Fatal("expected password_changed_at to be set")
	}

	// Token is one-time: reusing it fails.
	err := svc.ResetPassword(context.Background(), application.ResetPasswordInput{
		Token:    token,
		Password: "another-strong-pass",
	})
	if err == nil {
		t.Fatal("expected reused token to be rejected")
	}
	var ae *apperr.AppError
	if !errors.As(err, &ae) || ae.Code != apperr.ErrCodeBadRequest {
		t.Fatalf("expected bad request, got %v", err)
	}
}

func TestResetPassword_RejectsShortPassword(t *testing.T) {
	svc := &application.Services{
		Users:     newMemUsers(),
		Passwords: plainHasher{},
		Cache:     newMemCache(),
	}
	err := svc.ResetPassword(context.Background(), application.ResetPasswordInput{
		Token:    "whatever",
		Password: "short",
	})
	if err == nil {
		t.Fatal("expected short password rejection")
	}
	var ae *apperr.AppError
	if !errors.As(err, &ae) || ae.Code != apperr.ErrCodeBadRequest {
		t.Fatalf("expected bad request, got %v", err)
	}
}

// extractResetToken derives the raw token from the single reset entry in the cache by brute
// force is not possible (sha256-keyed); instead we intercept it via a fresh request using a
// deterministic mailer link. We re-issue the token here through the recording mailer captured
// during RequestPasswordReset.
func extractResetToken(t *testing.T, svc *application.Services) string {
	t.Helper()
	mailer, ok := svc.Mail.(*recordingMailer)
	if !ok || len(mailer.sent) == 0 {
		t.Fatal("no reset email captured")
	}
	body := mailer.sent[len(mailer.sent)-1].body
	const marker = "token="
	idx := indexOf(body, marker)
	if idx < 0 {
		t.Fatalf("reset link not found in body: %s", body)
	}
	rest := body[idx+len(marker):]
	// Token runs until the first whitespace/newline.
	end := 0
	for end < len(rest) && rest[end] != '\n' && rest[end] != ' ' && rest[end] != '\r' {
		end++
	}
	return rest[:end]
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
