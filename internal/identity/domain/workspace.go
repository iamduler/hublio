package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type WorkspaceStatus string

const (
	WorkspaceStatusActive   WorkspaceStatus = "active"
	WorkspaceStatusDisabled WorkspaceStatus = "disabled"
)

// Workspace isolates operations inside an Organization and owns API keys.
type Workspace struct {
	eventRecorder

	id             uuid.UUID
	organizationID uuid.UUID
	name           string
	environment    string
	status         WorkspaceStatus
	createdAt      time.Time
	updatedAt      time.Time
	deletedAt      *time.Time
}

func NewWorkspace(id, organizationID uuid.UUID, name, environment string, now time.Time) (*Workspace, error) {
	name = strings.TrimSpace(name)
	environment = strings.TrimSpace(environment)
	if name == "" || len(name) > 150 {
		return nil, ErrInvalidName
	}
	if environment == "" || len(environment) > 50 {
		return nil, ErrInvalidEnvironment
	}
	if id == uuid.Nil || organizationID == uuid.Nil {
		return nil, ErrInvalidName
	}

	ws := &Workspace{
		id:             id,
		organizationID: organizationID,
		name:           name,
		environment:    environment,
		status:         WorkspaceStatusActive,
		createdAt:      now.UTC(),
		updatedAt:      now.UTC(),
	}
	ws.record(EventWorkspaceCreated, id, now.UTC(), map[string]any{
		"organization_id": organizationID.String(),
		"name":            name,
		"environment":     environment,
	})
	return ws, nil
}

func ReconstituteWorkspace(
	id, organizationID uuid.UUID,
	name, environment string,
	status WorkspaceStatus,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
) *Workspace {
	return &Workspace{
		id:             id,
		organizationID: organizationID,
		name:           name,
		environment:    environment,
		status:         status,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
		deletedAt:      deletedAt,
	}
}

func (w *Workspace) ID() uuid.UUID              { return w.id }
func (w *Workspace) OrganizationID() uuid.UUID  { return w.organizationID }
func (w *Workspace) Name() string               { return w.name }
func (w *Workspace) Environment() string        { return w.environment }
func (w *Workspace) Status() WorkspaceStatus    { return w.status }
func (w *Workspace) CreatedAt() time.Time       { return w.createdAt }
func (w *Workspace) UpdatedAt() time.Time       { return w.updatedAt }
func (w *Workspace) DeletedAt() *time.Time      { return w.deletedAt }

func (w *Workspace) CanExecuteIntents() bool {
	return w.status == WorkspaceStatusActive && w.deletedAt == nil
}

func (w *Workspace) Update(name, environment string, now time.Time) error {
	if !w.CanExecuteIntents() && w.status == WorkspaceStatusDisabled {
		// Disabled workspaces may still be renamed before re-enable; archived/deleted cannot.
	}
	if w.deletedAt != nil {
		return ErrWorkspaceDisabled
	}
	name = strings.TrimSpace(name)
	environment = strings.TrimSpace(environment)
	if name == "" || len(name) > 150 {
		return ErrInvalidName
	}
	if environment == "" || len(environment) > 50 {
		return ErrInvalidEnvironment
	}
	w.name = name
	w.environment = environment
	w.updatedAt = now.UTC()
	w.record(EventWorkspaceUpdated, w.id, w.updatedAt, map[string]any{"name": name, "environment": environment})
	return nil
}

func (w *Workspace) Disable(now time.Time) error {
	if w.status != WorkspaceStatusActive {
		return ErrInvalidTransition
	}
	w.status = WorkspaceStatusDisabled
	w.updatedAt = now.UTC()
	w.record(EventWorkspaceDisabled, w.id, w.updatedAt, nil)
	return nil
}

func (w *Workspace) Enable(now time.Time) error {
	if w.status != WorkspaceStatusDisabled {
		return ErrInvalidTransition
	}
	if w.deletedAt != nil {
		return ErrWorkspaceDisabled
	}
	w.status = WorkspaceStatusActive
	w.updatedAt = now.UTC()
	w.record(EventWorkspaceEnabled, w.id, w.updatedAt, nil)
	return nil
}
