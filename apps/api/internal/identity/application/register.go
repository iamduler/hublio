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
	org, err := s.Orgs.FindByID(ctx, in.OrganizationID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if !org.CanSubmitIntents() {
		return nil, apperr.New("organization is not active", apperr.ErrCodeForbidden)
	}

	user, err := s.Users.FindByID(ctx, in.ActorUserID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if user.OrganizationID() != org.ID() {
		return nil, apperr.New("user does not belong to organization", apperr.ErrCodeForbidden)
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
	mem, err := domain.NewMembership(wsID, user.ID(), domain.WorkspaceRoleOwner, now)
	if err != nil {
		return nil, mapDomainErr(err)
	}

	if err := s.Workspaces.Save(ctx, ws); err != nil {
		return nil, mapRepoErr(err)
	}
	if err := s.Memberships.Save(ctx, mem); err != nil {
		return nil, mapRepoErr(err)
	}
	return ws, nil
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
		errors.Is(err, domain.ErrInvalidRole):
		return apperr.Wrap(err, err.Error(), apperr.ErrCodeBadRequest)
	case errors.Is(err, domain.ErrInvalidTransition),
		errors.Is(err, domain.ErrOrganizationBlocked),
		errors.Is(err, domain.ErrWorkspaceDisabled),
		errors.Is(err, domain.ErrUserCannotLogin),
		errors.Is(err, domain.ErrAPIKeyDisabled),
		errors.Is(err, domain.ErrAPIKeyExpired):
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
