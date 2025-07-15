package game

import (
	"context"
	"fmt"
	"time"

	"github.com/mailgun/mailgun-go/v4"

	"panzerstadt/async-multiplayer/config"
)

type Notifier interface {
	Notify(recipientEmail string, subject string, body string) error
}

type MailgunNotifier struct {
	mg     *mailgun.MailgunImpl
	domain string
}

func NewMailgunNotifier(cfg config.Config) *MailgunNotifier {
	mg := mailgun.NewMailgun(cfg.MailgunDomain, cfg.MailgunAPIKey)
	return &MailgunNotifier{
		mg:     mg,
		domain: cfg.MailgunDomain,
	}
}

func (m *MailgunNotifier) Notify(recipientEmail string, subject string, body string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	sender := "Async Multiplayer <noreply@async-multiplayer.com>"

	message := m.mg.NewMessage(sender, subject, body, recipientEmail)

	_, _, err := m.mg.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	fmt.Printf("Email sent to %s: %s\n", recipientEmail, subject)
	return nil
}
