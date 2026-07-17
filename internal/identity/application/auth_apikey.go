package application

import (
	"context"
	"strings"
	"time"

	"hublio/internal/identity/domain"
	"hublio/internal/platform/apikey"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/auth"
	"hublio/internal/platform/id"

	"github.com/google/uuid"
)

type LoginInput struct {
	Email    string
	Password string
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	User         *domain.User
}

func (s *Services) Login(ctx context.Context, tokens auth.TokenService, in LoginInput) (*LoginResult, error) {
	user, err := s.Users.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(in.Email)))
	if err != nil {
		return nil, apperr.New("invalid email or password", apperr.ErrCodeUnauthorized)
	}
	if !user.CanLogin() {
		return nil, apperr.New("invalid email or password", apperr.ErrCodeUnauthorized)
	}
	if err := s.Passwords.Compare(user.PasswordHash(), in.Password); err != nil {
		return nil, apperr.New("invalid email or password", apperr.ErrCodeUnauthorized)
	}

	now := s.clock().Now()
	if err := user.RecordLogin(now); err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.Users.Update(ctx, user); err != nil {
		return nil, mapRepoErr(err)
	}

	subject := auth.TokenSubject{
		UserID:         user.ID().String(),
		Email:          user.Email(),
		Role:           "member",
		OrganizationID: user.OrganizationID().String(),
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

func (s *Services) Logout(ctx context.Context, tokens auth.TokenService, refreshToken string) error {
	_ = ctx
	return tokens.RevokeRefreshToken(refreshToken)
}

type CreateAPIKeyInput struct {
	WorkspaceID uuid.UUID
	ActorUserID uuid.UUID
	Name        string
	ExpiresAt   *time.Time
}

type CreateAPIKeyResult struct {
	APIKey    *domain.APIKey
	Plaintext string
}

func (s *Services) CreateAPIKey(ctx context.Context, in CreateAPIKeyInput) (*CreateAPIKeyResult, error) {
	if err := s.assertWorkspaceMember(ctx, in.WorkspaceID, in.ActorUserID); err != nil {
		return nil, err
	}
	ws, err := s.Workspaces.FindByID(ctx, in.WorkspaceID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if !ws.CanExecuteIntents() {
		return nil, apperr.New("workspace is disabled", apperr.ErrCodeForbidden)
	}

	generated, err := apikey.Generate()
	if err != nil {
		return nil, apperr.Wrap(err, "failed to generate api key", apperr.ErrCodeInternal)
	}
	keyID, err := id.NewV7()
	if err != nil {
		return nil, apperr.Wrap(err, "failed to generate api key id", apperr.ErrCodeInternal)
	}

	now := s.clock().Now()
	key, err := domain.NewAPIKey(keyID, ws.ID(), in.Name, generated.Hash, generated.Prefix, in.ExpiresAt, now)
	if err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.APIKeys.Save(ctx, key); err != nil {
		return nil, mapRepoErr(err)
	}
	return &CreateAPIKeyResult{APIKey: key, Plaintext: generated.Plaintext}, nil
}

func (s *Services) DisableAPIKey(ctx context.Context, workspaceID, keyID, actorUserID uuid.UUID) (*domain.APIKey, error) {
	if err := s.assertWorkspaceMember(ctx, workspaceID, actorUserID); err != nil {
		return nil, err
	}
	key, err := s.APIKeys.FindByID(ctx, keyID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if key.WorkspaceID() != workspaceID {
		return nil, apperr.New("api key not found", apperr.ErrCodeNotFound)
	}
	if err := key.Disable(s.clock().Now()); err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.APIKeys.Update(ctx, key); err != nil {
		return nil, mapRepoErr(err)
	}
	return key, nil
}

func (s *Services) RotateAPIKey(ctx context.Context, workspaceID, keyID, actorUserID uuid.UUID) (*CreateAPIKeyResult, error) {
	if err := s.assertWorkspaceMember(ctx, workspaceID, actorUserID); err != nil {
		return nil, err
	}
	key, err := s.APIKeys.FindByID(ctx, keyID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if key.WorkspaceID() != workspaceID {
		return nil, apperr.New("api key not found", apperr.ErrCodeNotFound)
	}
	generated, err := apikey.Generate()
	if err != nil {
		return nil, apperr.Wrap(err, "failed to generate api key", apperr.ErrCodeInternal)
	}
	if err := key.Rotate(generated.Hash, generated.Prefix, s.clock().Now()); err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.APIKeys.Update(ctx, key); err != nil {
		return nil, mapRepoErr(err)
	}
	return &CreateAPIKeyResult{APIKey: key, Plaintext: generated.Plaintext}, nil
}

func (s *Services) ListAPIKeys(ctx context.Context, workspaceID, actorUserID uuid.UUID) ([]*domain.APIKey, error) {
	if err := s.assertWorkspaceMember(ctx, workspaceID, actorUserID); err != nil {
		return nil, err
	}
	keys, err := s.APIKeys.ListByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return keys, nil
}

func (s *Services) GetOrganization(ctx context.Context, organizationID uuid.UUID) (*domain.Organization, error) {
	org, err := s.Orgs.FindByID(ctx, organizationID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return org, nil
}

func (s *Services) ListWorkspaces(ctx context.Context, organizationID, actorUserID uuid.UUID) ([]*domain.Workspace, error) {
	user, err := s.Users.FindByID(ctx, actorUserID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if user.OrganizationID() != organizationID {
		return nil, apperr.New("forbidden", apperr.ErrCodeForbidden)
	}
	list, err := s.Workspaces.ListByOrganization(ctx, organizationID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	return list, nil
}

func (s *Services) SuspendOrganization(ctx context.Context, organizationID, actorUserID uuid.UUID) (*domain.Organization, error) {
	return s.changeOrganization(ctx, organizationID, actorUserID, func(o *domain.Organization, now time.Time) error {
		return o.Suspend(now)
	})
}

func (s *Services) ActivateOrganization(ctx context.Context, organizationID, actorUserID uuid.UUID) (*domain.Organization, error) {
	return s.changeOrganization(ctx, organizationID, actorUserID, func(o *domain.Organization, now time.Time) error {
		return o.Activate(now)
	})
}

func (s *Services) changeOrganization(
	ctx context.Context,
	organizationID, actorUserID uuid.UUID,
	fn func(*domain.Organization, time.Time) error,
) (*domain.Organization, error) {
	user, err := s.Users.FindByID(ctx, actorUserID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if user.OrganizationID() != organizationID {
		return nil, apperr.New("forbidden", apperr.ErrCodeForbidden)
	}
	org, err := s.Orgs.FindByID(ctx, organizationID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	now := s.clock().Now()
	if err := fn(org, now); err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.Orgs.Update(ctx, org); err != nil {
		return nil, mapRepoErr(err)
	}
	return org, nil
}

func (s *Services) SetWorkspaceStatus(ctx context.Context, workspaceID, actorUserID uuid.UUID, enable bool) (*domain.Workspace, error) {
	if err := s.assertWorkspaceMember(ctx, workspaceID, actorUserID); err != nil {
		return nil, err
	}
	ws, err := s.Workspaces.FindByID(ctx, workspaceID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	now := s.clock().Now()
	if enable {
		err = ws.Enable(now)
	} else {
		err = ws.Disable(now)
	}
	if err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.Workspaces.Update(ctx, ws); err != nil {
		return nil, mapRepoErr(err)
	}
	return ws, nil
}

func (s *Services) assertWorkspaceMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
	_, err := s.Memberships.Find(ctx, workspaceID, userID)
	if err != nil {
		return mapRepoErr(err)
	}
	return nil
}
