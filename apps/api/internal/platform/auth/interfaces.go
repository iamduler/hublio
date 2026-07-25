package auth

import "github.com/golang-jwt/jwt/v5"

type TokenService interface {
	GenerateAccessToken(subject TokenSubject) (string, error)
	GenerateRefreshToken(subject TokenSubject) (RefreshToken, error)
	StoreRefreshToken(token RefreshToken) error
	ValidateRefreshToken(token string) (RefreshToken, error)
	ParseToken(tokenString string) (*jwt.Token, jwt.MapClaims, error)
	DecryptAccessTokenPayload(tokenString string) (*EncryptedPayload, error)
	RevokeRefreshToken(token string) error
}
