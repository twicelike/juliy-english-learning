package mail

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/english/backend/internal/config"
)

type Mailer interface {
	SendEmail(to []string, subject string, htmlBody string) error
}

type smtpMailer struct {
	host     string
	port     int
	user     string
	password string
}

func NewSMTPMailer(cfg *config.Config) Mailer {
	return &smtpMailer{
		host:     cfg.SMTPHost,
		port:     cfg.SMTPPort,
		user:     cfg.SMTPUser,
		password: cfg.SMTPPassword,
	}
}

func (m *smtpMailer) SendEmail(to []string, subject string, htmlBody string) error {
	// If no user is provided, we can fallback to just printing it out (Mock behavior)
	if m.user == "" || m.password == "" {
		fmt.Printf("\n--- MOCK EMAIL ---\nTo: %v\nSubject: %s\nBody: %s\n------------------\n", to, subject, htmlBody)
		return nil
	}

	auth := smtp.PlainAuth("", m.user, m.password, m.host)

	addr := fmt.Sprintf("%s:%d", m.host, m.port)

	msg := []byte("To: " + strings.Join(to, ",") + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
		htmlBody + "\r\n")

	return smtp.SendMail(addr, auth, m.user, to, msg)
}
