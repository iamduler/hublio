package domain

import "errors"

var (
	ErrInvalidID            = errors.New("events: invalid id")
	ErrInvalidAggregateType = errors.New("events: invalid aggregate type")
	ErrInvalidCategory      = errors.New("events: invalid category")
	ErrInvalidEventName     = errors.New("events: event name is required and must be <= 150 chars")
	ErrInvalidActorType     = errors.New("events: invalid actor type")
	ErrInvalidAction        = errors.New("events: action is required and must be <= 150 chars")
	ErrInvalidResourceType  = errors.New("events: resource type is required and must be <= 100 chars")
	ErrNotFound             = errors.New("events: not found")
)
