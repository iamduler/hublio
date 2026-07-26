package infrastructure_test

import (
	"testing"
	"time"

	"hublio/internal/identity/infrastructure"

	"github.com/pquerna/otp/totp"
)

func TestTOTPAdapter_VerifyAcceptsOneStepOfSkew(t *testing.T) {
	adapter := infrastructure.NewTOTPAdapter()

	provisioned, err := adapter.Generate("user@example.com")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if provisioned.Secret == "" {
		t.Fatal("expected a secret")
	}

	now := time.Date(2026, 7, 26, 10, 30, 0, 0, time.UTC)
	code, err := totp.GenerateCode(provisioned.Secret, now)
	if err != nil {
		t.Fatalf("generate code: %v", err)
	}

	tests := []struct {
		name string
		at   time.Time
		code string
		want bool
	}{
		{name: "current step", at: now, code: code, want: true},
		{name: "one step early", at: now.Add(-30 * time.Second), code: code, want: true},
		{name: "one step late", at: now.Add(30 * time.Second), code: code, want: true},
		{name: "two steps late", at: now.Add(90 * time.Second), code: code, want: false},
		{name: "wrong code", at: now, code: "000000", want: false},
		{name: "empty code", at: now, code: "", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := adapter.Verify(provisioned.Secret, tc.code, tc.at); got != tc.want {
				t.Fatalf("Verify() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAESMFASecretCipher_RoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{name: "raw 32 bytes", key: "dev-only-insecure-32-byte-key!!!"},
		{name: "64 hex chars", key: "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"},
		{name: "base64", key: "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8="},
		{name: "too short", key: "short-key", wantErr: true},
		{name: "empty", key: "", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cipher, err := infrastructure.NewAESMFASecretCipher(tc.key)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error for an unusable key")
				}
				return
			}
			if err != nil {
				t.Fatalf("new cipher: %v", err)
			}

			const secret = "JBSWY3DPEHPK3PXP"
			ciphertext, err := cipher.Encrypt(secret)
			if err != nil {
				t.Fatalf("encrypt: %v", err)
			}
			if ciphertext == secret {
				t.Fatal("ciphertext must not equal the plaintext secret")
			}
			plaintext, err := cipher.Decrypt(ciphertext)
			if err != nil {
				t.Fatalf("decrypt: %v", err)
			}
			if plaintext != secret {
				t.Fatalf("round trip = %q, want %q", plaintext, secret)
			}
		})
	}
}
