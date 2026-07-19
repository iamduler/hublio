package domain

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrMissingParam     = errors.New("transformation: missing operation parameter")
	ErrUnknownOperation = errors.New("transformation: unknown operation type")
	ErrTypeConversion   = errors.New("transformation: type conversion failed")
	ErrInvalidTimezone  = errors.New("transformation: invalid timezone")
)

// ValidationError reports Canonical fields required by ValidateRequired that were missing or
// null. It stops the Pipeline before the Document ever reaches a Connector Runtime.
type ValidationError struct {
	MissingFields []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("transformation: missing required fields: %s", strings.Join(e.MissingFields, ", "))
}
