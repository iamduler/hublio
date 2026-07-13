package app

import (
	v1handler "shopping-cart/internal/handler/v1"
	"shopping-cart/internal/repository"
	"shopping-cart/internal/routes"
	v1routes "shopping-cart/internal/routes/v1"
	v1service "shopping-cart/internal/service/v1"
)

type UserModule struct {
	routes routes.Routes
}

func NewUserModule(ctx *ModuleContext) *UserModule {
	// Initialize repository
	userRepo := repository.NewSqlUserRepository(ctx.DB)

	// Initialize service
	userService := v1service.NewUserService(userRepo, ctx.Redis)

	// Initialize handler
	userHandler := v1handler.NewUserHandler(userService)

	// Initialize routes
	userRoutes := v1routes.NewUserRoutes(userHandler)

	return &UserModule{
		routes: userRoutes,
	}
}

func (m *UserModule) Routes() routes.Routes {
	return m.routes
}
