package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"shopping-cart/internal/db/sqlc"
	"shopping-cart/internal/utils"
	"shopping-cart/pkg/cache"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {
	cache cache.RedisCacheService
}

type EncrytedPayload struct {
	UserUUID string `json:"user_uuid"`
	Email    string `json:"email"`
	Role     int32  `json:"role"`
}

type RefreshToken struct {
	Token     string    `json:"token"`
	UserUUID  string    `json:"user_uuid"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
}

var (
	jwtSecretKey  = []byte(utils.GetEnv("JWT_SECRET_KEY", "shopping-cart-jwt-secret-key"))
	jwtEncryptKey = []byte(utils.GetEnv("JWT_ENCRYPT_KEY", "12345678901234567890123456789012"))
)

const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
)

func NewJWTService(cache cache.RedisCacheService) TokenService {
	return &JWTService{
		cache: cache,
	}
}

// =-=-=-=-=-=-=-=-=-=-=-=-= ACCESS TOKEN =-=-=-=-=-=-=-=-=-=-=-=-=
func (s *JWTService) GenerateAccessToken(user sqlc.User) (string, error) {
	payload := EncrytedPayload{
		UserUUID: user.Uuid.String(),
		Email:    user.Email,
		Role:     int32(user.Level),
	}

	rawData, err := json.Marshal(payload)

	if err != nil {
		return "", err
	}

	encryptedData, err := utils.EncryptAES(rawData, jwtEncryptKey)

	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"data": encryptedData,
		"jti":  uuid.NewString(),
		"exp":  time.Now().Add(AccessTokenTTL).Unix(),
		"iat":  time.Now().Unix(),
		"iss":  "shopping-cart",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecretKey)
}

func (s *JWTService) ParseToken(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return jwtSecretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, nil, utils.NewError("Invalid token", utils.ErrCodeUnauthorized)
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return nil, nil, utils.NewError("Invalid token claims", utils.ErrCodeUnauthorized)
	}

	return token, claims, nil
}

func (s *JWTService) DecryptAccessTokenPayload(tokenString string) (*EncrytedPayload, error) {
	_, claims, err := s.ParseToken(tokenString)

	if err != nil {
		return nil, utils.WrapError(err, "Failed to parse token", utils.ErrCodeInternal)
	}

	encryptedData, ok := claims["data"].(string)

	if !ok {
		return nil, utils.NewError("Invalid token payload", utils.ErrCodeUnauthorized)
	}

	decryptetBytes, err := utils.DecryptAES(encryptedData, jwtEncryptKey)

	if err != nil {
		return nil, utils.WrapError(err, "Failed to decrypt token payload", utils.ErrCodeInternal)
	}

	var payload EncrytedPayload

	if err = json.Unmarshal(decryptetBytes, &payload); err != nil {
		return nil, utils.WrapError(err, "Failed to unmarshal token payload", utils.ErrCodeInternal)
	}

	return &payload, nil
}

// =-=-=-=-=-=-=-=-=-=-=-=-= REFRESH TOKEN =-=-=-=-=-=-=-=-=-=-=-=-=
func (s *JWTService) GenerateRefreshToken(user sqlc.User) (RefreshToken, error) {
	tokenBytes := make([]byte, 32)

	if _, err := rand.Read(tokenBytes); err != nil {
		return RefreshToken{}, err
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)

	return RefreshToken{
		Token:     token,
		UserUUID:  user.Uuid.String(),
		ExpiresAt: time.Now().Add(RefreshTokenTTL),
		Revoked:   false,
	}, nil
}

func (s *JWTService) StoreRefreshToken(token RefreshToken) error {
	cacheKey := "refresh_token:" + token.Token
	return s.cache.Set(cacheKey, token, RefreshTokenTTL)
}

func (s *JWTService) ValidateRefreshToken(token string) (RefreshToken, error) {
	cacheKey := "refresh_token:" + token
	var storedToken RefreshToken

	err := s.cache.Get(cacheKey, &storedToken)

	if err != nil || storedToken.Revoked || storedToken.ExpiresAt.Before(time.Now()) {
		return RefreshToken{}, utils.NewError("Invalid refresh token", utils.ErrCodeUnauthorized)
	}

	return storedToken, nil
}

func (s *JWTService) RevokeRefreshToken(token string) error {
	cacheKey := "refresh_token:" + token
	var refreshToken RefreshToken

	err := s.cache.Get(cacheKey, &refreshToken)

	if err != nil {
		return utils.WrapError(err, "Failed to get refresh token", utils.ErrCodeInternal)
	}

	refreshToken.Revoked = true
	return s.cache.Set(cacheKey, refreshToken, time.Until(refreshToken.ExpiresAt))
}
