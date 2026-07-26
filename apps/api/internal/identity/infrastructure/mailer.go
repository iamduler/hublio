package infrastructure

import (
	"context"

	"hublio/internal/platform/mail"
)

// MailerAdapter adapts the platform mail service to the identity application's Mailer port so
// the Application layer never depends on provider/transport concerns.
type MailerAdapter struct {
	svc mail.EmailProviderService
}

func NewMailerAdapter(svc mail.EmailProviderService) *MailerAdapter {
	return &MailerAdapter{svc: svc}
}

func (a *MailerAdapter) Send(ctx context.Context, to, subject, body string) error {
	return a.svc.SendMail(ctx, &mail.Email{
		To:       []mail.Address{{Email: to}},
		Subject:  subject,
		Text:     body,
		Category: "identity",
	})
}
