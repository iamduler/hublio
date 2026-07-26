package server

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"syscall"

	"hublio/internal/platform/apperr"
)

// wrapListenError wraps HTTP listen failures with actionable context.
// Port conflicts (EADDRINUSE) are mapped to CONFLICT so operators know to
// change SERVER_PORT or free the address.
func wrapListenError(addr string, err error) error {
	if err == nil {
		return nil
	}

	if isAddrInUse(err) {
		msg := fmt.Sprintf(
			"HTTP listen address %s is already in use; set SERVER_PORT to a free port or stop the other process (common on this host: Apache on :8080)",
			addr,
		)
		return apperr.Wrap(err, msg, apperr.ErrCodeConflict)
	}

	return apperr.Wrap(
		err,
		fmt.Sprintf("failed to start HTTP server on %s", addr),
		apperr.ErrCodeInternal,
	)
}

func isAddrInUse(err error) bool {
	if errors.Is(err, syscall.EADDRINUSE) {
		return true
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if errors.Is(opErr.Err, syscall.EADDRINUSE) {
			return true
		}
		var errno syscall.Errno
		if errors.As(opErr.Err, &errno) && errno == syscall.EADDRINUSE {
			return true
		}
	}

	// Fallback for platforms / wrappers that only expose the message.
	return strings.Contains(strings.ToLower(err.Error()), "address already in use")
}
