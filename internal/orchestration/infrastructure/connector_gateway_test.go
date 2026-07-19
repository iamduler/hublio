package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"testing"

	integrationdomain "hublio/internal/integration/domain"
	orchestrationapp "hublio/internal/orchestration/application"
	"hublio/internal/platform/apperr"

	"github.com/google/uuid"
)

type stubRuntime struct {
	code string
	err  error
}

func (s stubRuntime) Code() string { return s.code }
func (s stubRuntime) Verify(context.Context, integrationdomain.VerifyInput) error {
	return nil
}
func (s stubRuntime) Health(context.Context, integrationdomain.HealthInput) error {
	return nil
}
func (s stubRuntime) Invoke(context.Context, integrationdomain.InvokeInput) (integrationdomain.InvokeOutput, error) {
	return integrationdomain.InvokeOutput{}, s.err
}

type stubRegistry struct {
	rt integrationdomain.Runtime
}

func (r stubRegistry) Resolve(string) (integrationdomain.Runtime, error) {
	return r.rt, nil
}

func TestConnectorGateway_MapInvokeErr(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		code apperr.ErrorCode
	}{
		{"auth", integrationdomain.ErrRuntimeAuthFailed, apperr.ErrCodeUnauthorized},
		{"missing creds", integrationdomain.ErrRuntimeMissingCredentials, apperr.ErrCodeUnauthorized},
		{"invalid payload", integrationdomain.ErrRuntimeInvalidPayload, apperr.ErrCodeBadRequest},
		{"unsupported", integrationdomain.ErrRuntimeUnsupportedCapability, apperr.ErrCodeBadRequest},
		{"missing config", integrationdomain.ErrRuntimeMissingConfig, apperr.ErrCodeBadRequest},
		{"not found", integrationdomain.ErrRuntimeNotFound, apperr.ErrCodeNotFound},
		{"provider", integrationdomain.ErrRuntimeProviderRejected, apperr.ErrCodeBadGateway},
		{"unknown", errors.New("boom"), apperr.ErrCodeBadGateway},
		{"wrapped auth", fmt.Errorf("%w: UnAuthorize", integrationdomain.ErrRuntimeAuthFailed), apperr.ErrCodeUnauthorized},
		{"wrapped payload", fmt.Errorf("%w: items required", integrationdomain.ErrRuntimeInvalidPayload), apperr.ErrCodeBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gw := NewConnectorGateway(stubRegistry{rt: stubRuntime{code: "misa", err: tt.err}})
			_, invokeErr := gw.Invoke(context.Background(), "misa", orchestrationapp.InvokeRequest{
				ConnectionID: uuid.Must(uuid.NewV7()),
				Capability:   "invoice.create",
			})
			ae, ok := invokeErr.(*apperr.AppError)
			if !ok {
				t.Fatalf("expected AppError, got %T %v", invokeErr, invokeErr)
			}
			if ae.Code != tt.code {
				t.Fatalf("code = %s, want %s", ae.Code, tt.code)
			}
		})
	}
}
