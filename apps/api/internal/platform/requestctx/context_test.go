package requestctx_test

import (
	"context"
	"testing"

	"hublio/internal/platform/requestctx"
)

func TestRequestContextRoundTrip(t *testing.T) {
	ctx := context.Background()
	ctx = requestctx.With(ctx, requestctx.KeyCorrelationID, "c-1")
	ctx = requestctx.With(ctx, requestctx.KeyRequestID, "r-1")
	ctx = requestctx.With(ctx, requestctx.KeyOrganizationID, "o-1")
	ctx = requestctx.With(ctx, requestctx.KeyWorkspaceID, "w-1")

	if requestctx.CorrelationID(ctx) != "c-1" {
		t.Fatal("correlation")
	}
	if requestctx.RequestID(ctx) != "r-1" {
		t.Fatal("request")
	}
	if requestctx.OrganizationID(ctx) != "o-1" || requestctx.WorkspaceID(ctx) != "w-1" {
		t.Fatal("tenant")
	}
}
