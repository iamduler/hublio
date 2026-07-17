package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	PrefixLength = 8
	SecretBytes  = 32
)

// GeneratedKey is returned once at creation time. Only Hash should be persisted.
type GeneratedKey struct {
	Prefix    string
	Secret    string // full plaintext key: prefix.secret — never store
	Plaintext string
	Hash      string
}

// Generate creates a new API key. Persist Hash and Prefix only.
func Generate() (GeneratedKey, error) {
	prefixBytes := make([]byte, PrefixLength)
	if _, err := rand.Read(prefixBytes); err != nil {
		return GeneratedKey{}, fmt.Errorf("apikey: generate prefix: %w", err)
	}
	prefix := hex.EncodeToString(prefixBytes)[:PrefixLength]

	secretBytes := make([]byte, SecretBytes)
	if _, err := rand.Read(secretBytes); err != nil {
		return GeneratedKey{}, fmt.Errorf("apikey: generate secret: %w", err)
	}
	secret := base64.RawURLEncoding.EncodeToString(secretBytes)
	plaintext := prefix + "." + secret

	return GeneratedKey{
		Prefix:    prefix,
		Secret:    secret,
		Plaintext: plaintext,
		Hash:      Hash(plaintext),
	}, nil
}

// Hash returns a SHA-256 hex digest of the plaintext API key.
func Hash(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

// PrefixFromPlaintext extracts the key prefix used for lookup.
func PrefixFromPlaintext(plaintext string) (string, bool) {
	parts := strings.SplitN(plaintext, ".", 2)
	if len(parts) != 2 || len(parts[0]) == 0 || len(parts[1]) == 0 {
		return "", false
	}
	return parts[0], true
}
