package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type OAuthProvider string

const (
	OAuthProviderGoogle    OAuthProvider = "google"
	OAuthProviderMicrosoft OAuthProvider = "microsoft"
	OAuthProviderGitHub    OAuthProvider = "github"
)

func ParseOAuthProvider(raw string) (OAuthProvider, error) {
	switch OAuthProvider(strings.TrimSpace(strings.ToLower(raw))) {
	case OAuthProviderGoogle:
		return OAuthProviderGoogle, nil
	case OAuthProviderMicrosoft:
		return OAuthProviderMicrosoft, nil
	case OAuthProviderGitHub:
		return OAuthProviderGitHub, nil
	default:
		return "", ErrInvalidOAuthProvider
	}
}

// OAuthIdentity links a verified IdP subject to a User (Identity configuration entity).
type OAuthIdentity struct {
	id              uuid.UUID
	userID          uuid.UUID
	provider        OAuthProvider
	providerSubject string
	email           string
	linkedAt        time.Time
	lastLoginAt     *time.Time
	createdAt       time.Time
	updatedAt       time.Time
}

func NewOAuthIdentity(
	id, userID uuid.UUID,
	provider OAuthProvider,
	providerSubject, email string,
	now time.Time,
) (*OAuthIdentity, error) {
	providerSubject = strings.TrimSpace(providerSubject)
	email = strings.TrimSpace(strings.ToLower(email))
	if id == uuid.Nil || userID == uuid.Nil {
		return nil, ErrInvalidOAuthIdentity
	}
	if provider != OAuthProviderGoogle && provider != OAuthProviderMicrosoft && provider != OAuthProviderGitHub {
		return nil, ErrInvalidOAuthProvider
	}
	if providerSubject == "" || len(providerSubject) > 255 {
		return nil, ErrInvalidOAuthIdentity
	}
	if email == "" || !strings.Contains(email, "@") || len(email) > 255 {
		return nil, ErrInvalidEmail
	}
	at := now.UTC()
	return &OAuthIdentity{
		id:              id,
		userID:          userID,
		provider:        provider,
		providerSubject: providerSubject,
		email:           email,
		linkedAt:        at,
		createdAt:       at,
		updatedAt:       at,
	}, nil
}

func ReconstituteOAuthIdentity(
	id, userID uuid.UUID,
	provider OAuthProvider,
	providerSubject, email string,
	linkedAt time.Time,
	lastLoginAt *time.Time,
	createdAt, updatedAt time.Time,
) *OAuthIdentity {
	return &OAuthIdentity{
		id:              id,
		userID:          userID,
		provider:        provider,
		providerSubject: providerSubject,
		email:           email,
		linkedAt:        linkedAt,
		lastLoginAt:     lastLoginAt,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}
}

func (o *OAuthIdentity) ID() uuid.UUID              { return o.id }
func (o *OAuthIdentity) UserID() uuid.UUID          { return o.userID }
func (o *OAuthIdentity) Provider() OAuthProvider    { return o.provider }
func (o *OAuthIdentity) ProviderSubject() string    { return o.providerSubject }
func (o *OAuthIdentity) Email() string              { return o.email }
func (o *OAuthIdentity) LinkedAt() time.Time        { return o.linkedAt }
func (o *OAuthIdentity) LastLoginAt() *time.Time    { return o.lastLoginAt }
func (o *OAuthIdentity) CreatedAt() time.Time       { return o.createdAt }
func (o *OAuthIdentity) UpdatedAt() time.Time       { return o.updatedAt }

func (o *OAuthIdentity) RecordLogin(now time.Time) {
	at := now.UTC()
	o.lastLoginAt = &at
	o.updatedAt = at
}
