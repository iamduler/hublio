package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PasswordHasher is implemented in Infrastructure (bcrypt/argon2).
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type OrganizationRepository interface {
	Save(ctx context.Context, org *Organization) error
	Update(ctx context.Context, org *Organization) error
	FindByID(ctx context.Context, id uuid.UUID) (*Organization, error)
	FindByName(ctx context.Context, name string) (*Organization, error)
}

type WorkspaceRepository interface {
	Save(ctx context.Context, ws *Workspace) error
	Update(ctx context.Context, ws *Workspace) error
	FindByID(ctx context.Context, id uuid.UUID) (*Workspace, error)
	ListByOrganization(ctx context.Context, organizationID uuid.UUID) ([]*Workspace, error)
}

type UserRepository interface {
	Save(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
}

type MembershipRepository interface {
	Save(ctx context.Context, membership *Membership) error
	Find(ctx context.Context, workspaceID, userID uuid.UUID) (*Membership, error)
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*Membership, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Membership, error)
}

type APIKeyRepository interface {
	Save(ctx context.Context, key *APIKey) error
	Update(ctx context.Context, key *APIKey) error
	FindByID(ctx context.Context, id uuid.UUID) (*APIKey, error)
	FindByPrefix(ctx context.Context, prefix string) (*APIKey, error)
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*APIKey, error)
	TouchLastUsed(ctx context.Context, id uuid.UUID, at time.Time) error
}
