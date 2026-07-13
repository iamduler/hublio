package v1routes

import (
	v1handler "shopping-cart/internal/handler/v1"

	"github.com/gin-gonic/gin"
)

type AuthRoutes struct {
	handler *v1handler.AuthHandler
}

func NewAuthRoutes(handler *v1handler.AuthHandler) *AuthRoutes {
	return &AuthRoutes{
		handler: handler,
	}
}

func (r *AuthRoutes) Register(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.POST("/login", r.handler.Login)
		auth.POST("/logout", r.handler.Logout)
		auth.POST("/refresh", r.handler.RefreshToken)
		auth.POST("/forgot-password", r.handler.ForgotPassword)
		auth.POST("/reset-password", r.handler.ResetPassword)
	}
}
