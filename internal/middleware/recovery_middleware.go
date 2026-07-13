package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func RecoveryMiddleware(logger *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				stackTrace := string(debug.Stack())

				stackLine := ExtractFirstAppStackLine([]byte(stackTrace))

				logger.Error().
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Str("ip", c.ClientIP()).
					Str("query", c.Request.URL.RawQuery).
					Str("user-agent", c.Request.UserAgent()).
					Str("referer", c.Request.Referer()).
					Str("protocol", c.Request.Proto).
					Str("host", c.Request.Host).
					Str("remote-addr", c.Request.RemoteAddr).
					Str("request-uri", c.Request.RequestURI).
					Str("panic-message", fmt.Sprintf("%v", r)).
					Str("stack-line", stackLine).
					Str("stack-trace", stackTrace).
					Msg("Panic recovered")

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "Please try again later.",
				})
			}
		}()

		c.Next()
	}
}

var stackLineRegex = regexp.MustCompile(`(.+\.go:\d+)`)

func isSkippedStackLine(line []byte) bool {
	skipPatterns := [][]byte{
		[]byte("/runtime/"),
		[]byte("/debug/"),
		[]byte("recovery_middleware.go"),
		[]byte("pkg/mod/"),
		[]byte("/middleware/"),
	}

	for _, pattern := range skipPatterns {
		if bytes.Contains(line, pattern) {
			return true
		}
	}

	return false
}

func ExtractFirstAppStackLine(stackTrace []byte) string {
	lines := bytes.Split(stackTrace, []byte("\n"))

	for _, line := range lines {
		if !bytes.Contains(line, []byte(".go:")) || isSkippedStackLine(line) {
			continue
		}

		cleanLine := strings.TrimSpace(string(line))

		if matches := stackLineRegex.FindStringSubmatch(cleanLine); len(matches) > 1 {
			return matches[1]
		}

		return cleanLine
	}

	return ""
}
