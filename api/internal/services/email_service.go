package services

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
)

// EmailConfig holds SMTP configuration
type EmailConfig struct {
	SMTPHost  string
	SMTPPort  string
	Username  string
	Password  string
	FromEmail string
	FromName  string
	UseTLS    bool
}

// Email represents an email message
type Email struct {
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	Body        string
	HTMLBody    string
	Attachments []string
	ReplyTo     string
}

// EmailService handles email sending
type EmailService struct {
	config *EmailConfig
}

// NewEmailService creates a new email service
func NewEmailService(config *EmailConfig) *EmailService {
	return &EmailService{
		config: config,
	}
}

// Send sends an email
func (s *EmailService) Send(email *Email) error {
	// Build message
	message, err := s.buildMessage(email)
	if err != nil {
		return fmt.Errorf("failed to build message: %w", err)
	}

	// Setup authentication
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.SMTPHost)

	// Combine all recipients
	recipients := append([]string{}, email.To...)
	recipients = append(recipients, email.Cc...)
	recipients = append(recipients, email.Bcc...)

	// Send email
	addr := s.config.SMTPHost + ":" + s.config.SMTPPort
	from := s.config.FromEmail

	if s.config.UseTLS {
		return s.sendWithTLS(addr, auth, from, recipients, message)
	}

	return smtp.SendMail(addr, auth, from, recipients, message)
}

// sendWithTLS sends email with TLS encryption
func (s *EmailService) sendWithTLS(addr string, auth smtp.Auth, from string, to []string, message []byte) error {
	// Create TLS config
	tlsConfig := &tls.Config{
		ServerName: s.config.SMTPHost,
	}

	// Connect to server
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.config.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Authenticate
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Set sender
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to add recipient %s: %w", recipient, err)
		}
	}

	// Send message
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = writer.Write(message)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return nil
}

// buildMessage builds the email message with headers and body
func (s *EmailService) buildMessage(email *Email) ([]byte, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Headers
	from := fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = strings.Join(email.To, ", ")
	if len(email.Cc) > 0 {
		headers["Cc"] = strings.Join(email.Cc, ", ")
	}
	headers["Subject"] = email.Subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = fmt.Sprintf("multipart/mixed; boundary=%s", writer.Boundary())
	if email.ReplyTo != "" {
		headers["Reply-To"] = email.ReplyTo
	}

	// Write headers
	for key, value := range headers {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	buf.WriteString("\r\n")

	// Text/HTML body
	if email.HTMLBody != "" && email.Body != "" {
		// Create alternative part for HTML and plain text
		altWriter := multipart.NewWriter(&buf)
		altBoundary := altWriter.Boundary()

		headers := textproto.MIMEHeader{}
		headers.Set("Content-Type", fmt.Sprintf("multipart/alternative; boundary=%s", altBoundary))
		if _, err := writer.CreatePart(headers); err != nil {
			return nil, err
		}

		// Plain text part
		buf.WriteString(fmt.Sprintf("\r\n--%s\r\n", altBoundary))
		buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		buf.WriteString(email.Body)

		// HTML part
		buf.WriteString(fmt.Sprintf("\r\n--%s\r\n", altBoundary))
		buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		buf.WriteString(email.HTMLBody)

		buf.WriteString(fmt.Sprintf("\r\n--%s--\r\n", altBoundary))
	} else if email.HTMLBody != "" {
		// HTML only
		part, err := writer.CreatePart(textproto.MIMEHeader{
			"Content-Type": []string{"text/html; charset=UTF-8"},
		})
		if err != nil {
			return nil, err
		}
		part.Write([]byte(email.HTMLBody))
	} else {
		// Plain text only
		part, err := writer.CreatePart(textproto.MIMEHeader{
			"Content-Type": []string{"text/plain; charset=UTF-8"},
		})
		if err != nil {
			return nil, err
		}
		part.Write([]byte(email.Body))
	}

	// Attachments
	for _, attachment := range email.Attachments {
		if err := s.addAttachment(writer, attachment); err != nil {
			return nil, fmt.Errorf("failed to add attachment %s: %w", attachment, err)
		}
	}

	writer.Close()

	return buf.Bytes(), nil
}

// addAttachment adds a file attachment to the email
func (s *EmailService) addAttachment(writer *multipart.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	filename := filepath.Base(filePath)

	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type":              []string{"application/octet-stream"},
		"Content-Transfer-Encoding": []string{"base64"},
		"Content-Disposition":       []string{fmt.Sprintf(`attachment; filename="%s"`, filename)},
	})
	if err != nil {
		return err
	}

	// Read file and encode to base64
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(data)

	// Write in 76-character lines (RFC 2045)
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		part.Write([]byte(encoded[i:end] + "\r\n"))
	}

	return nil
}

// LoadConfigFromEnv loads email config from environment variables
func LoadConfigFromEnv() *EmailConfig {
	return &EmailConfig{
		SMTPHost:  getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:  getEnv("SMTP_PORT", "587"),
		Username:  getEnv("SMTP_USERNAME", ""),
		Password:  getEnv("SMTP_PASSWORD", ""),
		FromEmail: getEnv("SMTP_FROM_EMAIL", ""),
		FromName:  getEnv("SMTP_FROM_NAME", "UNG Billing"),
		UseTLS:    getEnv("SMTP_USE_TLS", "true") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
