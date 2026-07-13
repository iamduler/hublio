package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func ApiKeyMiddleware() gin.HandlerFunc {
	expectedApiKey := os.Getenv("API_KEY")

	if expectedApiKey == "" {
		expectedApiKey = "secret-api-key"
	}

	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-KEY")

		if apiKey != expectedApiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		c.Next()
	}
}
