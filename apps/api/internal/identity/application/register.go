package application

import (
	"context"
	"errors"
	"strings"

	"hublio/internal/identity/domain"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/id"

	"github.com/google/uuid"
)

type RegisterInput struct {
	OrganizationName string
	Email            string
	Password         string
	FullName         string
	WorkspaceName    string
	Environment      string
}

type RegisterResult struct {
	Organization *domain.Organization
	Workspace    *domain.Workspace
	User         *domain.User
	Membership   *domain.Membership
}

// Register creates Organization + default Workspace + owner User in one transaction boundary (caller owns tx).
func (s *Services) Register(ctx context.Context, in RegisterInput) (*RegisterResult, error) {
	if strings.TrimSpace(in.Password) == "" || len(in.Password) < 8 {
		return nil, apperr.New("password must be at least 8 characters", apperr.ErrCodeBadRequest)
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
		return nil, apperr.Wrap(err, "failed to generate organization id", apperr.ErrCodeInternal)
	}
	userID, err := id.NewV7()
	if err != nil {
		return nil, apperr.Wrap(err, "failed to generate user id", apperr.ErrCodeInternal)
	}
	wsID, err := id.NewV7()
	if err != nil {
		return nil, apperr.Wrap(err, "failed to generate workspace id", apperr.ErrCodeInternal)
	}

	hash, err := s.Passwords.Hash(in.Password)
	if err != nil {
		return nil, apperr.Wrap(err, "failed to hash password", apperr.ErrCodeInternal)
	}

	org, err := domain.NewOrganization(orgID, in.OrganizationName, now)
	if err != nil {
		return nil, mapDomainErr(err)
	}
	user, err := domain.NewUser(userID, orgID, in.Email, in.FullName, hash, now)
	if err != nil {
		return nil, mapDomainErr(err)
	}
	ws, err := domain.NewWorkspace(wsID, orgID, wsName, env, now)
	if err != nil {
		return nil, mapDomainErr(err)
	}
	mem, err := domain.NewMembership(wsID, userID, domain.WorkspaceRoleOwner, now)
	if err != nil {
		return nil, mapDomainErr(err)
	}

	if err := s.Orgs.Save(ctx, org); err != nil {
		return nil, mapRepoErr(err)
	}
	if err := s.Users.Save(ctx, user); err != nil {
		return nil, mapRepoErr(err)
	}
	if err := s.Workspaces.Save(ctx, ws); err != nil {
		return nil, mapRepoErr(err)
	}
	if err := s.Memberships.Save(ctx, mem); err != nil {
		return nil, mapRepoErr(err)
	}

	return &RegisterResult{
		Organization: org,
		Workspace:    ws,
		User:         user,
		Membership:   mem,
	}, nil
}

type CreateWorkspaceInput struct {
	OrganizationID uuid.UUID
	ActorUserID    uuid.UUID
	Name           string
	Environment    string
}

