package server

import (
	"errors"
	"net"
	"strings"
	"syscall"
	"testing"

	"hublio/internal/platform/apperr"
)

func TestWrapListenError_AddrInUse(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		addr string
		err  error
		code apperr.ErrorCode
		want string
	}{
		{
			name: "syscall EADDRINUSE",
			addr: ":8080",
			err:  syscall.EADDRINUSE,
			code: apperr.ErrCodeConflict,
			want: "already in use",
		},
		{
			name: "net.OpError wrapping EADDRINUSE",
			addr: ":8080",
			err: &net.OpError{
				Op:  "listen",
				Net: "tcp",
				Err: syscall.EADDRINUSE,
			},
			code: apperr.ErrCodeConflict,
			want: "SERVER_PORT",
		},
		{
			name: "message fallback",
			addr: ":8081",
			err:  errors.New("listen tcp :8081: bind: address already in use"),
			code: apperr.ErrCodeConflict,
			want: ":8081",
		},
		{
			name: "other listen error",
			addr: ":8080",
			err:  errors.New("listen tcp :8080: bind: permission denied"),
			code: apperr.ErrCodeInternal,
			want: "failed to start HTTP server",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := wrapListenError(tc.addr, tc.err)
			var appErr *apperr.AppError
			if !errors.As(got, &appErr) {
				t.Fatalf("expected AppError, got %T (%v)", got, got)
			}
			if appErr.Code != tc.code {
				t.Fatalf("code: got %s want %s", appErr.Code, tc.code)
			}
			if !strings.Contains(strings.ToLower(got.Error()), strings.ToLower(tc.want)) {
				t.Fatalf("message %q does not contain %q", got.Error(), tc.want)
			}
			if appErr.Err == nil {
				t.Fatal("expected wrapped underlying error")
			}
		})
	}
}

func TestWrapListenError_Nil(t *testing.T) {
	t.Parallel()
	if wrapListenError(":8080", nil) != nil {
		t.Fatal("expected nil")
	}
}
