package sms

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/core/intl"
)

var ErrNoAvailableClient = errors.New("no available SMS client")

type SendOptions struct {
	MessageConfig config.SMSMessageConfig
	To            string
	Body          string
}

type RawClient interface {
	Send(from string, to string, body string) error
}

type Client struct {
	Context            context.Context
	MessagingConfig    *config.MessagingConfig
	LocalizationConfig *config.LocalizationConfig
	TwilioClient       *TwilioClient
	NexmoClient        *NexmoClient
}

func (c *Client) Send(opts SendOptions) error {
	var client RawClient
	switch c.MessagingConfig.SMSProvider {
	case config.SMSProviderNexmo:
		if c.NexmoClient == nil {
			return ErrNoAvailableClient
		}
		client = c.NexmoClient
	case config.SMSProviderTwilio:
		if c.TwilioClient == nil {
			return ErrNoAvailableClient
		}
		client = c.TwilioClient
	default:
		return ErrNoAvailableClient
	}

	tags := intl.GetPreferredLanguageTags(c.Context)
	from := intl.LocalizeStringMap(tags, intl.Fallback(c.LocalizationConfig.FallbackLanguage), opts.MessageConfig, "sender")
	return client.Send(from, opts.To, opts.Body)
}
