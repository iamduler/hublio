package middleware

import (
	"context"

	"shopping-cart/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")

		if traceID == "" {
			traceID = uuid.New().String()
		}

		// Set the trace ID in the context
		contextValue := context.WithValue(c.Request.Context(), logger.TraceIDKey, traceID)
		c.Request = c.Request.WithContext(contextValue)

		// Set the trace ID in the response header
		c.Writer.Header().Set("X-Trace-ID", traceID)

		// Set the trace ID in the Gin context
		c.Set(string(logger.TraceIDKey), traceID)

		c.Next()
	}
}
