package infrastructure

import (
	"errors"

	"hublio/internal/platform/crypto"
)

// AESSecretEncryptor implements application.SecretEncryptor using the platform AES-GCM helper.
// The key must be exactly 32 bytes (AES-256), read from CREDENTIAL_ENCRYPTION_KEY.
type AESSecretEncryptor struct {
	key []byte
}

func NewAESSecretEncryptor(key []byte) (*AESSecretEncryptor, error) {
	if len(key) != 32 {
		return nil, errors.New("integration: CREDENTIAL_ENCRYPTION_KEY must be exactly 32 bytes")
	}
	return &AESSecretEncryptor{key: key}, nil
}

func (e *AESSecretEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	ciphertext, err := crypto.EncryptAES(plaintext, e.key)
	if err != nil {
		return nil, err
	}
	return []byte(ciphertext), nil
}

func (e *AESSecretEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	return crypto.DecryptAES(string(ciphertext), e.key)
}
