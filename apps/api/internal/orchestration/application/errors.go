package application

import (
	"errors"

	"hublio/internal/orchestration/domain"
	"hublio/internal/platform/apperr"
)

func mapDomainErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidID),
		errors.Is(err, domain.ErrInvalidCapability),
		errors.Is(err, domain.ErrInvalidStepCount):
		return apperr.Wrap(err, err.Error(), apperr.ErrCodeBadRequest)
	case errors.Is(err, domain.ErrInvalidTransition),
		errors.Is(err, domain.ErrIntentImmutable),
		errors.Is(err, domain.ErrStepsIncomplete),
		errors.Is(err, domain.ErrStepNotFound),
		errors.Is(err, domain.ErrMaxRetriesExceeded),
		errors.Is(err, domain.ErrConflict):
		return apperr.Wrap(err, err.Error(), apperr.ErrCodeConflict)
	default:
		return apperr.Wrap(err, "domain error", apperr.ErrCodeBadRequest)
	}
}

func mapRepoErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, domain.ErrNotFound) {
		return apperr.New("resource not found", apperr.ErrCodeNotFound)
	}
	if errors.Is(err, domain.ErrConflict) {
		return apperr.New("resource already exists", apperr.ErrCodeConflict)
	}
	if ae, ok := err.(*apperr.AppError); ok {
		return ae
	}
	return apperr.Wrap(err, "persistence error", apperr.ErrCodeInternal)
}
