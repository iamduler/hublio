package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/cache"
	"hublio/internal/platform/crypto"
	"hublio/internal/platform/env"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenSubject is the platform-level identity carried in access tokens.
// Domain-specific user models must not leak into this package.
type TokenSubject struct {
	UserID         string
	Email          string
	Role           string
	OrganizationID string
}

type JWTService struct {
	cache cache.RedisCacheService
}

type EncryptedPayload struct {
	UserID         string `json:"user_id"`
	Email          string `json:"email"`
	Role           string `json:"role"`
	OrganizationID string `json:"organization_id"`
}

type RefreshToken struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
}

var (
	jwtSecretKey  = []byte(env.GetEnv("JWT_SECRET_KEY", "hublio-jwt-secret-key"))
	jwtEncryptKey = []byte(env.GetEnv("JWT_ENCRYPT_KEY", "12345678901234567890123456789012"))
)

const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
)

func NewJWTService(cacheService cache.RedisCacheService) TokenService {
	return &JWTService{
		cache: cacheService,
	}
}

func (s *JWTService) GenerateAccessToken(subject TokenSubject) (string, error) {
	payload := EncryptedPayload{
		UserID:         subject.UserID,
		Email:          subject.Email,
		Role:           subject.Role,
		OrganizationID: subject.OrganizationID,
	}

	rawData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	encryptedData, err := crypto.EncryptAES(rawData, jwtEncryptKey)
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"data": encryptedData,
		"jti":  uuid.NewString(),
		"exp":  time.Now().Add(AccessTokenTTL).Unix(),
		"iat":  time.Now().Unix(),
		"iss":  "hublio",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecretKey)
}

func (s *JWTService) ParseToken(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return jwtSecretKey, nil
	})
	if err != nil || !token.Valid {
		return nil, nil, apperr.New("Invalid token", apperr.ErrCodeUnauthorized)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, apperr.New("Invalid token claims", apperr.ErrCodeUnauthorized)
	}

	return token, claims, nil
}

func (s *JWTService) DecryptAccessTokenPayload(tokenString string) (*EncryptedPayload, error) {
	_, claims, err := s.ParseToken(tokenString)
	if err != nil {
		return nil, apperr.Wrap(err, "Failed to parse token", apperr.ErrCodeInternal)
	}

	encryptedData, ok := claims["data"].(string)
	if !ok {
		return nil, apperr.New("Invalid token payload", apperr.ErrCodeUnauthorized)
	}

	decryptedBytes, err := crypto.DecryptAES(encryptedData, jwtEncryptKey)
	if err != nil {
		return nil, apperr.Wrap(err, "Failed to decrypt token payload", apperr.ErrCodeInternal)
	}

	var payload EncryptedPayload
	if err = json.Unmarshal(decryptedBytes, &payload); err != nil {
		return nil, apperr.Wrap(err, "Failed to unmarshal token payload", apperr.ErrCodeInternal)
	}

	return &payload, nil
}

func (s *JWTService) GenerateRefreshToken(subject TokenSubject) (RefreshToken, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return RefreshToken{}, err
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)

	return RefreshToken{
		Token:     token,
		UserID:    subject.UserID,
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
		return RefreshToken{}, apperr.New("Invalid refresh token", apperr.ErrCodeUnauthorized)
	}

	return storedToken, nil
}

func (s *JWTService) RevokeRefreshToken(token string) error {
	cacheKey := "refresh_token:" + token
	var refreshToken RefreshToken

	err := s.cache.Get(cacheKey, &refreshToken)
	if err != nil {
		return apperr.Wrap(err, "Failed to get refresh token", apperr.ErrCodeInternal)
	}

	refreshToken.Revoked = true
	return s.cache.Set(cacheKey, refreshToken, time.Until(refreshToken.ExpiresAt))
}
