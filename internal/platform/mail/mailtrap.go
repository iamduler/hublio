package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/logging"
	"time"

	"github.com/rs/zerolog"
)

type MailtrapConfig struct {
	MailSender string
	NameSender string
	ApiKey     string
	ApiUrl     string
}

type MailtrapProvider struct {
	client *http.Client
	config *MailtrapConfig
	logger *zerolog.Logger
}

func NewMailtrapProvider(config *MailConfig) (EmailProviderService, error) {
	mailtrapConfig, ok := config.ProviderConfig["mailtrap"].(map[string]any)

	if !ok {
		return nil, apperr.New("mailtrap config not found", apperr.ErrCodeInternal)
	}

	return &MailtrapProvider{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		config: &MailtrapConfig{
			MailSender: mailtrapConfig["mail_sender"].(string),
			NameSender: mailtrapConfig["name_sender"].(string),
			ApiKey:     mailtrapConfig["api_key"].(string),
			ApiUrl:     mailtrapConfig["api_url"].(string),
		},
		logger: config.Logger,
	}, nil
}

func (m *MailtrapProvider) SendMail(ctx context.Context, email *Email) error {
	traceId := logging.GetTraceID(ctx)
	start := time.Now()

	time.Sleep(10 * time.Second)

	email.From = Address{
		Name:  m.config.NameSender,
		Email: m.config.MailSender,
	}

	payload, err := json.Marshal(email)

	if err != nil {
		return apperr.Wrap(err, "Failed to marshal email", apperr.ErrCodeInternal)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, m.config.ApiUrl, bytes.NewReader(payload))

	if err != nil {
		return apperr.Wrap(err, "Failed to create request", apperr.ErrCodeInternal)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+m.config.ApiKey)

	// Send request
	response, err := m.client.Do(request)

	if err != nil {
		m.logger.Error().Err(err).
			Str("trace_id", traceId).
			Dur("duration", time.Since(start)).
			Str("action", "send_mail").
			Str("url", m.config.ApiUrl).
			Msg("Failed to send request")

		return apperr.Wrap(err, "Failed to send request", apperr.ErrCodeInternal)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)

		m.logger.Error().
			Str("trace_id", traceId).
			Dur("duration", time.Since(start)).
			Str("action", "send_mail").
			Str("url", m.config.ApiUrl).
			Int("status_code", response.StatusCode).
			Str("response_body", string(body)).
			Msg("Unexpected response status code")

		return apperr.New(fmt.Sprintf("Unexpected response status code %d: %s", response.StatusCode, string(body)), apperr.ErrCodeBadGateway)
	}

	return nil
}
