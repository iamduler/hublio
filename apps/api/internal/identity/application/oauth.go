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
	"hublio/internal/platform/id"
)

const oauthOnboardingTTL = 10 * time.Minute

// OAuthProfile is the verified identity returned by an IdP after code exchange.
type OAuthProfile struct {
	Provider  domain.OAuthProvider
	Subject   string
	Email     string
	FullName  string
	Verified  bool
}

// OAuthExchange exchanges an authorization code with the IdP and returns a verified profile.
// Infrastructure implements this; Application never talks to Google/Microsoft/GitHub HTTP APIs directly.
type OAuthExchange interface {
	Exchange(ctx context.Context, provider domain.OAuthProvider, code, codeVerifier, redirectURI string) (*OAuthProfile, error)
	Configured(provider domain.OAuthProvider) bool
}

type OAuthCallbackInput struct {
	Provider     string
	Code         string
	CodeVerifier string
	RedirectURI  string
}

type OAuthCallbackResult struct {
	Status           string // "authenticated" | "onboarding_required"
	AccessToken      string
	RefreshToken     string
	OnboardingToken  string
	Email            string
	FullName         string
	User             *domain.User
}

type oauthOnboardingPayload struct {
	Provider  string `json:"provider"`
	Subject   string `json:"subject"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
}

func (s *Services) OAuthCallback(ctx context.Context, tokens auth.TokenService, in OAuthCallbackInput) (*OAuthCallbackResult, error) {
	if s.OAuth == nil {
		return nil, apperr.New("oauth is not configured", apperr.ErrCodeBadRequest)
	}
	provider, err := domain.ParseOAuthProvider(in.Provider)
	if err != nil {
		return nil, apperr.Wrap(err, "invalid oauth provider", apperr.ErrCodeBadRequest)
	}
	if !s.OAuth.Configured(provider) {
		return nil, apperr.New("oauth provider is not configured", apperr.ErrCodeBadRequest)
	}
	if strings.TrimSpace(in.Code) == "" || strings.TrimSpace(in.CodeVerifier) == "" || strings.TrimSpace(in.RedirectURI) == "" {
		return nil, apperr.New("missing oauth callback parameters", apperr.ErrCodeBadRequest)
	}

	profile, err := s.OAuth.Exchange(ctx, provider, in.Code, in.CodeVerifier, in.RedirectURI)
	if err != nil {
		return nil, apperr.Wrap(err, "oauth exchange failed", apperr.ErrCodeUnauthorized)
	}
	if profile == nil || !profile.Verified || strings.TrimSpace(profile.Email) == "" || strings.TrimSpace(profile.Subject) == "" {
		return nil, apperr.Wrap(domain.ErrOAuthEmailUnverified, "oauth email must be verified", apperr.ErrCodeForbidden)
	}

	email := strings.ToLower(strings.TrimSpace(profile.Email))
	fullName := strings.TrimSpace(profile.FullName)
	if fullName == "" {
		fullName = strings.Split(email, "@")[0]
	}

	if existing, err := s.Users.FindByEmail(ctx, email); err == nil {
		if existing.IsPlatformAdmin() {
			return nil, apperr.Wrap(domain.ErrOAuthPlatformAdmin, "platform admin must use password login", apperr.ErrCodeForbidden)
		}
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, mapRepoErr(err)
	}

	if identity, err := s.OAuthIdentities.FindByProviderSubject(ctx, provider, profile.Subject); err == nil {
		user, err := s.Users.FindByID(ctx, identity.UserID())
		if err != nil {
			return nil, mapRepoErr(err)
		}
		if user.IsPlatformAdmin() {
			return nil, apperr.Wrap(domain.ErrOAuthPlatformAdmin, "platform admin must use password login", apperr.ErrCodeForbidden)
		}
		result, err := s.issueLoginTokens(ctx, tokens, user)
		if err != nil {
			return nil, err
		}
		identity.RecordLogin(s.clock().Now())
		_ = s.OAuthIdentities.Update(ctx, identity)
		return &OAuthCallbackResult{
			Status:       "authenticated",
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
			User:         result.User,
		}, nil
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, mapRepoErr(err)
	}

	if user, err := s.Users.FindByEmail(ctx, email); err == nil {
		if user.IsPlatformAdmin() {
			return nil, apperr.Wrap(domain.ErrOAuthPlatformAdmin, "platform admin must use password login", apperr.ErrCodeForbidden)
		}
		now := s.clock().Now()
		identityID, err := id.NewV7()
		if err != nil {
			return nil, apperr.Wrap(err, "failed to generate oauth identity id", apperr.ErrCodeInternal)
		}
		identity, err := domain.NewOAuthIdentity(identityID, user.ID(), provider, profile.Subject, email, now)
		if err != nil {
			return nil, mapDomainErr(err)
		}
		if err := s.OAuthIdentities.Save(ctx, identity); err != nil {
			return nil, mapRepoErr(err)
		}
		user.MarkEmailVerified(now)
		result, err := s.issueLoginTokens(ctx, tokens, user)
		if err != nil {
			return nil, err
		}
		return &OAuthCallbackResult{
			Status:       "authenticated",
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
			User:         result.User,
		}, nil
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, mapRepoErr(err)
	}

	onboardingToken, err := s.storeOAuthOnboarding(oauthOnboardingPayload{
		Provider: string(provider),
		Subject:  profile.Subject,
		Email:    email,
		FullName: fullName,
	})
	if err != nil {
		return nil, err
	}
	return &OAuthCallbackResult{
		Status:          "onboarding_required",
		OnboardingToken: onboardingToken,
		Email:           email,
		FullName:        fullName,
	}, nil
}

type CompleteOAuthRegistrationInput struct {
	OnboardingToken  string
	OrganizationName string
	WorkspaceName    string
	Environment      string
}

func (s *Services) CompleteOAuthRegistration(
	ctx context.Context,
	tokens auth.TokenService,
	in CompleteOAuthRegistrationInput,
) (*LoginResult, *domain.Organization, *domain.Workspace, *domain.Membership, *domain.OAuthIdentity, error) {
	payload, err := s.consumeOAuthOnboarding(in.OnboardingToken)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	provider, err := domain.ParseOAuthProvider(payload.Provider)
	if err != nil {
		return nil, nil, nil, nil, nil, apperr.Wrap(err, "invalid oauth provider", apperr.ErrCodeBadRequest)
	}
	orgName := strings.TrimSpace(in.OrganizationName)
	if orgName == "" {
		return nil, nil, nil, nil, nil, apperr.New("organization name is required", apperr.ErrCodeBadRequest)
	}

	if _, err := s.Users.FindByEmail(ctx, payload.Email); err == nil {
		return nil, nil, nil, nil, nil, apperr.New("user already exists", apperr.ErrCodeConflict)
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, nil, nil, nil, nil, mapRepoErr(err)
	}

	wsName := in.WorkspaceName
	if strings.TrimSpace(wsName) == "" {
		wsName = "default"
	}
	env := in.Environment
	if strings.TrimSpace(env) == "" {
		env = "production"
	}

	now := s.clock().Now()
	orgID, err := id.NewV7()
	if err != nil {
		return nil, nil, nil, nil, nil, apperr.Wrap(err, "failed to generate organization id", apperr.ErrCodeInternal)
	}
	userID, err := id.NewV7()
	if err != nil {
		return nil, nil, nil, nil, nil, apperr.Wrap(err, "failed to generate user id", apperr.ErrCodeInternal)
	}
	wsID, err := id.NewV7()
	if err != nil {
		return nil, nil, nil, nil, nil, apperr.Wrap(err, "failed to generate workspace id", apperr.ErrCodeInternal)
	}
	identityID, err := id.NewV7()
	if err != nil {
		return nil, nil, nil, nil, nil, apperr.Wrap(err, "failed to generate oauth identity id", apperr.ErrCodeInternal)
	}

	org, err := domain.NewOrganization(orgID, orgName, now)
	if err != nil {
		return nil, nil, nil, nil, nil, mapDomainErr(err)
	}
	user, err := domain.NewOAuthUser(userID, orgID, payload.Email, payload.FullName, now)
	if err != nil {
		return nil, nil, nil, nil, nil, mapDomainErr(err)
	}
	ws, err := domain.NewWorkspace(wsID, orgID, wsName, env, now)
	if err != nil {
		return nil, nil, nil, nil, nil, mapDomainErr(err)
	}
	mem, err := domain.NewMembership(wsID, userID, domain.WorkspaceRoleOwner, now)
	if err != nil {
		return nil, nil, nil, nil, nil, mapDomainErr(err)
	}
	identity, err := domain.NewOAuthIdentity(identityID, userID, provider, payload.Subject, payload.Email, now)
	if err != nil {
		return nil, nil, nil, nil, nil, mapDomainErr(err)
	}

	if err := s.Orgs.Save(ctx, org); err != nil {
		return nil, nil, nil, nil, nil, mapRepoErr(err)
	}
	if err := s.Users.Save(ctx, user); err != nil {
		return nil, nil, nil, nil, nil, mapRepoErr(err)
	}
	if err := s.Workspaces.Save(ctx, ws); err != nil {
		return nil, nil, nil, nil, nil, mapRepoErr(err)
	}
	if err := s.Memberships.Save(ctx, mem); err != nil {
		return nil, nil, nil, nil, nil, mapRepoErr(err)
	}
	if err := s.OAuthIdentities.Save(ctx, identity); err != nil {
		return nil, nil, nil, nil, nil, mapRepoErr(err)
	}

	login, err := s.issueLoginTokens(ctx, tokens, user)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	return login, org, ws, mem, identity, nil
}

type OAuthOnboardingPreview struct {
	Provider string
	Email    string
	FullName string
}

func (s *Services) PeekOAuthOnboarding(token string) (*OAuthOnboardingPreview, error) {
	if s.Cache == nil {
		return nil, apperr.New("onboarding store unavailable", apperr.ErrCodeInternal)
	}
	key := oauthOnboardingCacheKey(token)
	var payload oauthOnboardingPayload
	if err := s.Cache.Get(key, &payload); err != nil {
		return nil, apperr.New("onboarding session expired", apperr.ErrCodeUnauthorized)
	}
	return &OAuthOnboardingPreview{
		Provider: payload.Provider,
		Email:    payload.Email,
		FullName: payload.FullName,
	}, nil
}

func (s *Services) issueLoginTokens(ctx context.Context, tokens auth.TokenService, user *domain.User) (*LoginResult, error) {
	now := s.clock().Now()
	if err := user.RecordLogin(now); err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.Users.Update(ctx, user); err != nil {
		return nil, mapRepoErr(err)
	}
	return s.issueTokens(tokens, user)
}

// issueTokens mints a new access + refresh pair without updating last_login.
func (s *Services) issueTokens(tokens auth.TokenService, user *domain.User) (*LoginResult, error) {
	subject := auth.TokenSubject{
		UserID:          user.ID().String(),
		Email:           user.Email(),
		Role:            "member",
		OrganizationID:  user.OrganizationID().String(),
		IsPlatformAdmin: user.IsPlatformAdmin(),
	}
	access, err := tokens.GenerateAccessToken(subject)
	if err != nil {
		return nil, apperr.Wrap(err, "failed to issue access token", apperr.ErrCodeInternal)
	}
	refresh, err := tokens.GenerateRefreshToken(subject)
	if err != nil {
		return nil, apperr.Wrap(err, "failed to issue refresh token", apperr.ErrCodeInternal)
	}
	if err := tokens.StoreRefreshToken(refresh); err != nil {
		return nil, apperr.Wrap(err, "failed to store refresh token", apperr.ErrCodeInternal)
	}
	return &LoginResult{
		AccessToken:  access,
		RefreshToken: refresh.Token,
		User:         user,
	}, nil
}

func (s *Services) storeOAuthOnboarding(payload oauthOnboardingPayload) (string, error) {
	if s.Cache == nil {
		return "", apperr.New("onboarding store unavailable", apperr.ErrCodeInternal)
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", apperr.Wrap(err, "failed to generate onboarding token", apperr.ErrCodeInternal)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	if err := s.Cache.Set(oauthOnboardingCacheKey(token), payload, oauthOnboardingTTL); err != nil {
		return "", apperr.Wrap(err, "failed to store onboarding token", apperr.ErrCodeInternal)
	}
	return token, nil
}

func (s *Services) consumeOAuthOnboarding(token string) (*oauthOnboardingPayload, error) {
	if s.Cache == nil {
		return nil, apperr.New("onboarding store unavailable", apperr.ErrCodeInternal)
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, apperr.New("onboarding session expired", apperr.ErrCodeUnauthorized)
	}
	key := oauthOnboardingCacheKey(token)
	var payload oauthOnboardingPayload
	if err := s.Cache.Get(key, &payload); err != nil {
		return nil, apperr.New("onboarding session expired", apperr.ErrCodeUnauthorized)
	}
	_ = s.Cache.Delete(key)
	return &payload, nil
}

func oauthOnboardingCacheKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return "oauth_onboarding:" + hex.EncodeToString(sum[:])
}