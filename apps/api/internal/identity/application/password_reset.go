package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"hublio/internal/identity/domain"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/logging"

	"github.com/google/uuid"
)

const passwordResetTTL = time.Hour

type passwordResetPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// RequestPasswordReset issues a one-time reset token for a password user and emails the reset
// link. It is intentionally anti-enumeration: the caller always sees a generic success, so this
// method never reports whether the email exists, lacks a password (OAuth-only), or is inactive.
func (s *Services) RequestPasswordReset(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil
	}

	user, err := s.Users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return mapRepoErr(err)
	}

	// OAuth-only or inactive users cannot reset a password: stay silent (anti-enumeration).
	if !user.HasPassword() || !user.CanLogin() {
		return nil
	}

	token, err := s.storePasswordReset(passwordResetPayload{
		UserID: user.ID().String(),
		Email:  user.Email(),
	})
	if err != nil {
		// Do not leak infrastructure failures as differentiated responses.
		if logging.Log != nil {
			logging.Log.Error().Err(err).Msg("identity: failed to store password reset token")
		}
		return nil
	}

	link := fmt.Sprintf("%s/reset-password?token=%s", s.webAppURL(), token)
	subject := "Reset your Hublio password"
	body := fmt.Sprintf(
		"We received a request to reset your Hublio password.\n\n"+
			"Reset it using the link below (valid for 1 hour):\n%s\n\n"+
			"If you did not request this, you can safely ignore this email.",
		link,
	)

	// Best-effort delivery: the mail service logs its own failures; never differentiate here.
	_ = s.mailer().Send(ctx, user.Email(), subject, body)
	return nil
}

// ResetPasswordInput carries the one-time token and the new plaintext password.
type ResetPasswordInput struct {
	Token    string
	Password string
}

// ResetPassword consumes a one-time reset token and sets a new password hash. The caller owns
// the transaction boundary.
func (s *Services) ResetPassword(ctx context.Context, in ResetPasswordInput) error {
	if strings.TrimSpace(in.Password) == "" || len(in.Password) < 8 {
		return apperr.New("password must be at least 8 characters", apperr.ErrCodeBadRequest)
	}

	payload, err := s.consumePasswordReset(in.Token)
	if err != nil {
		return err
	}

	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return apperr.New("invalid or expired reset token", apperr.ErrCodeBadRequest)
	}

	user, err := s.Users.FindByID(ctx, userID)
	if err != nil {
		return mapRepoErr(err)
	}

	hash, err := s.Passwords.Hash(in.Password)
	if err != nil {
		return apperr.Wrap(err, "failed to hash password", apperr.ErrCodeInternal)
	}
	if err := user.SetPasswordHash(hash, s.clock().Now()); err != nil {
		return mapDomainErr(err)
	}
	if err := s.Users.Update(ctx, user); err != nil {
		return mapRepoErr(err)
	}
	return nil
}

func (s *Services) storePasswordReset(payload passwordResetPayload) (string, error) {
	if s.Cache == nil {
		return "", apperr.New("reset token store unavailable", apperr.ErrCodeInternal)
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", apperr.Wrap(err, "failed to generate reset token", apperr.ErrCodeInternal)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	if err := s.Cache.Set(passwordResetCacheKey(token), payload, passwordResetTTL); err != nil {
		return "", apperr.Wrap(err, "failed to store reset token", apperr.ErrCodeInternal)
	}
	return token, nil
}

func (s *Services) consumePasswordReset(token string) (*passwordResetPayload, error) {
	if s.Cache == nil {
		return nil, apperr.New("reset token store unavailable", apperr.ErrCodeInternal)
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, apperr.New("invalid or expired reset token", apperr.ErrCodeBadRequest)
	}
	key := passwordResetCacheKey(token)
	var payload passwordResetPayload
	if err := s.Cache.Get(key, &payload); err != nil {
		return nil, apperr.New("invalid or expired reset token", apperr.ErrCodeBadRequest)
	}
	_ = s.Cache.Delete(key)
	return &payload, nil
}

func passwordResetCacheKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return "password_reset:" + hex.EncodeToString(sum[:])
}
