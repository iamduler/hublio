package mail

import (
	"context"
	"hublio/internal/platform/config"
	"hublio/internal/platform/apperr"
	"hublio/internal/platform/logging"
	"time"

	"github.com/rs/zerolog"
)

type Email struct {
	From     Address   `json:"from"`
	To       []Address `json:"to"`
	Subject  string    `json:"subject"`
	Text     string    `json:"text"`
	Category string    `json:"category"`
}

type Address struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

type MailConfig struct {
	ProviderConfig map[string]any
	ProviderType   ProviderType
	MaxRetries     int
	Timeout        time.Duration
	Logger         *zerolog.Logger
}

type MailService struct {
	provider EmailProviderService
	config   *MailConfig
	logger   *zerolog.Logger
}

func NewMailService(cfg *config.Config, logger *zerolog.Logger, providerFactory ProviderFactory) (EmailProviderService, error) {
	config := &MailConfig{
		ProviderConfig: cfg.MailProviderConfig,
		ProviderType:   ProviderType(cfg.MailProviderType),
		MaxRetries:     3,
		Timeout:        10 * time.Second,
		Logger:         logger,
	}

	provider, err := providerFactory.CreateProvider(config)

	if err != nil {
		return nil, err
	}

	return &MailService{
		provider: provider,
		config:   config,
		logger:   logger,
	}, nil
}

func (m *MailService) SendMail(ctx context.Context, email *Email) error {
	traceId := logging.GetTraceID(ctx)
	start := time.Now()

	var lastError error

	logInstance := m.logger.With().Str("trace_id", traceId).
		Str("action", "send_mail").
		Interface("to", email.To).
		Interface("from", email.From).
		Str("subject", email.Subject).
		Str("text", email.Text).
		Str("category", email.Category).
		Logger()

	for i := 0; i < m.config.MaxRetries; i++ {
		startAttempt := time.Now()
		err := m.provider.SendMail(ctx, email)

		if err == nil {
			logInstance.Info().Dur("duration", time.Since(startAttempt)).Int("attempt", i+1).Msg("Mail sent successfully")
			return nil
		}

		lastError = err
		logInstance.Warn().Dur("duration", time.Since(startAttempt)).Int("attempt", i+1).Err(err).Msg("Failed to send mail, retrying...")

		time.Sleep(time.Duration(i+1) * time.Second)
	}

	logInstance.Error().Dur("duration", time.Since(start)).Err(lastError).Msg("Failed to send mail after all retries")

	return apperr.Wrap(lastError, "Failed to send mail after all retries", apperr.ErrCodeInternal)
}
