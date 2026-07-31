package application_test

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"hublio/internal/identity/application"
	"hublio/internal/identity/domain"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/auth"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// rotatingTokens is an in-memory TokenService that stores refresh tokens for rotate tests.
type rotatingTokens struct {
	mu      sync.Mutex
	byToken map[string]auth.RefreshToken
	n       int
}

func newRotatingTokens() *rotatingTokens {
	return &rotatingTokens{byToken: map[string]auth.RefreshToken{}}
}

func (t *rotatingTokens) GenerateAccessToken(subject auth.TokenSubject) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.n++
	return "access-" + subject.UserID + "-" + itoa(t.n), nil
}

func (t *rotatingTokens) GenerateRefreshToken(subject auth.TokenSubject) (auth.RefreshToken, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.n++
	tok := "refresh-" + subject.UserID + "-" + itoa(t.n)
	return auth.RefreshToken{
		Token:     tok,
		UserID:    subject.UserID,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
		Revoked:   false,
	}, nil
}

func (t *rotatingTokens) StoreRefreshToken(token auth.RefreshToken) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.byToken[token.Token] = token
	return nil
}

func (t *rotatingTokens) ValidateRefreshToken(token string) (auth.RefreshToken, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	stored, ok := t.byToken[token]
	if !ok || stored.Revoked || stored.ExpiresAt.Before(time.Now().UTC()) {
		return auth.RefreshToken{}, apperr.New("Invalid refresh token", apperr.ErrCodeUnauthorized)
	}
	return stored, nil
}

func (t *rotatingTokens) RevokeRefreshToken(token string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	stored, ok := t.byToken[token]
	if !ok {
		return apperr.New("Invalid refresh token", apperr.ErrCodeUnauthorized)
	}
	stored.Revoked = true
	t.byToken[token] = stored
	return nil
}

func (t *rotatingTokens) ParseToken(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	_, _ = tokenString, jwt.MapClaims{}
	return nil, nil, apperr.New("unused", apperr.ErrCodeInternal)
}

func (t *rotatingTokens) DecryptAccessTokenPayload(tokenString string) (*auth.EncryptedPayload, error) {
	_ = tokenString
	return nil, apperr.New("unused", apperr.ErrCodeInternal)
}

func (t *rotatingTokens) isRevoked(token string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	stored, ok := t.byToken[token]
	return ok && stored.Revoked
}

var _ auth.TokenService = (*rotatingTokens)(nil)

func itoa(n int) string {
	return strconv.Itoa(n)
}

func activeUser(t *testing.T, email string) *domain.User {
	t.Helper()
	now := time.Now().UTC()
	return domain.ReconstituteUser(
		uuid.Must(uuid.NewV7()),
		uuid.Must(uuid.NewV7()),
		email,
		"Test User",
		"hash",
		true,
		false,
		domain.UserStatusActive,
		now, now, nil, nil, nil, nil,
	)
}

func TestRefreshTokens(t *testing.T) {
	t.Parallel()

	t.Run("rotates valid refresh token", func(t *testing.T) {
		t.Parallel()
		users := newMemUsers()
		user := activeUser(t, "refresh@example.com")
		_ = users.Save(context.Background(), user)
		tokens := newRotatingTokens()
		svc := &application.Services{Users: users}

		seed, err := tokens.GenerateRefreshToken(auth.TokenSubject{UserID: user.ID().String()})
		if err != nil {
			t.Fatal(err)
		}
		if err := tokens.StoreRefreshToken(seed); err != nil {
			t.Fatal(err)
		}

		got, err := svc.RefreshTokens(context.Background(), tokens, seed.Token)
		if err != nil {
			t.Fatalf("RefreshTokens: %v", err)
		}
		if got.AccessToken == "" || got.RefreshToken == "" {
			t.Fatalf("missing tokens: %+v", got)
		}
		if got.RefreshToken == seed.Token {
			t.Fatal("expected new refresh token")
		}
		if !tokens.isRevoked(seed.Token) {
			t.Fatal("old refresh should be revoked")
		}
		if _, err := tokens.ValidateRefreshToken(seed.Token); err == nil {
			t.Fatal("old refresh must fail validate")
		}
		if user.LastLoginAt() != nil {
			t.Fatal("refresh must not bump last_login")
		}
	})

	t.Run("rejects missing token", func(t *testing.T) {
		t.Parallel()
		users := newMemUsers()
		tokens := newRotatingTokens()
		svc := &application.Services{Users: users}
		_, err := svc.RefreshTokens(context.Background(), tokens, "nope")
		if err == nil {
			t.Fatal("expected error")
		}
		ae, ok := err.(*apperr.AppError)
		if !ok || ae.Code != apperr.ErrCodeUnauthorized {
			t.Fatalf("got %#v", err)
		}
	})

	t.Run("rejects revoked token", func(t *testing.T) {
		t.Parallel()
		users := newMemUsers()
		user := activeUser(t, "revoked@example.com")
		_ = users.Save(context.Background(), user)
		tokens := newRotatingTokens()
		svc := &application.Services{Users: users}
		seed, _ := tokens.GenerateRefreshToken(auth.TokenSubject{UserID: user.ID().String()})
		_ = tokens.StoreRefreshToken(seed)
		_ = tokens.RevokeRefreshToken(seed.Token)
		_, err := svc.RefreshTokens(context.Background(), tokens, seed.Token)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects expired token", func(t *testing.T) {
		t.Parallel()
		users := newMemUsers()
		user := activeUser(t, "expired@example.com")
		_ = users.Save(context.Background(), user)
		tokens := newRotatingTokens()
		svc := &application.Services{Users: users}
		expired := auth.RefreshToken{
			Token:     "expired-token",
			UserID:    user.ID().String(),
			ExpiresAt: time.Now().UTC().Add(-time.Minute),
		}
		_ = tokens.StoreRefreshToken(expired)
		_, err := svc.RefreshTokens(context.Background(), tokens, expired.Token)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects inactive user", func(t *testing.T) {
		t.Parallel()
		users := newMemUsers()
		now := time.Now().UTC()
		user := domain.ReconstituteUser(
			uuid.Must(uuid.NewV7()),
			uuid.Must(uuid.NewV7()),
			"inactive@example.com",
			"Inactive",
			"hash",
			false,
			false,
			domain.UserStatusSuspended,
			now, now, nil, nil, nil, nil,
		)
		_ = users.Save(context.Background(), user)
		tokens := newRotatingTokens()
		svc := &application.Services{Users: users}
		seed, _ := tokens.GenerateRefreshToken(auth.TokenSubject{UserID: user.ID().String()})
		_ = tokens.StoreRefreshToken(seed)
		_, err := svc.RefreshTokens(context.Background(), tokens, seed.Token)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("login still bumps last_login", func(t *testing.T) {
		t.Parallel()
		users := newMemUsers()
		user := newPasswordUser(t, "loginbump@example.com")
		_ = users.Save(context.Background(), user)
		tokens := newRotatingTokens()
		svc := &application.Services{
			Users:     users,
			Passwords: plainHasher{},
		}
		got, err := svc.Login(context.Background(), tokens, application.LoginInput{
			Email:    "loginbump@example.com",
			Password: "original",
		})
		if err != nil {
			t.Fatalf("Login: %v", err)
		}
		if got.MFARequired {
			t.Fatal("unexpected MFA")
		}
		reloaded, err := users.FindByID(context.Background(), user.ID())
		if err != nil {
			t.Fatal(err)
		}
		if reloaded.LastLoginAt() == nil {
			t.Fatal("login should set last_login")
		}
	})
}
