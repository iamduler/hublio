package domain

import (
	"context"

	"github.com/google/uuid"
)

type ConnectorRepository interface {
	Save(ctx context.Context, connector *Connector) error
	Update(ctx context.Context, connector *Connector) error
	FindByID(ctx context.Context, id uuid.UUID) (*Connector, error)
	FindByCode(ctx context.Context, code string) (*Connector, error)
	List(ctx context.Context) ([]*Connector, error)
}

type ConnectionRepository interface {
	Save(ctx context.Context, conn *Connection) error
	Update(ctx context.Context, conn *Connection) error
	FindByID(ctx context.Context, id uuid.UUID) (*Connection, error)
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*Connection, error)
}

type CredentialRepository interface {
	Save(ctx context.Context, cred *Credential) error
	Update(ctx context.Context, cred *Credential) error
	FindByID(ctx context.Context, id uuid.UUID) (*Credential, error)
	FindActiveByConnection(ctx context.Context, connectionID uuid.UUID) (*Credential, error)
	ListByConnection(ctx context.Context, connectionID uuid.UUID) ([]*Credential, error)
}
