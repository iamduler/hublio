package httpx

import (
	"net/http"

	"hublio/internal/platform/apperr"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Status     string `json:"status"`
	Message    string `json:"message,omitempty"`
	Data       any    `json:"data,omitempty"`
	Pagination any    `json:"pagination,omitempty"`
}

func ResponseError(c *gin.Context, err error) {
	if appErr, ok := err.(*apperr.AppError); ok {
		response := gin.H{
			"error": appErr.Message,
			"code":  appErr.Code,
		}

		if appErr.Err != nil {
			response["detail"] = appErr.Err.Error()
		}

		c.JSON(httpStatusFromCode(appErr.Code), response)
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"error": err.Error(),
		"code":  apperr.ErrCodeInternal,
	})
}

func ResponseSuccess(c *gin.Context, status int, message string, data ...any) {
	resp := APIResponse{
		Status:  "success",
		Message: CapitalizeFirstLetter(message),
	}

	if len(data) > 0 && data[0] != nil {
		if m, ok := data[0].(map[string]any); ok {
			if p, exists := m["pagination"]; exists {
				resp.Pagination = p
			}

			if d, exists := m["data"]; exists {
				resp.Data = d
			} else {
				resp.Data = m
			}
		} else {
			resp.Data = data[0]
		}
	}

	c.JSON(status, resp)
}

func ResponseStatusCode(c *gin.Context, status int) {
	c.Status(status)
}

func ResponseValidation(c *gin.Context, data any) {
	c.JSON(http.StatusBadRequest, data)
}

func httpStatusFromCode(code apperr.ErrorCode) int {
	switch code {
	case apperr.ErrCodeBadRequest:
		return http.StatusBadRequest
	case apperr.ErrCodeNotFound:
		return http.StatusNotFound
	case apperr.ErrCodeConflict:
		return http.StatusConflict
	case apperr.ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case apperr.ErrCodeForbidden:
		return http.StatusForbidden
	case apperr.ErrCodeInternal:
		return http.StatusInternalServerError
	case apperr.ErrCodeTooManyRequests:
		return http.StatusTooManyRequests
	case apperr.ErrCodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case apperr.ErrCodeBadGateway:
		return http.StatusBadGateway
	case apperr.ErrCodeGatewayTimeout:
		return http.StatusGatewayTimeout
	}

	return http.StatusInternalServerError
}
