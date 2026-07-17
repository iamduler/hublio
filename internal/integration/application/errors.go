package application

import (
	"errors"

	"hublio/internal/integration/domain"
	"hublio/internal/platform/apperr"
)

func mapDomainErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidCode),
		errors.Is(err, domain.ErrInvalidName),
		errors.Is(err, domain.ErrInvalidVendor),
		errors.Is(err, domain.ErrInvalidCategory),
		errors.Is(err, domain.ErrInvalidVersion),
		errors.Is(err, domain.ErrInvalidEnvironment),
		errors.Is(err, domain.ErrInvalidCredentialType),
		errors.Is(err, domain.ErrEmptySecret):
		return apperr.Wrap(err, err.Error(), apperr.ErrCodeBadRequest)
	case errors.Is(err, domain.ErrInvalidTransition),
		errors.Is(err, domain.ErrConnectorRemoved),
		errors.Is(err, domain.ErrConnectorNotUsable),
		errors.Is(err, domain.ErrCapabilityNotFound),
		errors.Is(err, domain.ErrCredentialNotActive),
		errors.Is(err, domain.ErrConnectionNotActive),
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
