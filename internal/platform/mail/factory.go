package mail

import (
	"fmt"
	"hublio/internal/platform/apperr"
)

type ProviderType string

const (
	ProviderMailtrap ProviderType = "mailtrap"
)

type ProviderFactory interface {
	CreateProvider(config *MailConfig) (EmailProviderService, error)
}

type MailtrapProviderFactory struct {
}

func (f *MailtrapProviderFactory) CreateProvider(config *MailConfig) (EmailProviderService, error) {
	return NewMailtrapProvider(config)
}

func NewProviderFactory(providerType ProviderType) (ProviderFactory, error) {
	switch providerType {
	case ProviderMailtrap:
		return &MailtrapProviderFactory{}, nil
	default:
		return nil, apperr.New(fmt.Sprintf("provider type not supported: %s", providerType), apperr.ErrCodeInternal)
	}
}
