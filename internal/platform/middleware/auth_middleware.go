package middleware

import (
	"net/http"
	"strings"

	"hublio/internal/platform/auth"
	"hublio/internal/platform/cache"
	"hublio/internal/platform/requestctx"

	"github.com/gin-gonic/gin"
)

var (
	jwtService   auth.TokenService
	cacheService cache.RedisCacheService
)

func InitAuthMiddleware(service auth.TokenService, redisCache cache.RedisCacheService) {
	jwtService = service
	cacheService = redisCache
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "message": "Missing or invalid authorization header"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		_, claims, err := jwtService.ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "message": "Missing or invalid authorization header"})
			return
		}

		if jti, ok := claims["jti"].(string); ok {
			key := "token_blacklist:" + jti
			exists, err := cacheService.Exists(key)
			if err == nil && exists {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "message": "Token revoked"})
				return
			}
		}

		payload, err := jwtService.DecryptAccessTokenPayload(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "message": "Invalid token"})
			return
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

		c.Next()
	}
}
