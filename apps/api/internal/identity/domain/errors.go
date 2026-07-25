package domain

import "errors"

var (
	ErrInvalidName         = errors.New("identity: invalid name")
	ErrInvalidEmail        = errors.New("identity: invalid email")
	ErrInvalidPassword     = errors.New("identity: invalid password")
	ErrInvalidEnvironment  = errors.New("identity: invalid environment")
	ErrInvalidRole         = errors.New("identity: invalid role")
	ErrInvalidTransition   = errors.New("identity: invalid status transition")
	ErrOrganizationBlocked = errors.New("identity: organization cannot perform this action")
	ErrWorkspaceDisabled   = errors.New("identity: workspace is disabled")
	ErrUserCannotLogin     = errors.New("identity: user cannot login")
	ErrAPIKeyDisabled      = errors.New("identity: api key is disabled")
	ErrAPIKeyExpired       = errors.New("identity: api key is expired")
	ErrNotFound            = errors.New("identity: not found")
	ErrConflict            = errors.New("identity: conflict")
)
