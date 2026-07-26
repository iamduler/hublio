package infrastructure

import (
	"encoding/base64"
	"encoding/hex"
	"errors"

	"hublio/internal/identity/application"
	"hublio/internal/platform/crypto"
)

// AESMFASecretCipher implements application.MFASecretCipher using the platform AES-GCM helper.
// The key must resolve to exactly 32 bytes (AES-256), read from MFA_ENCRYPTION_KEY.
type AESMFASecretCipher struct {
	key []byte
}

// NewAESMFASecretCipher accepts the key as 64 hex characters, standard base64, or 32 raw bytes.
func NewAESMFASecretCipher(key string) (*AESMFASecretCipher, error) {
	decoded, err := decodeEncryptionKey(key)
	if err != nil {
		return nil, err
	}
	return &AESMFASecretCipher{key: decoded}, nil
}

func (c *AESMFASecretCipher) Encrypt(plaintext string) (string, error) {
	return crypto.EncryptAES([]byte(plaintext), c.key)
}

func (c *AESMFASecretCipher) Decrypt(ciphertext string) (string, error) {
	plaintext, err := crypto.DecryptAES(ciphertext, c.key)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func decodeEncryptionKey(key string) ([]byte, error) {
	if raw, err := hex.DecodeString(key); err == nil && len(raw) == 32 {
		return raw, nil
	}
	if raw, err := base64.StdEncoding.DecodeString(key); err == nil && len(raw) == 32 {
		return raw, nil
	}
	if len(key) == 32 {
		return []byte(key), nil
	}
	return nil, errors.New("identity: MFA_ENCRYPTION_KEY must be 32 bytes (raw, 64 hex chars, or base64)")
}

var _ application.MFASecretCipher = (*AESMFASecretCipher)(nil)
