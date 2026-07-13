package app

import (
	v1handler "shopping-cart/internal/handler/v1"
	"shopping-cart/internal/repository"
	"shopping-cart/internal/routes"
	v1routes "shopping-cart/internal/routes/v1"
	v1service "shopping-cart/internal/service/v1"
	"shopping-cart/pkg/auth"
	"shopping-cart/pkg/cache"
	"shopping-cart/pkg/mail"
	"shopping-cart/pkg/rabbitmq"
)

type AuthModule struct {
	routes routes.Routes
}

func NewAuthModule(ctx *ModuleContext, tokenService auth.TokenService, cache cache.RedisCacheService, mailService mail.EmailProviderService, rabbitMQService rabbitmq.RabbitMQService) *AuthModule {
	// Initialize repository
	userRepo := repository.NewSqlUserRepository(ctx.DB)

	// Initialize service
	authService := v1service.NewAuthService(userRepo, tokenService, cache, mailService, rabbitMQService)

	// Initialize handler
	authHandler := v1handler.NewAuthHandler(authService)

	// Initialize routes
	authRoutes := v1routes.NewAuthRoutes(authHandler)

	return &AuthModule{
		routes: authRoutes,
	}
}

func (m *AuthModule) Routes() routes.Routes {
	return m.routes
}
