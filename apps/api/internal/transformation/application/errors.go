package application

import (
	"errors"

	"hublio/internal/platform/apperr"
	"hublio/internal/transformation/domain"
)

func mapDomainErr(err error) error {
	var verr *domain.ValidationError
	if errors.As(err, &verr) {
		return apperr.Wrap(err, err.Error(), apperr.ErrCodeBadRequest)
	}
	switch {
	case errors.Is(err, domain.ErrMissingParam),
		errors.Is(err, domain.ErrUnknownOperation),
		errors.Is(err, domain.ErrTypeConversion),
		errors.Is(err, domain.ErrInvalidTimezone):
		return apperr.Wrap(err, err.Error(), apperr.ErrCodeBadRequest)
	default:
		return apperr.Wrap(err, "transformation error", apperr.ErrCodeBadRequest)
	}
}
