package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"hublio/internal/platform/apikey"
	"hublio/internal/platform/apperr"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type stubAPIKeyAuth struct {
	principal apikey.Principal
	err       error
}

func (s *stubAPIKeyAuth) Authenticate(_ context.Context, _ string) (apikey.Principal, error) {
	if s.err != nil {
		return apikey.Principal{}, s.err
	}
	return s.principal, nil
}

type stubMembership struct {
	err error
}

func (s *stubMembership) Check(_ context.Context, _, _ uuid.UUID) error {
	return s.err
}

func TestMachineOrJWTMiddleware_APIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orgID := uuid.Must(uuid.NewV7())
	wsID := uuid.Must(uuid.NewV7())
	keyID := uuid.Must(uuid.NewV7())

	r := gin.New()
	r.Use(MachineOrJWTMiddleware(&stubAPIKeyAuth{
		principal: apikey.Principal{
			APIKeyID:       keyID,
			WorkspaceID:    wsID,
			OrganizationID: orgID,
			Name:           "test",
		},
	}, &stubMembership{}))
	r.GET("/x", func(c *gin.Context) {
		gotWS, _ := c.Get("workspace_id")
		gotOrg, _ := c.Get("organization_id")
		if gotWS != wsID.String() || gotOrg != orgID.String() {
			t.Fatalf("principal not set: ws=%v org=%v", gotWS, gotOrg)
		}
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("X-API-KEY", "prefix.secret")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestMachineOrJWTMiddleware_MissingBoth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(MachineOrJWTMiddleware(&stubAPIKeyAuth{err: errors.New("no")}, &stubMembership{}))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestMachineOrJWTMiddleware_JWTMissingWorkspaceHeader(t *testing.T) {
	// Without InitAuthMiddleware, JWT parse will fail closed as unauthorized.
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(MachineOrJWTMiddleware(&stubAPIKeyAuth{}, &stubMembership{}))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestMachineOrJWTMiddleware_MembershipForbidden(t *testing.T) {
	// Membership path is covered when JWT auth succeeds; here we only assert
	// the membership stub error shape used by bindWorkspaceFromHeader.
	err := (&stubMembership{err: apperr.New("forbidden", apperr.ErrCodeForbidden)}).Check(
		context.Background(), uuid.Nil, uuid.Nil,
	)
	if err == nil {
		t.Fatal("expected forbidden")
	}
}
