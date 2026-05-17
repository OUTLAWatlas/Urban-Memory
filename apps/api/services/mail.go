package services

import (
	"fmt"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
)

const (
	defaultSMTPHost = "smtp.gmail.com"
	defaultSMTPPort = "587"
)

type MailSender interface {
	SendEmail(to string, subject string, body string) error
}

type SMTPMailer struct {
	host        string
	port        string
	fromEmail   string
	password    string
	disableAuth bool
}

func NewSMTPMailerFromEnv() (*SMTPMailer, error) {
	fromEmail := strings.TrimSpace(os.Getenv("SMTP_EMAIL"))
	if fromEmail == "" {
		return nil, fmt.Errorf("SMTP_EMAIL is required")
	}
	if _, err := mail.ParseAddress(fromEmail); err != nil {
		return nil, fmt.Errorf("invalid SMTP_EMAIL: %w", err)
	}

	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	if host == "" {
		host = defaultSMTPHost
	}

	disableAuth := parseBoolEnv("SMTP_DISABLE_AUTH")
	if host == "127.0.0.1" || strings.EqualFold(host, "localhost") {
		// Local SMTP sinks (MailHog/Mailpit) typically don't support AUTH.
		disableAuth = true
	}

	password := strings.TrimSpace(os.Getenv("SMTP_APP_PASSWORD"))
	// Google app passwords are commonly shown in grouped blocks; normalize pasted values.
	password = strings.ReplaceAll(password, " ", "")
	if password == "" && !disableAuth {
		return nil, fmt.Errorf("SMTP_APP_PASSWORD is required")
	}

	port := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	if port == "" {
		port = defaultSMTPPort
	}

	return &SMTPMailer{
		host:        host,
		port:        port,
		fromEmail:   fromEmail,
		password:    password,
		disableAuth: disableAuth,
	}, nil
}

func (m *SMTPMailer) SendEmail(to string, subject string, body string) error {
	if m == nil {
		return fmt.Errorf("mail sender is not configured")
	}

	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("recipient email is required")
	}
	if _, err := mail.ParseAddress(to); err != nil {
		return fmt.Errorf("invalid recipient email: %w", err)
	}

	subject = strings.TrimSpace(subject)
	if subject == "" {
		return fmt.Errorf("email subject is required")
	}
	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("email body is required")
	}

	var auth smtp.Auth
	if !m.disableAuth {
		auth = smtp.PlainAuth("", m.fromEmail, m.password, m.host)
	}
	message := buildSMTPMessage(m.fromEmail, to, subject, body)
	targetAddress := fmt.Sprintf("%s:%s", m.host, m.port)

	// Diagnostic logging to help debug silent failures in development.
	// Avoid printing sensitive values like the SMTP password.
	logMsg := fmt.Sprintf("sending smtp email to %s via %s (from=%s, auth_disabled=%v)", to, targetAddress, m.fromEmail, m.disableAuth)
	fmt.Println(logMsg)

	if err := smtp.SendMail(targetAddress, auth, m.fromEmail, []string{to}, []byte(message)); err != nil {
		errMsg := fmt.Errorf("send smtp email: %w", err)
		fmt.Println("smtp send failed:", errMsg)
		return errMsg
	}

	fmt.Println("smtp send success:", to, "via", targetAddress)
	return nil
}

func parseBoolEnv(key string) bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	return raw == "1" || raw == "true" || raw == "yes" || raw == "on"
}

func buildSMTPMessage(from string, to string, subject string, body string) string {
	headers := []string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=\"utf-8\"",
		"Content-Transfer-Encoding: 8bit",
	}

	return strings.Join(headers, "\r\n") + "\r\n\r\n" + body
}
