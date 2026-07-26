package application_test

import (
	"context"
	"strings"

	"hublio/internal/identity/application"
)

// sentMail captures a single Send call for assertions.
type sentMail struct {
	to      string
	subject string
	body    string
}

// recordingMailer records outbound mail instead of sending it.
type recordingMailer struct {
	sent []sentMail
}

func (m *recordingMailer) Send(ctx context.Context, to, subject, body string) error {
	_ = ctx
	m.sent = append(m.sent, sentMail{to: to, subject: subject, body: body})
	return nil
}

var _ application.Mailer = (*recordingMailer)(nil)

// countPrefix returns the number of cache keys that start with the given prefix.
func (m *memCache) countPrefix(prefix string) int {
	n := 0
	for k := range m.data {
		if strings.HasPrefix(k, prefix) {
			n++
		}
	}
	return n
}
