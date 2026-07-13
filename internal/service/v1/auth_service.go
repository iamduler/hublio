package v1service

import (
	"fmt"
	"shopping-cart/internal/db/sqlc"
	"shopping-cart/internal/repository"
	"shopping-cart/internal/utils"
	"shopping-cart/pkg/auth"
	"shopping-cart/pkg/cache"
	"shopping-cart/pkg/logger"
	"shopping-cart/pkg/mail"
	"shopping-cart/pkg/rabbitmq"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
)

type authService struct {
	userRepo        repository.UserRepository
	tokenService    auth.TokenService
	cacheService    cache.RedisCacheService
	mailService     mail.EmailProviderService
	rabbitMQService rabbitmq.RabbitMQService
}

type LoginAttempt struct {
	limiter     *rate.Limiter
	lastRequest time.Time
}

var (
	mu               sync.Mutex
	clients          = make(map[string]*LoginAttempt)
	LoginAttemptTTL  = 5 * time.Minute
	MaxLoginAttempts = 5
)

func NewAuthService(userRepo repository.UserRepository, tokenService auth.TokenService, cache cache.RedisCacheService, mailService mail.EmailProviderService, rabbitMQService rabbitmq.RabbitMQService) *authService {
	return &authService{
		userRepo:        userRepo,
		tokenService:    tokenService,
		cacheService:    cache,
		mailService:     mailService,
		rabbitMQService: rabbitMQService,
	}
}

func (s *authService) getClientIP(c *gin.Context) string {
	ip := c.ClientIP()

	if ip == "" {
		ip = c.Request.RemoteAddr // Get the client's IP address from the request when the client is using a proxy
	}

	return ip
}

func (s *authService) getLoginAttempt(ip string) *rate.Limiter {
	mu.Lock() // Lock the mutex to prevent race conditions
	defer mu.Unlock()

	client, exists := clients[ip]

	if !exists {
		limiter := rate.NewLimiter(rate.Limit(float32(MaxLoginAttempts)/float32(LoginAttemptTTL.Seconds())), MaxLoginAttempts) // MaxLoginAttempts attempts per LoginAttemptTTL
		newClient := &LoginAttempt{limiter, time.Now()}                                                                        // Create a new client with the limiter and the last request time
		clients[ip] = newClient                                                                                                // Store the new client in the map
		return newClient.limiter
	}

	client.lastRequest = time.Now()
	return client.limiter
}

func (s *authService) shouldLogLoginAttempt(ip string) error {
	limiter := s.getLoginAttempt(ip)

	if !limiter.Allow() {
		return utils.NewError("Too many login attempts", utils.ErrCodeTooManyRequests)
	}

	return nil
}

func (s *authService) CleanUpOldClients(ip string) {
	mu.Lock()
	defer mu.Unlock()

	delete(clients, ip)
}

func (s *authService) Login(ctx *gin.Context, email, password string) (string, string, int, error) {
	context := ctx.Request.Context() // Get the context from the request
	ip := s.getClientIP(ctx)

	if err := s.shouldLogLoginAttempt(ip); err != nil {
		return "", "", 0, err
	}

	// Validate params
	email = utils.NormalizeString(email)

	// Get user by email
	user, err := s.userRepo.GetByEmail(context, email)

	if err != nil {
		s.getLoginAttempt(ip)
		return "", "", 0, utils.NewError("Invalid email or password", utils.ErrCodeUnauthorized)
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	if err != nil {
		s.getLoginAttempt(ip)
		return "", "", 0, utils.NewError("Invalid email or password", utils.ErrCodeUnauthorized)
	}

	// Generate access token
	accessToken, err := s.tokenService.GenerateAccessToken(user)

	if err != nil {
		return "", "", 0, utils.WrapError(err, "Failed to generate access token", utils.ErrCodeInternal)
	}

	// Generate refresh token
	refreshToken, err := s.tokenService.GenerateRefreshToken(user)

	if err != nil {
		return "", "", 0, utils.WrapError(err, "Failed to generate refresh token", utils.ErrCodeInternal)
	}

	// Store refresh token
	if err := s.tokenService.StoreRefreshToken(refreshToken); err != nil {
		return "", "", 0, utils.WrapError(err, "Failed to store refresh token", utils.ErrCodeInternal)
	}

	// Clean up old clients
	s.CleanUpOldClients(ip)

	return accessToken, refreshToken.Token, int(auth.AccessTokenTTL.Seconds()), nil
}

func (s *authService) Logout(ctx *gin.Context, refreshToken string) error {
	authHeader := ctx.GetHeader("Authorization")

	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return utils.NewError("Missing or invalid authorization header", utils.ErrCodeUnauthorized)
	}

	// Access token
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	_, claims, err := s.tokenService.ParseToken(tokenString)

	if err != nil {
		return utils.NewError("Invalid token", utils.ErrCodeUnauthorized)
	}

	if jti, ok := claims["jti"].(string); ok {
		expUnix, _ := claims["exp"].(float64)

		exp := time.Unix(int64(expUnix), 0) // Convert expiration time to Unix timestamp
		key := "token_blacklist:" + jti
		ttl := time.Until(exp)

		s.cacheService.Set(key, "revoked", ttl)
	}

	// Refresh token
	_, err = s.tokenService.ValidateRefreshToken(refreshToken)

	if err != nil {
		return utils.NewError("Expired or invalid refresh token", utils.ErrCodeUnauthorized)
	}

	// Revoke refresh token
	if err := s.tokenService.RevokeRefreshToken(refreshToken); err != nil {
		return utils.WrapError(err, "Failed to revoke refresh token", utils.ErrCodeInternal)
	}

	return nil
}

