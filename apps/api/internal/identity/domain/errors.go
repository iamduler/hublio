package domain

import "errors"

var (
	ErrInvalidName          = errors.New("identity: invalid name")
	ErrInvalidEmail         = errors.New("identity: invalid email")
	ErrInvalidPassword      = errors.New("identity: invalid password")
	ErrInvalidEnvironment   = errors.New("identity: invalid environment")
	ErrInvalidRole          = errors.New("identity: invalid role")
	ErrInvalidTransition    = errors.New("identity: invalid status transition")
	ErrOrganizationBlocked  = errors.New("identity: organization cannot perform this action")
	ErrWorkspaceDisabled    = errors.New("identity: workspace is disabled")
	ErrUserCannotLogin      = errors.New("identity: user cannot login")
	ErrAPIKeyDisabled       = errors.New("identity: api key is disabled")
	ErrAPIKeyExpired        = errors.New("identity: api key is expired")
	ErrInvalidOAuthProvider = errors.New("identity: invalid oauth provider")
	ErrInvalidOAuthIdentity = errors.New("identity: invalid oauth identity")
	ErrOAuthEmailUnverified = errors.New("identity: oauth email not verified")
	ErrOAuthPlatformAdmin   = errors.New("identity: platform admin cannot use oauth")
	ErrInvalidMFAConfig     = errors.New("identity: invalid mfa configuration")
	ErrMFAAlreadyEnabled    = errors.New("identity: mfa is already enabled")
	ErrMFANotEnabled        = errors.New("identity: mfa is not enabled")
	ErrInvalidMFACode       = errors.New("identity: invalid mfa code")
	ErrNotFound             = errors.New("identity: not found")
	ErrConflict             = errors.New("identity: conflict")
)
