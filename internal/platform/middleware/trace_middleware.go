package middleware

import (
	"context"

	"hublio/internal/platform/logging"
	"hublio/internal/platform/requestctx"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TraceMiddleware injects trace_id, correlation_id, and request_id into context and response headers.
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}

		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = traceID
		}

		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := c.Request.Context()
		ctx = requestctx.With(ctx, requestctx.KeyTraceID, traceID)
		ctx = requestctx.With(ctx, requestctx.KeyCorrelationID, correlationID)
		ctx = requestctx.With(ctx, requestctx.KeyRequestID, requestID)
		ctx = requestctx.With(ctx, requestctx.KeyIP, c.ClientIP())
		ctx = requestctx.With(ctx, requestctx.KeyUserAgent, c.Request.UserAgent())
		ctx = context.WithValue(ctx, logging.TraceIDKey, traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Writer.Header().Set("X-Trace-ID", traceID)
		c.Writer.Header().Set("X-Correlation-ID", correlationID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Set(string(requestctx.KeyTraceID), traceID)
		c.Set(string(requestctx.KeyCorrelationID), correlationID)
		c.Set(string(requestctx.KeyRequestID), requestID)

		c.Next()
	}
}
