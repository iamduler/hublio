package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type ConnectionStatus string

const (
	ConnectionStatusDraft              ConnectionStatus = "draft"
	ConnectionStatusVerifying          ConnectionStatus = "verifying"
	ConnectionStatusActive             ConnectionStatus = "active"
	ConnectionStatusVerificationFailed ConnectionStatus = "verification_failed"
	ConnectionStatusDisabled           ConnectionStatus = "disabled"
)

// Connection is a Workspace-scoped, Connector-referencing configuration aggregate.
// States: Draft -> Verifying -> Active | VerificationFailed; VerificationFailed -> Verifying;
// Active <-> Disabled.
type Connection struct {
	eventRecorder

	id                 uuid.UUID
	workspaceID        uuid.UUID
	connectorID        uuid.UUID
	name               string
	isDefault          bool
	description        string
	environment        string
	status             ConnectionStatus
	config             map[string]any
	retryPolicy        map[string]any
	timeoutSeconds     int
	activeCredentialID *uuid.UUID
	createdAt          time.Time
	updatedAt          time.Time
	deletedAt          *time.Time
}

func NewConnection(
	id, workspaceID, connectorID uuid.UUID,
	name string,
	isDefault bool,
	description, environment string,
	config, retryPolicy map[string]any,
	timeoutSeconds int,
	now time.Time,
) (*Connection, error) {
	name = strings.TrimSpace(name)
	environment = strings.TrimSpace(environment)

	if id == uuid.Nil || workspaceID == uuid.Nil || connectorID == uuid.Nil {
		return nil, ErrInvalidName
	}
	if name == "" || len(name) > 255 {
		return nil, ErrInvalidName
	}
	if environment == "" || len(environment) > 50 {
		return nil, ErrInvalidEnvironment
	}
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}

	conn := &Connection{
		id:             id,
		workspaceID:    workspaceID,
		connectorID:    connectorID,
		name:           name,
		isDefault:      isDefault,
		description:    strings.TrimSpace(description),
		environment:    environment,
		status:         ConnectionStatusDraft,
		config:         config,
		retryPolicy:    retryPolicy,
		timeoutSeconds: timeoutSeconds,
		createdAt:      now.UTC(),
		updatedAt:      now.UTC(),
	}
	conn.record(EventConnectionCreated, id, now.UTC(), map[string]any{
		"workspace_id": workspaceID.String(),
		"connector_id": connectorID.String(),
		"name":         name,
	})
	return conn, nil
}

func ReconstituteConnection(
	id, workspaceID, connectorID uuid.UUID,
	name string,
	isDefault bool,
	description, environment string,
	status ConnectionStatus,
	config, retryPolicy map[string]any,
	timeoutSeconds int,
	activeCredentialID *uuid.UUID,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
) *Connection {
	return &Connection{
		id:                 id,
		workspaceID:        workspaceID,
		connectorID:        connectorID,
		name:               name,
		isDefault:          isDefault,
		description:        description,
		environment:        environment,
		status:             status,
		config:             config,
		retryPolicy:        retryPolicy,
		timeoutSeconds:     timeoutSeconds,
		activeCredentialID: activeCredentialID,
		createdAt:          createdAt,
		updatedAt:          updatedAt,
		deletedAt:          deletedAt,
	}
}

func (c *Connection) ID() uuid.UUID                  { return c.id }
func (c *Connection) WorkspaceID() uuid.UUID         { return c.workspaceID }
func (c *Connection) ConnectorID() uuid.UUID         { return c.connectorID }
func (c *Connection) Name() string                   { return c.name }
func (c *Connection) IsDefault() bool                { return c.isDefault }
func (c *Connection) Description() string            { return c.description }
func (c *Connection) Environment() string            { return c.environment }
func (c *Connection) Status() ConnectionStatus       { return c.status }
func (c *Connection) Config() map[string]any         { return c.config }
func (c *Connection) RetryPolicy() map[string]any    { return c.retryPolicy }
func (c *Connection) TimeoutSeconds() int            { return c.timeoutSeconds }
func (c *Connection) ActiveCredentialID() *uuid.UUID { return c.activeCredentialID }
func (c *Connection) CreatedAt() time.Time           { return c.createdAt }
func (c *Connection) UpdatedAt() time.Time           { return c.updatedAt }
func (c *Connection) DeletedAt() *time.Time          { return c.deletedAt }

// CanExecuteIntents reports whether this Connection may be used by new Intents.
func (c *Connection) CanExecuteIntents() bool {
	return c.status == ConnectionStatusActive && c.deletedAt == nil
}

func (c *Connection) SetActiveCredential(credentialID uuid.UUID, now time.Time) {
	c.activeCredentialID = &credentialID
	c.updatedAt = now.UTC()
}

func (c *Connection) Update(name, description string, config, retryPolicy map[string]any, timeoutSeconds int, now time.Time) error {
	if c.deletedAt != nil {
		return ErrInvalidTransition
	}
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 255 {
		return ErrInvalidName
	}
	if timeoutSeconds <= 0 {
		timeoutSeconds = c.timeoutSeconds
	}
	c.name = name
	c.description = strings.TrimSpace(description)
	c.config = config
	c.retryPolicy = retryPolicy
	c.timeoutSeconds = timeoutSeconds
	c.updatedAt = now.UTC()
	return nil
}

func (c *Connection) StartVerify(now time.Time) error {
	if c.status != ConnectionStatusDraft && c.status != ConnectionStatusVerificationFailed {
		return ErrInvalidTransition
	}
	c.status = ConnectionStatusVerifying
	c.updatedAt = now.UTC()
	c.record(EventConnectionVerifying, c.id, now.UTC(), nil)
	return nil
}

func (c *Connection) MarkVerified(now time.Time) error {
	if c.status != ConnectionStatusVerifying {
		return ErrInvalidTransition
	}
	c.status = ConnectionStatusActive
	c.updatedAt = now.UTC()
	c.record(EventConnectionVerified, c.id, now.UTC(), nil)
	return nil
}

func (c *Connection) MarkVerificationFailed(reason string, now time.Time) error {
	if c.status != ConnectionStatusVerifying {
		return ErrInvalidTransition
	}
	c.status = ConnectionStatusVerificationFailed
	c.updatedAt = now.UTC()
	c.record(EventConnectionVerificationFailed, c.id, now.UTC(), map[string]any{"reason": reason})
	return nil
}

func (c *Connection) Disable(now time.Time) error {
	if c.status != ConnectionStatusActive {
		return ErrInvalidTransition
	}
	c.status = ConnectionStatusDisabled
	c.updatedAt = now.UTC()
	c.record(EventConnectionDisabled, c.id, now.UTC(), nil)
	return nil
}

// Enable re-activates a Disabled Connection that was previously verified.
func (c *Connection) Enable(now time.Time) error {
	if c.status != ConnectionStatusDisabled {
		return ErrInvalidTransition
	}
	c.status = ConnectionStatusActive
	c.updatedAt = now.UTC()
	c.record(EventConnectionEnabled, c.id, now.UTC(), nil)
	return nil
}
