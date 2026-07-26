package middleware

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSenitizeLoggedBody_RedactsSensitiveFields(t *testing.T) {
	fields := []string{"password", "secret", "otpauth_url", "recovery_codes", "access_token"}

	tests := []struct {
		name       string
		body       string
		mustRedact []string
		mustKeep   []string
	}{
		{
			name:       "mfa setup response",
			body:       `{"data":{"secret":"JBSWY3DPEHPK3PXP","otpauth_url":"otpauth://totp/x","recovery_codes":["abcd-efgh-ijkl"],"warning":"store now"}}`,
			mustRedact: []string{"JBSWY3DPEHPK3PXP", "otpauth://totp/x", "abcd-efgh-ijkl"},
			mustKeep:   []string{"store now"},
		},
		{
			name:       "login response",
			body:       `{"data":{"access_token":"jwt-value","user":{"email":"a@b.c"}}}`,
			mustRedact: []string{"jwt-value"},
			mustKeep:   []string{"a@b.c"},
		},
		{
			name:     "error envelope keeps its code",
			body:     `{"error":"invalid mfa code","code":"UNAUTHORIZED"}`,
			mustKeep: []string{"UNAUTHORIZED", "invalid mfa code"},
		},
		{
			name:       "array of objects",
			body:       `[{"password":"hunter2"},{"name":"keep"}]`,
			mustRedact: []string{"hunter2"},
			mustKeep:   []string{"keep"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var parsed any
			if err := json.Unmarshal([]byte(tc.body), &parsed); err != nil {
				t.Fatal(err)
			}
			out, err := json.Marshal(senitizeLoggedBody(parsed, fields))
			if err != nil {
				t.Fatal(err)
			}
			got := string(out)

			for _, secret := range tc.mustRedact {
				if strings.Contains(got, secret) {
					t.Fatalf("logged body still contains %q: %s", secret, got)
				}
			}
			for _, keep := range tc.mustKeep {
				if !strings.Contains(got, keep) {
					t.Fatalf("logged body dropped %q: %s", keep, got)
				}
			}
		})
	}
}
