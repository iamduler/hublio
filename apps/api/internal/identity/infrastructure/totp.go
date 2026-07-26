package infrastructure

import (
	"time"

	"hublio/internal/identity/application"

	"github.com/pquerna/otp/totp"
)

// totpIssuer labels the account inside authenticator apps.
const totpIssuer = "Hublio"

// TOTPAdapter implements application.TOTPVerifier with RFC 6238 (SHA1, 6 digits, 30s period),
// the profile every mainstream authenticator app supports.
type TOTPAdapter struct{}

func NewTOTPAdapter() *TOTPAdapter {
	return &TOTPAdapter{}
}

func (TOTPAdapter) Generate(accountName string) (application.TOTPSecret, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      totpIssuer,
		AccountName: accountName,
	})
	if err != nil {
		return application.TOTPSecret{}, err
	}
	return application.TOTPSecret{
		Secret:     key.Secret(),
		OTPAuthURL: key.URL(),
	}, nil
}

// Verify accepts the current 30s step plus one step of clock skew on either side.
func (TOTPAdapter) Verify(secret, code string, at time.Time) bool {
	ok, err := totp.ValidateCustom(code, secret, at.UTC(), totp.ValidateOpts{
		Period: 30,
		Skew:   1,
		Digits: 6,
	})
	return err == nil && ok
}

var _ application.TOTPVerifier = (*TOTPAdapter)(nil)