func (s *authService) RefreshToken(ctx *gin.Context, tokenString string) (string, string, int, error) {
	context := ctx.Request.Context() // Get the context from the request

	// Validate params -> get user uuid
	token, err := s.tokenService.ValidateRefreshToken(tokenString)

	if err != nil {
		return "", "", 0, utils.NewError("Expired or invalid refresh token", utils.ErrCodeUnauthorized)
	}

	// Get user by refresh token
	userUuid, _ := uuid.Parse(token.UserUUID)
	user, err := s.userRepo.GetByUuid(context, userUuid)

	if err != nil {
		return "", "", 0, utils.NewError("User not found", utils.ErrCodeNotFound)
	}

	// Generate access token
	accessToken, err := s.tokenService.GenerateAccessToken(user)

	if err != nil {
		return "", "", 0, utils.WrapError(err, "Failed to generate access token", utils.ErrCodeInternal)
	}

	// Generate refresh token
	refreshToken, err := s.tokenService.GenerateRefreshToken(user)

	if err != nil {
		return "", "", 0, utils.WrapError(err, "Failed to generate refresh token", utils.ErrCodeInternal)
	}

	// Revoke refresh token
	if err := s.tokenService.RevokeRefreshToken(tokenString); err != nil {
		return "", "", 0, utils.WrapError(err, "Failed to revoke refresh token", utils.ErrCodeInternal)
	}

	// Store refresh token
	if err := s.tokenService.StoreRefreshToken(refreshToken); err != nil {
		return "", "", 0, utils.WrapError(err, "Failed to store refresh token", utils.ErrCodeInternal)
	}

	return accessToken, refreshToken.Token, int(auth.AccessTokenTTL.Seconds()), nil
}

func (s *authService) ForgotPassword(ctx *gin.Context, email string) error {
	context := ctx.Request.Context() // Get the context from the request

	rateLimitKey := fmt.Sprintf("rate_limit:forgot_password:%s", email)

	if exists, err := s.cacheService.Exists(rateLimitKey); err == nil && exists {
		return utils.NewError("Too many requests for forgot password", utils.ErrCodeTooManyRequests)
	}

	// Validate params
	email = utils.NormalizeString(email)

	// Get user by email
	user, err := s.userRepo.GetByEmail(context, email)

	if err != nil {
		return utils.NewError("User not found", utils.ErrCodeNotFound)
	}

	token, err := utils.GenerateRandomString(16)

	if err != nil {
		return utils.NewError("Failed to generate random string", utils.ErrCodeInternal)
	}

	// Store rate limit
	if err := s.cacheService.Set(rateLimitKey, token, 5*time.Minute); err != nil {
		return utils.WrapError(err, "Failed to store rate limit", utils.ErrCodeInternal)
	}

	// Store reset password token
	if err := s.cacheService.Set("reset_password:"+token, user.Uuid.String(), 15*time.Minute); err != nil {
		return utils.WrapError(err, "Failed to store reset password", utils.ErrCodeInternal)
	}

	// Send email
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", utils.GetEnv("FRONTEND_URL", "http://localhost:3000"), token)

	// Log reset link
	logger.Log.Info().Msgf("✅ Reset link: %s", resetLink)

	// Send email
	sampleBody := `
		<p>Hello {{.FullName}},</p>
		<p>You are receiving this email because you requested a password reset for your account.</p>
		<p>Click the button below to reset your password:</p>
		<a href="{{.ResetLink}}">Reset Password</a>
		<p>If you did not request a password reset, please ignore this email.</p>
		<p>Thank you,</p>
		<p>The Shopping Cart Team</p>
	`

	body := strings.ReplaceAll(sampleBody, "{{.FullName}}", user.FullName)
	body = strings.ReplaceAll(body, "{{.ResetLink}}", resetLink)

	mailContent := &mail.Email{
		To: []mail.Address{{
			Name:  user.FullName,
			Email: user.Email,
		}},
		Subject:  "Reset Password",
		Text:     body,
		Category: "reset_password",
	}

	if err := s.rabbitMQService.Publish(context, "auth_email_queue", mailContent); err != nil {
		return utils.NewError("Failed to publish email to RabbitMQ", utils.ErrCodeInternal)
	}

	// if err := s.mailService.SendMail(context, mailContent); err != nil {
	// 	return utils.NewError("Failed to send email", utils.ErrCodeInternal)
	// }

	return nil
}

func (s *authService) ResetPassword(ctx *gin.Context, token, password string) error {
	context := ctx.Request.Context() // Get the context from the request

	// Get user by reset password token
	var userUuidString string
	err := s.cacheService.Get("reset_password:"+token, &userUuidString)

	if err == redis.Nil || userUuidString == "" {
		return utils.NewError("Invalid or expired reset password token", utils.ErrCodeUnauthorized)
	}

	if err != nil {
		return utils.WrapError(err, "Failed to get user by reset password token", utils.ErrCodeInternal)
	}

	userUuid, err := uuid.Parse(userUuidString)

	if err != nil {
		return utils.WrapError(err, "Failed to parse user UUID", utils.ErrCodeInternal)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return utils.WrapError(err, "Failed to hash password", utils.ErrCodeInternal)
	}

	// Update password
	input := sqlc.UpdatePasswordParams{
		Password: string(hashedPassword),
		Uuid:     userUuid,
	}

	_, err = s.userRepo.UpdatePassword(context, input)

	if err != nil {
		return utils.WrapError(err, "Failed to update password", utils.ErrCodeInternal)
	}

	// Clear cache
	if err := s.cacheService.Delete("reset_password:" + token); err != nil {
		logger.Log.Error().Err(err).Msg("❌ Failed to clear cache")
	}

	return nil
}
