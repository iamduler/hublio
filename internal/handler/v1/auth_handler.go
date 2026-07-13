package v1handler

import (
	"net/http"
	v1dto "shopping-cart/internal/dto/v1"
	v1service "shopping-cart/internal/service/v1"
	"shopping-cart/internal/utils"
	"shopping-cart/internal/validation"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service v1service.AuthService
}

func NewAuthHandler(service v1service.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input v1dto.LoginDTO

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidation(c, validation.HandleValidationErrors(err))
		return
	}

	// Login
	accessToken, refreshToken, expiresAt, err := h.service.Login(c, input.Email, input.Password)

	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	response := v1dto.LoginResponseDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresAt,
	}

	utils.ResponseSuccess(c, http.StatusOK, "Login successful", response)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var input v1dto.RefreshTokenDTO

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidation(c, validation.HandleValidationErrors(err))
		return
	}

	if err := h.service.Logout(c, input.RefreshToken); err != nil {
		utils.ResponseError(c, err)
		return
	}

	utils.ResponseSuccess(c, http.StatusOK, "Logout successful")
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var input v1dto.RefreshTokenDTO

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidation(c, validation.HandleValidationErrors(err))
		return
	}

	// Refresh token
	accessToken, refreshToken, expiresAt, err := h.service.RefreshToken(c, input.RefreshToken)

	if err != nil {
		utils.ResponseError(c, err)
		return
	}

	response := v1dto.LoginResponseDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresAt,
	}

	utils.ResponseSuccess(c, http.StatusOK, "Refresh token successful", response)
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var input v1dto.ForgotPasswordDTO

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidation(c, validation.HandleValidationErrors(err))
		return
	}

	if err := h.service.ForgotPassword(c, input.Email); err != nil {
		utils.ResponseError(c, err)
		return
	}

	utils.ResponseSuccess(c, http.StatusOK, "Forgot password email sent")
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var input v1dto.ResetPasswordDTO

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidation(c, validation.HandleValidationErrors(err))
		return
	}

	if err := h.service.ResetPassword(c, input.Token, input.Password); err != nil {
		utils.ResponseError(c, err)
		return
	}

	utils.ResponseSuccess(c, http.StatusOK, "Reset password successful")
}
