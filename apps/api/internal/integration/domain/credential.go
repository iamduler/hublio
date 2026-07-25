package domain

import (
	"time"

	"github.com/google/uuid"
)

type CredentialType string

const (
	CredentialTypeAPIKey      CredentialType = "api_key"
	CredentialTypeOAuth2      CredentialType = "oauth2"
	CredentialTypeBearerToken CredentialType = "bearer_token"
	CredentialTypeBasicAuth   CredentialType = "basic_auth"
	CredentialTypeJWT         CredentialType = "jwt"
)

func ParseCredentialType(v string) (CredentialType, error) {
	switch CredentialType(v) {
	case CredentialTypeAPIKey, CredentialTypeOAuth2, CredentialTypeBearerToken, CredentialTypeBasicAuth, CredentialTypeJWT:
		return CredentialType(v), nil
	default:
		return "", ErrInvalidCredentialType
	}
}

type CredentialStatus string

const (
	CredentialStatusActive  CredentialStatus = "active"
	CredentialStatusExpired CredentialStatus = "expired"
	CredentialStatusRevoked CredentialStatus = "revoked"
)

// Credential is a Connection child holding an opaque, already-encrypted secret.
// The Domain never sees plaintext; encryption/decryption is an Application/Infrastructure concern.
type Credential struct {
	eventRecorder

	id              uuid.UUID
	connectionID    uuid.UUID
	credType        CredentialType
	status          CredentialStatus
	version         int
	encryptedSecret []byte
	expiresAt       *time.Time
	rotatedAt       *time.Time
	createdAt       time.Time
	updatedAt       time.Time
	createdBy       uuid.UUID
}

func NewCredential(
	id, connectionID uuid.UUID,
	credType CredentialType,
	version int,
	encryptedSecret []byte,
	expiresAt *time.Time,
	createdBy uuid.UUID,
	now time.Time,
) (*Credential, error) {
	if id == uuid.Nil || connectionID == uuid.Nil || createdBy == uuid.Nil {
		return nil, ErrInvalidName
	}
	if _, err := ParseCredentialType(string(credType)); err != nil {
		return nil, err
	}
	if len(encryptedSecret) == 0 {
		return nil, ErrEmptySecret
	}
	if version < 1 {
		version = 1
	}

	cred := &Credential{
		id:              id,
		connectionID:    connectionID,
		credType:        credType,
		status:          CredentialStatusActive,
		version:         version,
		encryptedSecret: encryptedSecret,
		expiresAt:       expiresAt,
		createdBy:       createdBy,
		createdAt:       now.UTC(),
		updatedAt:       now.UTC(),
	}
	cred.record(EventCredentialCreated, id, now.UTC(), map[string]any{
		"connection_id": connectionID.String(),
		"type":          string(credType),
		"version":       version,
	})
	return cred, nil
}

func ReconstituteCredential(
	id, connectionID uuid.UUID,
	credType CredentialType,
	status CredentialStatus,
	version int,
	encryptedSecret []byte,
	expiresAt, rotatedAt *time.Time,
	createdAt, updatedAt time.Time,
	createdBy uuid.UUID,
) *Credential {
	return &Credential{
		id:              id,
		connectionID:    connectionID,
		credType:        credType,
		status:          status,
		version:         version,
		encryptedSecret: encryptedSecret,
		expiresAt:       expiresAt,
		rotatedAt:       rotatedAt,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
		createdBy:       createdBy,
	}
}

func (c *Credential) ID() uuid.UUID            { return c.id }
func (c *Credential) ConnectionID() uuid.UUID  { return c.connectionID }
func (c *Credential) Type() CredentialType     { return c.credType }
func (c *Credential) Status() CredentialStatus { return c.status }
func (c *Credential) Version() int             { return c.version }
func (c *Credential) EncryptedSecret() []byte  { return c.encryptedSecret }
func (c *Credential) ExpiresAt() *time.Time    { return c.expiresAt }
func (c *Credential) RotatedAt() *time.Time    { return c.rotatedAt }
func (c *Credential) CreatedAt() time.Time     { return c.createdAt }
func (c *Credential) UpdatedAt() time.Time     { return c.updatedAt }
func (c *Credential) CreatedBy() uuid.UUID     { return c.createdBy }

// IsUsable reports whether the Credential may be used to Verify/Invoke a Connector Runtime.
func (c *Credential) IsUsable(now time.Time) error {
	if c.status != CredentialStatusActive {
		return ErrCredentialNotActive
	}
	if c.expiresAt != nil && !c.expiresAt.After(now.UTC()) {
		return ErrCredentialNotActive
	}
	return nil
}

// Revoke marks the Credential permanently unusable; used when rotating or disabling.
func (c *Credential) Revoke(now time.Time) error {
	if c.status == CredentialStatusRevoked {
		return ErrInvalidTransition
	}
	at := now.UTC()
	c.status = CredentialStatusRevoked
	c.rotatedAt = &at
	c.updatedAt = at
	c.record(EventCredentialRevoked, c.id, at, map[string]any{"connection_id": c.connectionID.String()})
	return nil
}

func (c *Credential) MarkExpired(now time.Time) error {
	if c.status != CredentialStatusActive {
		return ErrInvalidTransition
	}
	c.status = CredentialStatusExpired
	c.updatedAt = now.UTC()
	return nil
}

// RotateCredential revokes the previous active Credential (if any) and returns a new
// Credential with an incremented version. Keeping the revoke+increment invariant here (rather
// than in Application) ensures it stays testable without PostgreSQL.
func RotateCredential(
	newID uuid.UUID,
	previous *Credential,
	connectionID uuid.UUID,
	credType CredentialType,
	encryptedSecret []byte,
	expiresAt *time.Time,
	createdBy uuid.UUID,
	now time.Time,
) (*Credential, error) {
	version := 1
	if previous != nil {
		if err := previous.Revoke(now); err != nil {
			return nil, err
		}
		version = previous.Version() + 1
	}
	next, err := NewCredential(newID, connectionID, credType, version, encryptedSecret, expiresAt, createdBy, now)
	if err != nil {
		return nil, err
	}
	next.record(EventCredentialRotated, next.id, now.UTC(), map[string]any{
		"connection_id": connectionID.String(),
		"version":       version,
	})
	return next, nil
}
