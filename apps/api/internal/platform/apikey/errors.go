package apikey

import "hublio/internal/platform/apperr"

var ErrUnauthorized = apperr.New("invalid api key", apperr.ErrCodeUnauthorized)
