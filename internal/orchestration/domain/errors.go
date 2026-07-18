package domain

import "errors"

var (
	ErrInvalidID          = errors.New("orchestration: invalid id")
	ErrInvalidCapability  = errors.New("orchestration: invalid capability")
	ErrInvalidTransition  = errors.New("orchestration: invalid status transition")
	ErrIntentImmutable    = errors.New("orchestration: intent is immutable")
	ErrStepsIncomplete    = errors.New("orchestration: not all steps succeeded")
	ErrStepNotFound       = errors.New("orchestration: step not found")
	ErrInvalidStepCount   = errors.New("orchestration: default steps require exactly 5 ids")
	ErrNotFound           = errors.New("orchestration: not found")
	ErrConflict           = errors.New("orchestration: conflict")
	ErrMaxRetriesExceeded = errors.New("orchestration: retry attempts exhausted")
)
