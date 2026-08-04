package application_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"hublio/internal/identity/application"
	"hublio/internal/identity/domain"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/auth"
	"hublio/internal/platform/cache"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type memCache struct {
	data map[string][]byte
}

func newMemCache() *memCache { return &memCache{data: map[string][]byte{}} }

func (m *memCache) Get(key string, dest any) error {
	raw, ok := m.data[key]
	if !ok {
		return errors.New("missing")
	}
	return json.Unmarshal(raw, dest)
}

func (m *memCache) Set(key string, value any, ttl time.Duration) error {
	_ = ttl
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	m.data[key] = raw
	return nil
}

func (m *memCache) Delete(key string) error {
	delete(m.data, key)
	return nil
}

func (m *memCache) Exists(key string) (bool, error) {
	_, ok := m.data[key]
	return ok, nil
}

var _ cache.RedisCacheService = (*memCache)(nil)

type fakeOAuth struct {
	profile *application.OAuthProfile
}

func (f fakeOAuth) Exchange(ctx context.Context, provider domain.OAuthProvider, code, codeVerifier, redirectURI string) (*application.OAuthProfile, error) {
	_, _, _, _, _ = ctx, provider, code, codeVerifier, redirectURI
	return f.profile, nil
}

func (f fakeOAuth) Configured(provider domain.OAuthProvider) bool {
	_ = provider
	return true
}

type memUsers struct {
	byEmail map[string]*domain.User
	byID    map[uuid.UUID]*domain.User
}

func newMemUsers() *memUsers {
	return &memUsers{byEmail: map[string]*domain.User{}, byID: map[uuid.UUID]*domain.User{}}
}

func (m *memUsers) Save(ctx context.Context, user *domain.User) error {
	_ = ctx
	m.byEmail[user.Email()] = user
	m.byID[user.ID()] = user
	return nil
}

func (m *memUsers) Update(ctx context.Context, user *domain.User) error {
	return m.Save(ctx, user)
}

