package middleware

import (
	"net/http"

	"hublio/internal/platform/apikey"
	"hublio/internal/platform/requestctx"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// APIKeyMiddleware authenticates Workspace-scoped API keys via Authenticator.
// Fail-closed: missing authenticator or invalid key => 401.
func APIKeyMiddleware(auth apikey.Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		if auth == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "api key authenticator not configured",
			})
			return
		}

		raw := c.GetHeader("X-API-KEY")
		if raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "missing X-API-KEY",
			})
			return
		}

		principal, err := auth.Authenticate(c.Request.Context(), raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "invalid api key",
			})
			return
		}

		ctx := c.Request.Context()
		if principal.APIKeyID != uuid.Nil {
			ctx = requestctx.With(ctx, requestctx.KeyAPIKeyID, principal.APIKeyID.String())
		}
		if principal.WorkspaceID != uuid.Nil {
			ctx = requestctx.With(ctx, requestctx.KeyWorkspaceID, principal.WorkspaceID.String())
			c.Set("workspace_id", principal.WorkspaceID.String())
		}
		if principal.OrganizationID != uuid.Nil {
			ctx = requestctx.With(ctx, requestctx.KeyOrganizationID, principal.OrganizationID.String())
			c.Set("organization_id", principal.OrganizationID.String())
		}
		c.Request = c.Request.WithContext(ctx)
		c.Set("api_key_name", principal.Name)

		c.Next()
	}
}
