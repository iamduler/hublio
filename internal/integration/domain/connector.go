package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type ConnectorStatus string

const (
	ConnectorStatusRegistered ConnectorStatus = "registered"
	ConnectorStatusEnabled    ConnectorStatus = "enabled"
	ConnectorStatusDisabled   ConnectorStatus = "disabled"
	ConnectorStatusRemoved    ConnectorStatus = "removed"
)

type ConnectorCategory string

const (
	ConnectorCategorySource        ConnectorCategory = "source"
	ConnectorCategoryDestination   ConnectorCategory = "destination"
	ConnectorCategoryBidirectional ConnectorCategory = "bidirectional"
	ConnectorCategoryUtility       ConnectorCategory = "utility"
)

func ParseConnectorCategory(v string) (ConnectorCategory, error) {
	switch ConnectorCategory(v) {
	case ConnectorCategorySource, ConnectorCategoryDestination, ConnectorCategoryBidirectional, ConnectorCategoryUtility:
		return ConnectorCategory(v), nil
	default:
		return "", ErrInvalidCategory
	}
}

type CapabilityStatus string

const (
	CapabilityStatusEnabled  CapabilityStatus = "enabled"
	CapabilityStatusDisabled CapabilityStatus = "disabled"
)

// Capability describes one operation a Connector exposes (child of Connector).
type Capability struct {
	id          uuid.UUID
	connectorID uuid.UUID
	code        string
	displayName string
	status      CapabilityStatus
	isAsync     bool
	createdAt   time.Time
	updatedAt   time.Time
}

func newCapability(id, connectorID uuid.UUID, code, displayName string, isAsync bool, now time.Time) (*Capability, error) {
	code = strings.TrimSpace(code)
	displayName = strings.TrimSpace(displayName)
	if id == uuid.Nil || connectorID == uuid.Nil {
		return nil, ErrInvalidCode
	}
	if code == "" || len(code) > 150 {
		return nil, ErrInvalidCode
	}
	if displayName == "" || len(displayName) > 255 {
		return nil, ErrInvalidName
	}
	return &Capability{
		id:          id,
		connectorID: connectorID,
		code:        code,
		displayName: displayName,
		status:      CapabilityStatusEnabled,
		isAsync:     isAsync,
		createdAt:   now.UTC(),
		updatedAt:   now.UTC(),
	}, nil
}

