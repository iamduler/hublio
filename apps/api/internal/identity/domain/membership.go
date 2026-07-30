package domain

import (
	"time"

	"github.com/google/uuid"
)

type WorkspaceRole string

const (
	WorkspaceRoleOwner  WorkspaceRole = "owner"
	WorkspaceRoleAdmin  WorkspaceRole = "admin"
	WorkspaceRoleMember WorkspaceRole = "member"
)

func ParseWorkspaceRole(role string) (WorkspaceRole, error) {
	switch WorkspaceRole(role) {
	case WorkspaceRoleOwner, WorkspaceRoleAdmin, WorkspaceRoleMember:
		return WorkspaceRole(role), nil
	default:
		return "", ErrInvalidRole
	}
}

// Membership links a User to a Workspace with a role.
type Membership struct {
	eventRecorder

	workspaceID uuid.UUID
	userID      uuid.UUID
	role        WorkspaceRole
	createdAt   time.Time
}

func NewMembership(workspaceID, userID uuid.UUID, role WorkspaceRole, now time.Time) (*Membership, error) {
	if workspaceID == uuid.Nil || userID == uuid.Nil {
		return nil, ErrInvalidRole
	}
	if _, err := ParseWorkspaceRole(string(role)); err != nil {
		return nil, err
	}
	m := &Membership{
		workspaceID: workspaceID,
		userID:      userID,
		role:        role,
		createdAt:   now.UTC(),
	}
	m.record(EventMembershipAdded, userID, now.UTC(), map[string]any{
		"workspace_id": workspaceID.String(),
		"role":         string(role),
	})
	return m, nil
}

func ReconstituteMembership(workspaceID, userID uuid.UUID, role WorkspaceRole, createdAt time.Time) *Membership {
	return &Membership{
		workspaceID: workspaceID,
		userID:      userID,
		role:        role,
		createdAt:   createdAt,
	}
}

func (m *Membership) WorkspaceID() uuid.UUID { return m.workspaceID }
func (m *Membership) UserID() uuid.UUID      { return m.userID }
func (m *Membership) Role() WorkspaceRole    { return m.role }
func (m *Membership) CreatedAt() time.Time   { return m.createdAt }

// WorkspaceMember is a read model for listing members with user profile fields.
type WorkspaceMember struct {
	workspaceID uuid.UUID
	userID      uuid.UUID
	email       string
	fullName    string
	role        WorkspaceRole
	createdAt   time.Time
}

func ReconstituteWorkspaceMember(
	workspaceID, userID uuid.UUID,
	email, fullName string,
	role WorkspaceRole,
	createdAt time.Time,
) *WorkspaceMember {
	return &WorkspaceMember{
		workspaceID: workspaceID,
		userID:      userID,
		email:       email,
		fullName:    fullName,
		role:        role,
		createdAt:   createdAt,
	}
}

func (m *WorkspaceMember) WorkspaceID() uuid.UUID { return m.workspaceID }
func (m *WorkspaceMember) UserID() uuid.UUID      { return m.userID }
func (m *WorkspaceMember) Email() string          { return m.email }
func (m *WorkspaceMember) FullName() string       { return m.fullName }
func (m *WorkspaceMember) Role() WorkspaceRole    { return m.role }
func (m *WorkspaceMember) CreatedAt() time.Time   { return m.createdAt }
