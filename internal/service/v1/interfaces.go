package v1service

import (
	"shopping-cart/internal/db/sqlc"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserService interface {
	GetAllUsers(ctx *gin.Context, search, orderBy, sort string, page, limit int32, isDeleted bool) ([]sqlc.User, int64, error)
	CreateUser(ctx *gin.Context, params sqlc.CreateUserParams) (sqlc.User, error)
	GetUserByUuid(ctx *gin.Context, uuid uuid.UUID) (sqlc.User, error)
	UpdateUser(ctx *gin.Context, params sqlc.UpdateUserParams) (sqlc.User, error)
	SoftDeleteUser(ctx *gin.Context, uuid uuid.UUID) (sqlc.User, error)
	RestoreUser(ctx *gin.Context, uuid uuid.UUID) (sqlc.User, error)
	DeleteUser(ctx *gin.Context, uuid uuid.UUID) error
}

type AuthService interface {
	Login(ctx *gin.Context, email, password string) (string, string, int, error)
	Logout(ctx *gin.Context, tokenString string) error
	RefreshToken(ctx *gin.Context, tokenString string) (string, string, int, error)
	ForgotPassword(ctx *gin.Context, email string) error
	ResetPassword(ctx *gin.Context, email, password string) error
}
