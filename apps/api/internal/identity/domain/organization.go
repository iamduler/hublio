package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type OrganizationStatus string

const (
	OrganizationStatusActive    OrganizationStatus = "active"
	OrganizationStatusSuspended OrganizationStatus = "suspended"
	OrganizationStatusArchived  OrganizationStatus = "archived"
)

// Organization is the tenant root aggregate.
type Organization struct {
	eventRecorder

	id        uuid.UUID
	name      string
	status    OrganizationStatus
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
}

func NewOrganization(id uuid.UUID, name string, now time.Time) (*Organization, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 255 {
		return nil, ErrInvalidName
	}
	if id == uuid.Nil {
		return nil, ErrInvalidName
	}

	org := &Organization{
		id:        id,
		name:      name,
		status:    OrganizationStatusActive,
		createdAt: now.UTC(),
		updatedAt: now.UTC(),
	}
	org.record(EventOrganizationCreated, id, now.UTC(), map[string]any{"name": name})
	return org, nil
}

func ReconstituteOrganization(
	id uuid.UUID,
	name string,
	status OrganizationStatus,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
) *Organization {
	return &Organization{
		id:        id,
		name:      name,
		status:    status,
		createdAt: createdAt,
		updatedAt: updatedAt,
		deletedAt: deletedAt,
	}
}

func (o *Organization) ID() uuid.UUID                 { return o.id }
func (o *Organization) Name() string                  { return o.name }
func (o *Organization) Status() OrganizationStatus    { return o.status }
func (o *Organization) CreatedAt() time.Time          { return o.createdAt }
func (o *Organization) UpdatedAt() time.Time          { return o.updatedAt }
func (o *Organization) DeletedAt() *time.Time         { return o.deletedAt }

// CanSubmitIntents reports whether new Intents may be accepted for this tenant.
func (o *Organization) CanSubmitIntents() bool {
	return o.status == OrganizationStatusActive && o.deletedAt == nil
}

func (o *Organization) Update(name string, now time.Time) error {
	if !o.isMutable() {
		return ErrOrganizationBlocked
	}
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 255 {
		return ErrInvalidName
	}
	o.name = name
	o.updatedAt = now.UTC()
	o.record(EventOrganizationUpdated, o.id, o.updatedAt, map[string]any{"name": name})
	return nil
}

func (o *Organization) Suspend(now time.Time) error {
	if o.status != OrganizationStatusActive {
		return ErrInvalidTransition
	}
	o.status = OrganizationStatusSuspended
	o.updatedAt = now.UTC()
	o.record(EventOrganizationSuspended, o.id, o.updatedAt, nil)
	return nil
}

func (o *Organization) Activate(now time.Time) error {
	if o.status != OrganizationStatusSuspended {
		return ErrInvalidTransition
	}
	o.status = OrganizationStatusActive
	o.updatedAt = now.UTC()
	o.record(EventOrganizationActivated, o.id, o.updatedAt, nil)
	return nil
}

func (o *Organization) Archive(now time.Time) error {
	if o.status == OrganizationStatusArchived {
		return ErrInvalidTransition
	}
	at := now.UTC()
	o.status = OrganizationStatusArchived
	o.deletedAt = &at
	o.updatedAt = at
	o.record(EventOrganizationArchived, o.id, at, nil)
	return nil
}

func (o *Organization) isMutable() bool {
	return o.status != OrganizationStatusArchived && o.deletedAt == nil
}
