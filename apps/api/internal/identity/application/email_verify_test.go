package application_test

import (
	"context"
	"testing"
	"time"

	"hublio/internal/identity/application"
	"hublio/internal/identity/domain"

	"github.com/google/uuid"
)

func TestRequestEmailVerification_AntiEnumerationAndVerifiedSkip(t *testing.T) {
	verified, err := domain.NewUser(
		uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), "verified@example.com", "V", "hashed:x", time.Now().UTC(),
	)
	if err != nil {
		t.Fatal(err)
	}
	verified.MarkEmailVerified(time.Now().UTC())

	tests := []struct {
		name       string
		email      string
		seed       *domain.User
		wantStored bool
		wantMail   bool
	}{
		{name: "unknown email stays silent", email: "missing@example.com", wantStored: false, wantMail: false},
		{name: "unverified user gets code", email: "user@example.com", seed: newPasswordUser(t, "user@example.com"), wantStored: true, wantMail: true},
		{name: "already verified stays silent", email: "verified@example.com", seed: verified, wantStored: false, wantMail: false},
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

			if err := svc.RequestEmailVerification(context.Background(), tc.email); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := cache.countPrefix("email_verify:") > 0; got != tc.wantStored {
				t.Fatalf("stored = %v, want %v", got, tc.wantStored)
			}
			if got := len(mailer.sent) > 0; got != tc.wantMail {
				t.Fatalf("mail sent = %v, want %v", got, tc.wantMail)
			}
		})
	}
}

func TestVerifyEmail(t *testing.T) {
	setup := func() (*application.Services, *memUsers, *domain.User) {
		users := newMemUsers()
		user := newPasswordUser(t, "verify@example.com")
		_ = users.Save(context.Background(), user)
		svc := &application.Services{
			Users:     users,
			Passwords: plainHasher{},
			Cache:     newMemCache(),
			Mail:      &recordingMailer{},
		}
		if err := svc.RequestEmailVerification(context.Background(), "verify@example.com"); err != nil {
			t.Fatal(err)
		}
		return svc, users, user
	}

	t.Run("wrong code is rejected and does not verify", func(t *testing.T) {
		svc, users, user := setup()
		if err := svc.VerifyEmail(context.Background(), "verify@example.com", "000000"); err == nil {
			t.Fatal("expected wrong code rejection")
		}
		got, _ := users.FindByID(context.Background(), user.ID())
		if got.EmailVerifiedAt() != nil {
			t.Fatal("email should not be verified after wrong code")
		}
	})

	t.Run("correct code verifies and consumes", func(t *testing.T) {
		svc, users, user := setup()
		mailer := svc.Mail.(*recordingMailer)
		code := extractCode(t, mailer.sent[len(mailer.sent)-1].body)

		if err := svc.VerifyEmail(context.Background(), "verify@example.com", code); err != nil {
			t.Fatalf("verify failed: %v", err)
		}
		got, _ := users.FindByID(context.Background(), user.ID())
		if got.EmailVerifiedAt() == nil {
			t.Fatal("expected email verified")
		}

		// Code is one-time: a second attempt fails.
		if err := svc.VerifyEmail(context.Background(), "verify@example.com", code); err == nil {
			t.Fatal("expected consumed code to be rejected")
		}
	})
}

// extractCode pulls the 6-digit code out of the verification email body.
func extractCode(t *testing.T, body string) string {
	t.Helper()
	const marker = "code is: "
	idx := indexOf(body, marker)
	if idx < 0 {
		t.Fatalf("code not found in body: %s", body)
	}
	rest := body[idx+len(marker):]
	if len(rest) < 6 {
		t.Fatalf("unexpected code body: %s", body)
	}
	return rest[:6]
}
