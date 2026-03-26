package notifications

import (
	"fmt"
	"time"

	gomail "gopkg.in/gomail.v2"
)

// MailerConfig holds SMTP connection settings.
type MailerConfig struct {
	Host     string
	Port     int
	From     string
	FromName string
}

// Mailer sends emails via SMTP using gomail.
type Mailer struct {
	cfg    MailerConfig
	dialer *gomail.Dialer
}

// NewMailer creates a new Mailer instance.
func NewMailer(cfg MailerConfig) *Mailer {
	d := gomail.NewDialer(cfg.Host, cfg.Port, "", "")
	return &Mailer{cfg: cfg, dialer: d}
}

// Send composes and delivers an HTML email with a plain-text fallback.
// It retries up to 3 times with exponential back-off (1s → 2s → 4s).
func (m *Mailer) Send(to, subject, htmlBody, textBody string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", fmt.Sprintf("%s <%s>", m.cfg.FromName, m.cfg.From))
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", textBody)
	msg.AddAlternative("text/html", htmlBody)

	var lastErr error
	backoff := time.Second
	for attempt := 1; attempt <= 3; attempt++ {
		if err := m.dialer.DialAndSend(msg); err != nil {
			lastErr = err
			time.Sleep(backoff)
			backoff *= 2
			continue
		}
		return nil
	}
	return fmt.Errorf("mailer: all retries exhausted sending to %s: %w", to, lastErr)
}
