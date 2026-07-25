package domain

import (
	"crypto/hmac"
)

// WebhookSecretHeader is the HTTP header Nhanh (and other origins) must send.
const WebhookSecretHeader = "X-Hublio-Webhook-Secret"

// MatchWebhookSecret compares expected vs provided secrets in constant time.
// Never log either value.
func MatchWebhookSecret(expected, provided string) bool {
	if expected == "" || provided == "" {
		return false
	}
	return hmac.Equal([]byte(expected), []byte(provided))
}
