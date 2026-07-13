package routes

import (
	"shopping-cart/internal/middleware"
	v1routes "shopping-cart/internal/routes/v1"
	"shopping-cart/internal/utils"
	"shopping-cart/pkg/auth"
	"shopping-cart/pkg/cache"

	"net/http"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

type Routes interface {
	Register(r *gin.RouterGroup)
}

/*
- Hàm này dùng để đăng ký tất cả các routes vào engine của Gin
- Miễn là các routes implement interface Routes, thì sẽ được đăng ký vào engine của Gin
*/
func RegisterRoutes(r *gin.Engine, authService auth.TokenService, cacheService cache.RedisCacheService, routes ...Routes) {
	httpLogger := utils.NewLoggerWithPath("access.log", "info")
	recoveryLogger := utils.NewLoggerWithPath("recovery.log", "error")
	rateLimiterLogger := utils.NewLoggerWithPath("rate_limiter.log", "warn")

	r.Use(gzip.Gzip(gzip.DefaultCompression))

	r.Use(
		middleware.RateLimitMiddleware(rateLimiterLogger),
		middleware.CORSMiddleware(),
		middleware.TraceMiddleware(),
		middleware.LoggerMiddleware(httpLogger),
		middleware.RecoveryMiddleware(recoveryLogger),
		middleware.ApiKeyMiddleware(),
	)

	middleware.InitAuthMiddleware(authService, cacheService)

	api := r.Group("/api/v1")

	protected := api.Group("")
	protected.Use(
		middleware.AuthMiddleware(),
	)

	for _, route := range routes {
		switch route.(type) {
		case *v1routes.AuthRoutes:
			route.Register(api)
		default:
			route.Register(protected)
		}
	}

	// Handle 404 Not Found
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found", "path": c.Request.URL.Path})
	})
}
