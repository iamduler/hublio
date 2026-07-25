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
		errors.Is(err, domain.ErrEmptySecret),
		errors.Is(err, domain.ErrInvalidSyncRouteTrigger),
		errors.Is(err, domain.ErrInvalidResourceTypes),
		errors.Is(err, domain.ErrInvalidActivityGroup),
		errors.Is(err, domain.ErrInvalidActivityStep),
		errors.Is(err, domain.ErrInvalidReverseConfig),
		errors.Is(err, domain.ErrInvalidSchedule),
		errors.Is(err, domain.ErrWebhookSecretRequired),
		errors.Is(err, domain.ErrInvalidFilter),
		errors.Is(err, domain.ErrResourceTypeNotAllowed),
		errors.Is(err, domain.ErrFilterRejected),
		errors.Is(err, domain.ErrWebhookNotConfigured):
		return apperr.Wrap(err, err.Error(), apperr.ErrCodeBadRequest)
	case errors.Is(err, domain.ErrWebhookSecretMismatch):
		return apperr.Wrap(err, "unauthorized", apperr.ErrCodeUnauthorized)
	case errors.Is(err, domain.ErrInvalidTransition),
		errors.Is(err, domain.ErrConnectorRemoved),
		errors.Is(err, domain.ErrConnectorNotUsable),
		errors.Is(err, domain.ErrCapabilityNotFound),
		errors.Is(err, domain.ErrCredentialNotActive),
		errors.Is(err, domain.ErrConnectionNotActive),
		errors.Is(err, domain.ErrConflict),
		errors.Is(err, domain.ErrSyncRouteRemoved),
		errors.Is(err, domain.ErrSyncRouteNotEditable),
		errors.Is(err, domain.ErrSyncRouteNotEnabled):
		return apperr.Wrap(err, err.Error(), apperr.ErrCodeConflict)
	default:
		return apperr.Wrap(err, "domain error", apperr.ErrCodeBadRequest)
	}
}

// mapRuntimeErr translates Connector Runtime sentinels (and wraps) for Verify/Health paths.
// Provider ErrorCode strings may appear in the wrapped message; never include secrets.
func mapRuntimeErr(err error) error {
	if err == nil {
		return nil
	}
	if ae, ok := err.(*apperr.AppError); ok {
		return ae
	}
	switch {
	case errors.Is(err, domain.ErrRuntimeAuthFailed),
		errors.Is(err, domain.ErrRuntimeMissingCredentials):
		return apperr.Wrap(err, "connector authentication failed", apperr.ErrCodeUnauthorized)
	case errors.Is(err, domain.ErrRuntimeInvalidPayload),
		errors.Is(err, domain.ErrRuntimeMissingConfig),
		errors.Is(err, domain.ErrRuntimeUnsupportedCapability):
		return apperr.Wrap(err, "connector rejected request", apperr.ErrCodeBadRequest)
	case errors.Is(err, domain.ErrRuntimeNotFound):
		return apperr.Wrap(err, "connector resource not found", apperr.ErrCodeNotFound)
	case errors.Is(err, domain.ErrRuntimeProviderRejected):
		return apperr.Wrap(err, "connector provider rejected request", apperr.ErrCodeBadGateway)
	default:
		return apperr.Wrap(err, "connector runtime error", apperr.ErrCodeBadGateway)
	}
}

// verifyFailureReason stores a short, secret-free string on the Connection after a failed Verify.
func verifyFailureReason(err error) string {
	if err == nil {
		return ""
	}
	if ae, ok := err.(*apperr.AppError); ok {
		if ae.Err != nil {
			return ae.Err.Error()
		}
		return ae.Message
	}
	return err.Error()
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
