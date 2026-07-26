package middleware

import (
	"context"
	"net/http"
	"strings"

	"hublio/internal/platform/apikey"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/httpx"
	"hublio/internal/platform/requestctx"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WorkspaceMembership verifies a user belongs to a workspace.
// Implemented in the composition root over Identity MembershipRepository.
type WorkspaceMembership interface {
	Check(ctx context.Context, workspaceID, userID uuid.UUID) error
}

const workspaceIDHeader = "X-Workspace-ID"

// MachineOrJWTMiddleware authenticates either:
//   - X-API-KEY (workspace-scoped machine principal), or
//   - Bearer JWT + X-Workspace-ID (user principal; membership enforced).
//
// API-key auth remains for external/machine clients. JWT is used by the
// Next.js proxy so the dashboard no longer needs a minted workspace API key.
func MachineOrJWTMiddleware(apiKeyAuth apikey.Authenticator, membership WorkspaceMembership) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-API-KEY") != "" {
			if !authenticateAPIKey(c, apiKeyAuth) {
				return
			}
			c.Next()
			return
		}

		if !authenticateBearerJWT(c) {
			return
		}
		if !bindWorkspaceFromHeader(c, membership) {
			return
		}
		c.Next()
	}
}

func authenticateAPIKey(c *gin.Context, apiKeyAuth apikey.Authenticator) bool {
	if apiKeyAuth == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "api key authenticator not configured",
		})
		return false
	}

	raw := c.GetHeader("X-API-KEY")
	if raw == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "missing X-API-KEY",
		})
		return false
	}

	principal, err := apiKeyAuth.Authenticate(c.Request.Context(), raw)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "invalid api key",
		})
		return false
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
	return true
}

func authenticateBearerJWT(c *gin.Context) bool {
	if jwtService == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "auth not configured",
		})
		return false
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Missing or invalid authorization header",
		})
		return false
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	_, claims, err := jwtService.ParseToken(tokenString)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Missing or invalid authorization header",
		})
		return false
	}

	if jti, ok := claims["jti"].(string); ok {
		key := "token_blacklist:" + jti
		exists, err := cacheService.Exists(key)
		if err == nil && exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Token revoked",
			})
			return false
		}
	}

	payload, err := jwtService.DecryptAccessTokenPayload(tokenString)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Invalid token",
		})
		return false
	}

	c.Set("user_id", payload.UserID)
	c.Set("user_email", payload.Email)
	c.Set("user_role", payload.Role)
	c.Set("organization_id", payload.OrganizationID)
	c.Set("access_token", tokenString)

	ctx := c.Request.Context()
	ctx = requestctx.With(ctx, requestctx.KeyUserID, payload.UserID)
	ctx = requestctx.With(ctx, requestctx.KeyOrganizationID, payload.OrganizationID)
	c.Request = c.Request.WithContext(ctx)
	return true
}

func bindWorkspaceFromHeader(c *gin.Context, membership WorkspaceMembership) bool {
	raw := c.GetHeader(workspaceIDHeader)
	if raw == "" {
		httpx.ResponseError(c, apperr.New("missing X-Workspace-ID header", apperr.ErrCodeBadRequest))
		c.Abort()
		return false
	}

	workspaceID, err := uuid.Parse(raw)
	if err != nil {
		httpx.ResponseError(c, apperr.New("invalid X-Workspace-ID", apperr.ErrCodeBadRequest))
		c.Abort()
		return false
	}

	userRaw, _ := c.Get("user_id")
	userStr, _ := userRaw.(string)
	userID, err := uuid.Parse(userStr)
	if err != nil {
		httpx.ResponseError(c, apperr.New("unauthorized", apperr.ErrCodeUnauthorized))
		c.Abort()
		return false
	}

	if membership != nil {
		if err := membership.Check(c.Request.Context(), workspaceID, userID); err != nil {
			httpx.ResponseError(c, err)
			c.Abort()
			return false
		}
	}

	c.Set("workspace_id", workspaceID.String())
	ctx := requestctx.With(c.Request.Context(), requestctx.KeyWorkspaceID, workspaceID.String())
	c.Request = c.Request.WithContext(ctx)
	return true
}
