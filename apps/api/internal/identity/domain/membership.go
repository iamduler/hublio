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
