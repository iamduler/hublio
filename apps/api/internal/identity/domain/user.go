package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusArchived  UserStatus = "archived"
)

// User belongs to an Organization (child entity; not a separate Aggregate).
type User struct {
	eventRecorder

	id             uuid.UUID
	organizationID uuid.UUID
	email          string
	fullName       string
	passwordHash   string
	isActive       bool
	status         UserStatus
	createdAt      time.Time
	updatedAt      time.Time
	lastLoginAt    *time.Time
	deletedAt      *time.Time
}

func NewUser(id, organizationID uuid.UUID, email, fullName, passwordHash string, now time.Time) (*User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	fullName = strings.TrimSpace(fullName)
	if id == uuid.Nil || organizationID == uuid.Nil {
		return nil, ErrInvalidEmail
	}
	if email == "" || !strings.Contains(email, "@") || len(email) > 255 {
		return nil, ErrInvalidEmail
	}
	if fullName == "" || len(fullName) > 255 {
		return nil, ErrInvalidName
	}
	if passwordHash == "" {
		return nil, ErrInvalidPassword
	}

	u := &User{
		id:             id,
		organizationID: organizationID,
		email:          email,
		fullName:       fullName,
		passwordHash:   passwordHash,
		isActive:       true,
		status:         UserStatusActive,
		createdAt:      now.UTC(),
		updatedAt:      now.UTC(),
	}
	u.record(EventUserCreated, id, now.UTC(), map[string]any{
		"organization_id": organizationID.String(),
		"email":           email,
	})
	return u, nil
}

func ReconstituteUser(
	id, organizationID uuid.UUID,
	email, fullName, passwordHash string,
	isActive bool,
	status UserStatus,
	createdAt, updatedAt time.Time,
	lastLoginAt, deletedAt *time.Time,
) *User {
	return &User{
		id:             id,
		organizationID: organizationID,
		email:          email,
		fullName:       fullName,
		passwordHash:   passwordHash,
		isActive:       isActive,
		status:         status,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
		lastLoginAt:    lastLoginAt,
		deletedAt:      deletedAt,
	}
}

func (u *User) ID() uuid.UUID             { return u.id }
func (u *User) OrganizationID() uuid.UUID { return u.organizationID }
func (u *User) Email() string             { return u.email }
func (u *User) FullName() string          { return u.fullName }
func (u *User) PasswordHash() string      { return u.passwordHash }
func (u *User) IsActive() bool            { return u.isActive }
func (u *User) Status() UserStatus        { return u.status }
func (u *User) CreatedAt() time.Time      { return u.createdAt }
func (u *User) UpdatedAt() time.Time      { return u.updatedAt }
func (u *User) LastLoginAt() *time.Time   { return u.lastLoginAt }
func (u *User) DeletedAt() *time.Time     { return u.deletedAt }

func (u *User) CanLogin() bool {
	return u.isActive && u.status == UserStatusActive && u.deletedAt == nil
}

func (u *User) RecordLogin(now time.Time) error {
	if !u.CanLogin() {
		return ErrUserCannotLogin
	}
	at := now.UTC()
	u.lastLoginAt = &at
	u.updatedAt = at
	return nil
}

func (u *User) Suspend(now time.Time) error {
	if u.status != UserStatusActive {
		return ErrInvalidTransition
	}
	u.status = UserStatusSuspended
	u.isActive = false
	u.updatedAt = now.UTC()
	return nil
}

func (u *User) Activate(now time.Time) error {
	if u.status != UserStatusSuspended && u.status != UserStatusInactive {
		return ErrInvalidTransition
	}
	u.status = UserStatusActive
	u.isActive = true
	u.updatedAt = now.UTC()
	return nil
}
