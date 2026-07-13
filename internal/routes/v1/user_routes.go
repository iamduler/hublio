package v1routes

import (
	v1handler "shopping-cart/internal/handler/v1"

	"github.com/gin-gonic/gin"
)

type UserRoutes struct {
	handler *v1handler.UserHandler
}

func NewUserRoutes(handler *v1handler.UserHandler) *UserRoutes {
	return &UserRoutes{
		handler: handler,
	}
}

func (r *UserRoutes) Register(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		users.GET("", r.handler.GetAllUsers)
		users.POST("", r.handler.CreateUser)
		users.GET("/:uuid", r.handler.GetUserByUuid)
		users.GET("/soft-deleted", r.handler.GetSoftDeletedUsers)
		users.PUT("/:uuid", r.handler.UpdateUser)

		users.DELETE("/:uuid", r.handler.SoftDeleteUser)
		users.PUT("/:uuid/restore", r.handler.RestoreUser)
		users.DELETE("/:uuid/trash", r.handler.DeleteUser)
	}
}
