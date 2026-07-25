package apikey_test

import (
	"testing"

	"hublio/internal/platform/apikey"
)

func TestGenerateAndHash(t *testing.T) {
	key, err := apikey.Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if key.Prefix == "" || key.Plaintext == "" || key.Hash == "" {
		t.Fatal("expected non-empty prefix, plaintext, and hash")
	}

	if got := apikey.Hash(key.Plaintext); got != key.Hash {
		t.Fatalf("Hash mismatch: got %s want %s", got, key.Hash)
	}

	prefix, ok := apikey.PrefixFromPlaintext(key.Plaintext)
	if !ok || prefix != key.Prefix {
		t.Fatalf("PrefixFromPlaintext: got %q ok=%v want %q", prefix, ok, key.Prefix)
	}
}

func TestStaticAuthenticator(t *testing.T) {
	key, err := apikey.Generate()
	if err != nil {
		t.Fatal(err)
	}

	auth := apikey.NewStaticAuthenticator(key.Plaintext)
	principal, err := auth.Authenticate(t.Context(), key.Plaintext)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if principal.Name != "bootstrap" {
		t.Fatalf("unexpected principal name %q", principal.Name)
	}

	if _, err := auth.Authenticate(t.Context(), "bad.key"); err == nil {
		t.Fatal("expected unauthorized")
	}
}

func TestStubAuthenticator(t *testing.T) {
	auth := apikey.NewStubAuthenticator()
	if _, err := auth.Authenticate(t.Context(), "anything"); err == nil {
		t.Fatal("expected unauthorized")
	}
}
