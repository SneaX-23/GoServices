package services

import (
	"log/slog"

	"github.com/resend/resend-go/v3"
)

type Mailer interface {
	Send(to, subject, htmlBody string) error
}

type ResendMailer struct {
	client *resend.Client
	from   string
}

// initializes the Resend client
func NewResendMailer(apiKey, from string) *ResendMailer {
	if apiKey == "" {
		slog.Warn("Missing apikey", "api key", apiKey)
	}

	client := resend.NewClient(apiKey)

	return &ResendMailer{
		client: client,
		from:   from,
	}
}

func (m *ResendMailer) Send(to, subject, htmlBody string) error {
	params := &resend.SendEmailRequest{
		From:    m.from,
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
	}
	_, err := m.client.Emails.Send(params)

	return err
}
