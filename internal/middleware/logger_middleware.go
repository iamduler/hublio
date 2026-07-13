package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"shopping-cart/pkg/logger"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type CustomResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *CustomResponseWriter) Write(data []byte) (n int, err error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func LoggerMiddleware(httpLogger *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		contentType := c.GetHeader("Content-Type")
		requestBody := make(map[string]any)
		var formFiles []map[string]any
		var sensitiveFields = []string{"password", "pass", "confirm_password"}

		if strings.HasPrefix(contentType, "multipart/form-data") {
			// Content-Type is multipart/form-data
			if err := c.Request.ParseMultipartForm(32 << 20); err == nil && c.Request.MultipartForm != nil {
				// For value
				for key, value := range c.Request.MultipartForm.Value {
					if len(value) == 1 {
						requestBody[key] = value[0]
					} else {
						requestBody[key] = value
					}
				}

				// For file
				for key, files := range c.Request.MultipartForm.File {
					for _, file := range files {
						formFiles = append(formFiles, map[string]any{
							"field":    key,
							"filename": file.Filename,
							"size":     formatFileSize(file.Size),
							"type":     file.Header.Get("Content-Type"),
						})
					}
				}

				if len(formFiles) > 0 {
					requestBody["form-files"] = formFiles
				}
			}
		} else {
			// For other content types, read the request body
			bodyBytes, err := io.ReadAll(c.Request.Body) // Read the request body -> c.Request.Body will be empty after this

			if err != nil {
				httpLogger.Error().Err(err).Msg("Failed to read request body")
			}

			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reset the request body

			if strings.HasPrefix(contentType, "application/json") {
				// Content-Type is application/json
				_ = json.Unmarshal(bodyBytes, &requestBody)

			} else {
				// Content-Type is application/x-www-form-urlencoded
				values, _ := url.ParseQuery(string(bodyBytes))

				for key, value := range values {
					if len(value) == 1 {
						requestBody[key] = value[0]
					} else {
						requestBody[key] = value
					}
				}
			}
		}

		// Create a custom response writer to capture the response body
		customResponseWriter := &CustomResponseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}

		// Replace the default response writer with the custom response writer
		c.Writer = customResponseWriter

		c.Next()

		latency := time.Since(startTime)
		statusCode := c.Writer.Status()

		// Check response body
		responseContentType := c.Writer.Header().Get("Content-Type")
		responseBodyRaw := customResponseWriter.body.String()
		var responseBodyParsed interface{}

		// Parse the response body
		if strings.HasPrefix(responseContentType, "image/") ||
			strings.HasPrefix(responseContentType, "video/") ||
			strings.HasPrefix(responseContentType, "application/octet-stream") ||
			strings.HasPrefix(responseContentType, "application/pdf") {
			responseBodyParsed = "[BINARY DATA]"
		} else if strings.HasPrefix(responseContentType, "application/json") ||
			strings.HasPrefix(strings.TrimSpace(responseBodyRaw), "{") ||
			strings.HasPrefix(strings.TrimSpace(responseBodyRaw), "[") {
			if err := json.Unmarshal([]byte(responseBodyRaw), &responseBodyParsed); err != nil {
				responseBodyParsed = responseBodyRaw
			}
		} else {
			responseBodyParsed = responseBodyRaw
		}

		logEvent := httpLogger.Info()

		if statusCode >= 500 {
			logEvent = httpLogger.Error()
		} else if statusCode >= 400 {
			logEvent = httpLogger.Warn()
		}

		logEvent.Str("method", c.Request.Method).
			Str("trace-id", logger.GetTraceID(c.Request.Context())).
			Str("path", c.Request.URL.Path).
			Str("ip", c.ClientIP()).
			Str("query", c.Request.URL.RawQuery).
			Str("user-agent", c.Request.UserAgent()).
			Str("referer", c.Request.Referer()).
			Str("protocol", c.Request.Proto).
			Str("host", c.Request.Host).
			Str("remote-addr", c.Request.RemoteAddr).
			Str("request-uri", c.Request.RequestURI).
			Interface("headers", c.Request.Header).
			Interface("request-body", senitizeRequestBody(requestBody, sensitiveFields)).
			Interface("response-body", responseBodyParsed).
			Str("resquest-content-type", contentType).
			Str("response-content-type", responseContentType).
			Int("status-code", statusCode).
			Dur("latency", latency).
			Msg("HTTP Request")
	}
}

func formatFileSize(size int64) string {
	switch {
	case size < 1024:
		return fmt.Sprintf("%d B", size)
	case size < 1024*1024:
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	case size < 1024*1024*1024:
		return fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
	default:
		return fmt.Sprintf("%.2f GB", float64(size)/(1024*1024*1024))
	}
}

func senitizeRequestBody(requestBody map[string]any, sensitiveFields []string) map[string]any {
	senitizedBody := make(map[string]any)

	for key, value := range requestBody {
		lowerKey := strings.ToLower(key)
		shouldMask := false

		for _, s := range sensitiveFields {
			if lowerKey == s {
				shouldMask = true
				break
			}
		}

		if shouldMask {
			senitizedBody[key] = "********"
		} else {
			switch v := value.(type) {
			case map[string]any:
				senitizedBody[key] = senitizeRequestBody(v, sensitiveFields)
			case []any: // slice
				var senitizedSlice []any

				for _, item := range v {
					if m, ok := item.(map[string]any); ok {
						senitizedSlice = append(senitizedSlice, senitizeRequestBody(m, sensitiveFields))
					} else {
						senitizedSlice = append(senitizedSlice, item)
					}
				}

				senitizedBody[key] = senitizedSlice
			default:
				senitizedBody[key] = value
			}
		}
	}

	return senitizedBody
}