func (m *memUsers) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	_ = ctx
	u, ok := m.byID[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (m *memUsers) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	_ = ctx
	u, ok := m.byEmail[email]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (m *memUsers) ListByOrganization(ctx context.Context, organizationID uuid.UUID) ([]*domain.User, error) {
	_ = ctx
	out := make([]*domain.User, 0)
	for _, u := range m.byID {
		if u.OrganizationID() == organizationID && u.DeletedAt() == nil {
			out = append(out, u)
		}
	}
	return out, nil
}

type memOAuthIdentities struct {
	byKey map[string]*domain.OAuthIdentity
}

func newMemOAuthIdentities() *memOAuthIdentities {
	return &memOAuthIdentities{byKey: map[string]*domain.OAuthIdentity{}}
}

func (m *memOAuthIdentities) Save(ctx context.Context, identity *domain.OAuthIdentity) error {
	_ = ctx
	m.byKey[string(identity.Provider())+":"+identity.ProviderSubject()] = identity
	return nil
}

func (m *memOAuthIdentities) Update(ctx context.Context, identity *domain.OAuthIdentity) error {
	return m.Save(ctx, identity)
}

func (m *memOAuthIdentities) FindByProviderSubject(ctx context.Context, provider domain.OAuthProvider, subject string) (*domain.OAuthIdentity, error) {
	_ = ctx
	id, ok := m.byKey[string(provider)+":"+subject]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return id, nil
}

type tokenAdapter struct{}

func (tokenAdapter) GenerateAccessToken(subject auth.TokenSubject) (string, error) {
	return "access-" + subject.UserID, nil
}
func (tokenAdapter) GenerateRefreshToken(subject auth.TokenSubject) (auth.RefreshToken, error) {
	return auth.RefreshToken{Token: "refresh-" + subject.UserID, UserID: subject.UserID, ExpiresAt: time.Now().Add(time.Hour)}, nil
}
func (tokenAdapter) StoreRefreshToken(token auth.RefreshToken) error { return nil }
func (tokenAdapter) ValidateRefreshToken(token string) (auth.RefreshToken, error) {
	return auth.RefreshToken{}, errors.New("unused")
}
func (tokenAdapter) RevokeRefreshToken(token string) error { return nil }
func (tokenAdapter) ParseToken(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	return nil, nil, errors.New("unused")
}
func (tokenAdapter) DecryptAccessTokenPayload(tokenString string) (*auth.EncryptedPayload, error) {
	return nil, errors.New("unused")
}

var _ auth.TokenService = tokenAdapter{}

func TestOAuthCallback_RequiresOnboardingForNewUser(t *testing.T) {
	svc := &application.Services{
		Users:           newMemUsers(),
		OAuthIdentities: newMemOAuthIdentities(),
		OAuth: fakeOAuth{profile: &application.OAuthProfile{
			Provider: domain.OAuthProviderGoogle,
			Subject:  "sub-1",
			Email:    "new@example.com",
			FullName: "New User",
			Verified: true,
		}},
		Cache: newMemCache(),
	}

	result, err := svc.OAuthCallback(context.Background(), tokenAdapter{}, application.OAuthCallbackInput{
		Provider:     "google",
		Code:         "code",
		CodeVerifier: "verifier",
		RedirectURI:  "http://localhost/callback",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "onboarding_required" {
		t.Fatalf("status=%s", result.Status)
	}
	if result.OnboardingToken == "" || result.Email != "new@example.com" {
		t.Fatalf("unexpected onboarding payload: %+v", result)
	}
}

func TestOAuthCallback_RejectsPlatformAdmin(t *testing.T) {
	now := time.Now().UTC()
	admin := domain.ReconstituteUser(
		uuid.MustParse("01900000-0000-7000-8000-000000000011"),
		uuid.MustParse("01900000-0000-7000-8000-000000000001"),
		"admin@hublio.local",
		"Admin",
		"hash",
		true,
		true,
		domain.UserStatusActive,
		now, now, nil, nil, nil, nil,
	)
	users := newMemUsers()
	_ = users.Save(context.Background(), admin)

	svc := &application.Services{
		Users:           users,
		OAuthIdentities: newMemOAuthIdentities(),
		OAuth: fakeOAuth{profile: &application.OAuthProfile{
			Provider: domain.OAuthProviderGoogle,
			Subject:  "sub-admin",
			Email:    "admin@hublio.local",
			FullName: "Admin",
			Verified: true,
		}},
		Cache: newMemCache(),
	}

	_, err := svc.OAuthCallback(context.Background(), tokenAdapter{}, application.OAuthCallbackInput{
		Provider:     "google",
		Code:         "code",
		CodeVerifier: "verifier",
		RedirectURI:  "http://localhost/callback",
	})
	if err == nil {
		t.Fatal("expected platform admin oauth rejection")
	}
	var ae *apperr.AppError
	if !errors.As(err, &ae) || ae.Code != apperr.ErrCodeForbidden {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestOAuthCallback_RejectsUnverifiedEmail(t *testing.T) {
	svc := &application.Services{
		Users:           newMemUsers(),
		OAuthIdentities: newMemOAuthIdentities(),
		OAuth: fakeOAuth{profile: &application.OAuthProfile{
			Provider: domain.OAuthProviderGoogle,
			Subject:  "sub-2",
			Email:    "x@example.com",
			FullName: "X",
			Verified: false,
		}},
		Cache: newMemCache(),
	}
	_, err := svc.OAuthCallback(context.Background(), tokenAdapter{}, application.OAuthCallbackInput{
		Provider:     "google",
		Code:         "code",
		CodeVerifier: "verifier",
		RedirectURI:  "http://localhost/callback",
	})
	if err == nil {
		t.Fatal("expected unverified rejection")
	}
}

func TestOAuthCallback_LinksExistingTenantUser(t *testing.T) {
	now := time.Now().UTC()
	user, err := domain.NewUser(
		uuid.MustParse("01900000-0000-7000-8000-000000000012"),
		uuid.MustParse("01900000-0000-7000-8000-000000000002"),
		"demo@hublio.local",
		"Demo",
		"hash",
		now,
	)
	if err != nil {
		t.Fatal(err)
	}
	users := newMemUsers()
	_ = users.Save(context.Background(), user)
	idents := newMemOAuthIdentities()

	svc := &application.Services{
		Users:           users,
		OAuthIdentities: idents,
		OAuth: fakeOAuth{profile: &application.OAuthProfile{
			Provider: domain.OAuthProviderGoogle,
			Subject:  "sub-demo",
			Email:    "demo@hublio.local",
			FullName: "Demo",
			Verified: true,
		}},
		Cache: newMemCache(),
	}

	result, err := svc.OAuthCallback(context.Background(), tokenAdapter{}, application.OAuthCallbackInput{
		Provider:     "google",
		Code:         "code",
		CodeVerifier: "verifier",
		RedirectURI:  "http://localhost/callback",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "authenticated" || result.AccessToken == "" {
		t.Fatalf("expected authenticated, got %+v", result)
	}
	if _, err := idents.FindByProviderSubject(context.Background(), domain.OAuthProviderGoogle, "sub-demo"); err != nil {
		t.Fatal("expected identity linked")
	}
}