func (s *Services) CreateWorkspace(ctx context.Context, in CreateWorkspaceInput) (*domain.Workspace, error) {
	user, err := s.assertOrgAccess(ctx, in.OrganizationID, in.ActorUserID)
	if err != nil {
		return nil, err
	}
	org, err := s.Orgs.FindByID(ctx, in.OrganizationID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if !org.CanSubmitIntents() {
		return nil, apperr.New("organization is not active", apperr.ErrCodeForbidden)
	}

	now := s.clock().Now()
	wsID, err := id.NewV7()
	if err != nil {
		return nil, apperr.Wrap(err, "failed to generate workspace id", apperr.ErrCodeInternal)
	}
	ws, err := domain.NewWorkspace(wsID, org.ID(), in.Name, in.Environment, now)
	if err != nil {
		return nil, mapDomainErr(err)
	}

	if err := s.Workspaces.Save(ctx, ws); err != nil {
		return nil, mapRepoErr(err)
	}

	// Same-org actors become workspace owners. Platform admins creating for
	// another tenant do not join as members (invite later via members API).
	if user.OrganizationID() == org.ID() {
		mem, memErr := domain.NewMembership(wsID, user.ID(), domain.WorkspaceRoleOwner, now)
		if memErr != nil {
			return nil, mapDomainErr(memErr)
		}
		if err := s.Memberships.Save(ctx, mem); err != nil {
			return nil, mapRepoErr(err)
		}
	}
	return ws, nil
}

type CreateOrganizationInput struct {
	ActorUserID     uuid.UUID
	Name            string
	WorkspaceName   string
	Environment     string
}

type CreateOrganizationResult struct {
	Organization *domain.Organization
	Workspace    *domain.Workspace
}

// CreateOrganization creates a tenant org + default workspace (platform admin only).
// No owner user is created — ops invite members later.
func (s *Services) CreateOrganization(ctx context.Context, in CreateOrganizationInput) (*CreateOrganizationResult, error) {
	user, err := s.Users.FindByID(ctx, in.ActorUserID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if !user.IsPlatformAdmin() {
		return nil, apperr.New("forbidden", apperr.ErrCodeForbidden)
	}

	wsName := strings.TrimSpace(in.WorkspaceName)
	if wsName == "" {
		wsName = "default"
	}
	env := strings.TrimSpace(in.Environment)
	if env == "" {
		env = "production"
	}

	now := s.clock().Now()
	orgID, err := id.NewV7()
	if err != nil {
		return nil, apperr.Wrap(err, "failed to generate organization id", apperr.ErrCodeInternal)
	}
	wsID, err := id.NewV7()
	if err != nil {
		return nil, apperr.Wrap(err, "failed to generate workspace id", apperr.ErrCodeInternal)
	}

	org, err := domain.NewOrganization(orgID, in.Name, now)
	if err != nil {
		return nil, mapDomainErr(err)
	}
	ws, err := domain.NewWorkspace(wsID, orgID, wsName, env, now)
	if err != nil {
		return nil, mapDomainErr(err)
	}

	if err := s.Orgs.Save(ctx, org); err != nil {
		return nil, mapRepoErr(err)
	}
	if err := s.Workspaces.Save(ctx, ws); err != nil {
		return nil, mapRepoErr(err)
	}

	return &CreateOrganizationResult{Organization: org, Workspace: ws}, nil
}

type AddMemberInput struct {
	WorkspaceID uuid.UUID
	ActorUserID uuid.UUID
	Email       string
	Role        domain.WorkspaceRole
}

func (s *Services) AddUserToWorkspace(ctx context.Context, in AddMemberInput) (*domain.Membership, error) {
	ws, err := s.Workspaces.FindByID(ctx, in.WorkspaceID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	actor, err := s.Users.FindByID(ctx, in.ActorUserID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	actorMem, err := s.Memberships.Find(ctx, ws.ID(), actor.ID())
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if actorMem.Role() != domain.WorkspaceRoleOwner && actorMem.Role() != domain.WorkspaceRoleAdmin {
		return nil, apperr.New("insufficient workspace role", apperr.ErrCodeForbidden)
	}

	user, err := s.Users.FindByEmail(ctx, strings.TrimSpace(strings.ToLower(in.Email)))
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if user.OrganizationID() != ws.OrganizationID() {
		return nil, apperr.New("user must belong to the same organization", apperr.ErrCodeForbidden)
	}

	if _, err := s.Memberships.Find(ctx, ws.ID(), user.ID()); err == nil {
		return nil, apperr.New("user already a workspace member", apperr.ErrCodeConflict)
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, mapRepoErr(err)
	}

	mem, err := domain.NewMembership(ws.ID(), user.ID(), in.Role, s.clock().Now())
	if err != nil {
		return nil, mapDomainErr(err)
	}
	if err := s.Memberships.Save(ctx, mem); err != nil {
		return nil, mapRepoErr(err)
	}
	return mem, nil
}

func mapDomainErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidName),
		errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrInvalidPassword),
		errors.Is(err, domain.ErrInvalidEnvironment),
		errors.Is(err, domain.ErrInvalidRole),
		errors.Is(err, domain.ErrInvalidOAuthProvider),
		errors.Is(err, domain.ErrInvalidOAuthIdentity),
		errors.Is(err, domain.ErrInvalidMFAConfig):
		return apperr.Wrap(err, err.Error(), apperr.ErrCodeBadRequest)
	case errors.Is(err, domain.ErrInvalidMFACode):
		return apperr.Wrap(err, "invalid mfa code", apperr.ErrCodeUnauthorized)
	case errors.Is(err, domain.ErrOAuthEmailUnverified),
		errors.Is(err, domain.ErrOAuthPlatformAdmin):
		return apperr.Wrap(err, err.Error(), apperr.ErrCodeForbidden)
	case errors.Is(err, domain.ErrInvalidTransition),
		errors.Is(err, domain.ErrOrganizationBlocked),
		errors.Is(err, domain.ErrWorkspaceDisabled),
		errors.Is(err, domain.ErrUserCannotLogin),
		errors.Is(err, domain.ErrAPIKeyDisabled),
		errors.Is(err, domain.ErrAPIKeyExpired),
		errors.Is(err, domain.ErrMFAAlreadyEnabled),
		errors.Is(err, domain.ErrMFANotEnabled):
		return apperr.Wrap(err, err.Error(), apperr.ErrCodeConflict)
	default:
		return apperr.Wrap(err, "domain error", apperr.ErrCodeBadRequest)
	}
}

func mapRepoErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, domain.ErrNotFound) {
		return apperr.New("resource not found", apperr.ErrCodeNotFound)
	}
	if errors.Is(err, domain.ErrConflict) {
		return apperr.New("resource already exists", apperr.ErrCodeConflict)
	}
	if ae, ok := err.(*apperr.AppError); ok {
		return ae
	}
	return apperr.Wrap(err, "persistence error", apperr.ErrCodeInternal)
}
