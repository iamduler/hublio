package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
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

const (
	emailVerifyTTL      = 15 * time.Minute
	emailVerifyCodeSize = 6
)

type emailVerifyPayload struct {
	Code   string `json:"code"`
	UserID string `json:"user_id"`
}

// RequestEmailVerification generates (or refreshes) a 6-digit verification code for a user whose
// email is not yet verified and emails it. Anti-enumeration: the caller always sees a generic
// success, so this never reveals whether the email exists or is already verified.
func (s *Services) RequestEmailVerification(ctx context.Context, email string) error {
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
	if user.EmailVerifiedAt() != nil {
		return nil
	}

	code, err := generateNumericCode(emailVerifyCodeSize)
	if err != nil {
		if logging.Log != nil {
			logging.Log.Error().Err(err).Msg("identity: failed to generate email verification code")
		}
		return nil
	}

	if err := s.storeEmailVerify(email, emailVerifyPayload{Code: code, UserID: user.ID().String()}); err != nil {
		if logging.Log != nil {
			logging.Log.Error().Err(err).Msg("identity: failed to store email verification code")
		}
		return nil
	}

	subject := "Verify your Hublio email"
	body := fmt.Sprintf(
		"Your Hublio email verification code is: %s\n\n"+
			"It is valid for 15 minutes. If you did not request this, you can ignore this email.",
		code,
	)

	// Best-effort delivery: the mail service logs its own failures.
	_ = s.mailer().Send(ctx, user.Email(), subject, body)
	return nil
}

// VerifyEmail validates a submitted code against the stored one (constant-time), marks the email
// verified and consumes the code. The caller owns the transaction boundary.
func (s *Services) VerifyEmail(ctx context.Context, email, code string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	code = strings.TrimSpace(code)
	if email == "" || code == "" {
		return apperr.New("invalid or expired verification code", apperr.ErrCodeBadRequest)
	}

	payload, err := s.peekEmailVerify(email)
	if err != nil {
		return err
	}
	if subtle.ConstantTimeCompare([]byte(payload.Code), []byte(code)) != 1 {
		return apperr.New("invalid or expired verification code", apperr.ErrCodeBadRequest)
	}

	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return apperr.New("invalid or expired verification code", apperr.ErrCodeBadRequest)
	}
	user, err := s.Users.FindByID(ctx, userID)
	if err != nil {
		return mapRepoErr(err)
	}

	user.MarkEmailVerified(s.clock().Now())
	if err := s.Users.Update(ctx, user); err != nil {
		return mapRepoErr(err)
	}
	s.deleteEmailVerify(email)
	return nil
}

func (s *Services) storeEmailVerify(email string, payload emailVerifyPayload) error {
	if s.Cache == nil {
		return apperr.New("verification store unavailable", apperr.ErrCodeInternal)
	}
	if err := s.Cache.Set(emailVerifyCacheKey(email), payload, emailVerifyTTL); err != nil {
		return apperr.Wrap(err, "failed to store verification code", apperr.ErrCodeInternal)
	}
	return nil
}

func (s *Services) peekEmailVerify(email string) (*emailVerifyPayload, error) {
	if s.Cache == nil {
		return nil, apperr.New("verification store unavailable", apperr.ErrCodeInternal)
	}
	var payload emailVerifyPayload
	if err := s.Cache.Get(emailVerifyCacheKey(email), &payload); err != nil {
		return nil, apperr.New("invalid or expired verification code", apperr.ErrCodeBadRequest)
	}
	return &payload, nil
}

func (s *Services) deleteEmailVerify(email string) {
	if s.Cache == nil {
		return
	}
	_ = s.Cache.Delete(emailVerifyCacheKey(email))
}

func emailVerifyCacheKey(email string) string {
	sum := sha256.Sum256([]byte(email))
	return "email_verify:" + hex.EncodeToString(sum[:])
}

// generateNumericCode returns an n-digit numeric code using a cryptographically secure source.
func generateNumericCode(n int) (string, error) {
	const digits = "0123456789"
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, n)
	for i := range buf {
		out[i] = digits[int(buf[i])%len(digits)]
	}
	return string(out), nil
}