func ReconstituteCapability(
	id, connectorID uuid.UUID,
	code, displayName string,
	status CapabilityStatus,
	isAsync bool,
	createdAt, updatedAt time.Time,
) *Capability {
	return &Capability{
		id:          id,
		connectorID: connectorID,
		code:        code,
		displayName: displayName,
		status:      status,
		isAsync:     isAsync,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

func (k *Capability) ID() uuid.UUID            { return k.id }
func (k *Capability) ConnectorID() uuid.UUID   { return k.connectorID }
func (k *Capability) Code() string             { return k.code }
func (k *Capability) DisplayName() string      { return k.displayName }
func (k *Capability) Status() CapabilityStatus { return k.status }
func (k *Capability) IsAsync() bool            { return k.isAsync }
func (k *Capability) CreatedAt() time.Time     { return k.createdAt }
func (k *Capability) UpdatedAt() time.Time     { return k.updatedAt }
func (k *Capability) IsEnabled() bool          { return k.status == CapabilityStatusEnabled }

func (k *Capability) Enable(now time.Time) error {
	if k.status == CapabilityStatusEnabled {
		return ErrInvalidTransition
	}
	k.status = CapabilityStatusEnabled
	k.updatedAt = now.UTC()
	return nil
}

func (k *Capability) Disable(now time.Time) error {
	if k.status == CapabilityStatusDisabled {
		return ErrInvalidTransition
	}
	k.status = CapabilityStatusDisabled
	k.updatedAt = now.UTC()
	return nil
}

// Connector is the installed integration package aggregate.
// States: Registered -> Enabled <-> Disabled -> Removed (terminal).
type Connector struct {
	eventRecorder

	id               uuid.UUID
	code             string
	name             string
	vendor           string
	category         ConnectorCategory
	version          string
	status           ConnectorStatus
	description      string
	homepage         string
	documentationURL string
	capabilities     []*Capability
	createdAt        time.Time
	updatedAt        time.Time
	deletedAt        *time.Time
}

func NewConnector(
	id uuid.UUID,
	code, name, vendor string,
	category ConnectorCategory,
	version, description, homepage, documentationURL string,
	now time.Time,
) (*Connector, error) {
	code = strings.TrimSpace(strings.ToLower(code))
	name = strings.TrimSpace(name)
	vendor = strings.TrimSpace(vendor)
	version = strings.TrimSpace(version)

	if id == uuid.Nil {
		return nil, ErrInvalidCode
	}
	if code == "" || len(code) > 100 {
		return nil, ErrInvalidCode
	}
	if name == "" || len(name) > 255 {
		return nil, ErrInvalidName
	}
	if vendor == "" || len(vendor) > 255 {
		return nil, ErrInvalidVendor
	}
	if _, err := ParseConnectorCategory(string(category)); err != nil {
		return nil, err
	}
	if version == "" || len(version) > 50 {
		return nil, ErrInvalidVersion
	}

	c := &Connector{
		id:               id,
		code:             code,
		name:             name,
		vendor:           vendor,
		category:         category,
		version:          version,
		status:           ConnectorStatusRegistered,
		description:      strings.TrimSpace(description),
		homepage:         strings.TrimSpace(homepage),
		documentationURL: strings.TrimSpace(documentationURL),
		createdAt:        now.UTC(),
		updatedAt:        now.UTC(),
	}
	c.record(EventConnectorRegistered, id, now.UTC(), map[string]any{
		"code":     code,
		"name":     name,
		"vendor":   vendor,
		"category": string(category),
	})
	return c, nil
}

func ReconstituteConnector(
	id uuid.UUID,
	code, name, vendor string,
	category ConnectorCategory,
	version string,
	status ConnectorStatus,
	description, homepage, documentationURL string,
	capabilities []*Capability,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
) *Connector {
	return &Connector{
		id:               id,
		code:             code,
		name:             name,
		vendor:           vendor,
		category:         category,
		version:          version,
		status:           status,
		description:      description,
		homepage:         homepage,
		documentationURL: documentationURL,
		capabilities:     capabilities,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
		deletedAt:        deletedAt,
	}
}

func (c *Connector) ID() uuid.UUID               { return c.id }
func (c *Connector) Code() string                { return c.code }
func (c *Connector) Name() string                { return c.name }
func (c *Connector) Vendor() string              { return c.vendor }
func (c *Connector) Category() ConnectorCategory { return c.category }
func (c *Connector) Version() string             { return c.version }
func (c *Connector) Status() ConnectorStatus     { return c.status }
func (c *Connector) Description() string         { return c.description }
func (c *Connector) Homepage() string            { return c.homepage }
func (c *Connector) DocumentationURL() string    { return c.documentationURL }
func (c *Connector) Capabilities() []*Capability { return c.capabilities }
func (c *Connector) CreatedAt() time.Time        { return c.createdAt }
func (c *Connector) UpdatedAt() time.Time        { return c.updatedAt }
func (c *Connector) DeletedAt() *time.Time       { return c.deletedAt }

// IsUsable reports whether new Connections may reference this Connector.
func (c *Connector) IsUsable() bool {
	return c.status == ConnectorStatusEnabled && c.deletedAt == nil
}

// AddCapability appends a new Capability to the Connector's child list.
func (c *Connector) AddCapability(id uuid.UUID, code, displayName string, isAsync bool, now time.Time) (*Capability, error) {
	if c.status == ConnectorStatusRemoved {
		return nil, ErrConnectorRemoved
	}
	for _, existing := range c.capabilities {
		if existing.code == strings.TrimSpace(strings.ToLower(code)) {
			return nil, ErrConflict
		}
	}
	k, err := newCapability(id, c.id, strings.ToLower(code), displayName, isAsync, now)
	if err != nil {
		return nil, err
	}
	c.capabilities = append(c.capabilities, k)
	c.updatedAt = now.UTC()
	c.record(EventConnectorCapabilityAdded, c.id, now.UTC(), map[string]any{"capability_code": k.code})
	return k, nil
}

func (c *Connector) Enable(now time.Time) error {
	if c.status == ConnectorStatusRemoved {
		return ErrConnectorRemoved
	}
	if c.status != ConnectorStatusRegistered && c.status != ConnectorStatusDisabled {
		return ErrInvalidTransition
	}
	c.status = ConnectorStatusEnabled
	c.updatedAt = now.UTC()
	c.record(EventConnectorEnabled, c.id, now.UTC(), nil)
	return nil
}

func (c *Connector) Disable(now time.Time) error {
	if c.status != ConnectorStatusEnabled {
		return ErrInvalidTransition
	}
	c.status = ConnectorStatusDisabled
	c.updatedAt = now.UTC()
	c.record(EventConnectorDisabled, c.id, now.UTC(), nil)
	return nil
}

// Remove is a terminal transition; a removed Connector can never be re-enabled.
func (c *Connector) Remove(now time.Time) error {
	if c.status == ConnectorStatusRemoved {
		return ErrInvalidTransition
	}
	at := now.UTC()
	c.status = ConnectorStatusRemoved
	c.deletedAt = &at
	c.updatedAt = at
	c.record(EventConnectorRemoved, c.id, at, nil)
	return nil
}
