package email

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

// Config holds SMTP email configuration
type Config struct {
	SMTPHost  string `yaml:"smtp_host"`
	SMTPPort  int    `yaml:"smtp_port"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	FromEmail string `yaml:"from_email"`
	FromName  string `yaml:"from_name"`
	UseTLS    bool   `yaml:"use_tls"`
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

// Send sends an email using SMTP
func Send(cfg *Config, email *Email) error {
	if cfg.SMTPHost == "" {
		return fmt.Errorf("SMTP host not configured")
	}

	if len(email.To) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	// Build the email message
	message, err := buildMessage(cfg, email)
	if err != nil {
		return fmt.Errorf("failed to build message: %w", err)
	}

	// Connect to SMTP server
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SMTPHost)
	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)

	// Collect all recipients
	recipients := append([]string{}, email.To...)
	recipients = append(recipients, email.Cc...)
	recipients = append(recipients, email.Bcc...)

	// Send with TLS if enabled
	if cfg.UseTLS {
		return sendWithTLS(addr, auth, cfg.FromEmail, recipients, message)
	}

	// Send without TLS
	return smtp.SendMail(addr, auth, cfg.FromEmail, recipients, message)
}

// sendWithTLS sends email with TLS encryption
func sendWithTLS(addr string, auth smtp.Auth, from string, to []string, message []byte) error {
	// Connect to server
	host := strings.Split(addr, ":")[0]

	// TLS config
	tlsConfig := &tls.Config{
		ServerName: host,
	}

	// Connect with TLS
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Authenticate
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
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
		return fmt.Errorf("failed to initiate data transfer: %w", err)
	}

	_, err = writer.Write(message)
	if err != nil {
		writer.Close()
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return nil
}

// buildMessage constructs the email message with headers and body
func buildMessage(cfg *Config, email *Email) ([]byte, error) {
	var buf bytes.Buffer

	// Create multipart writer
	writer := multipart.NewWriter(&buf)
	boundary := writer.Boundary()

	// Write headers
	buf.WriteString(fmt.Sprintf("From: %s <%s>\r\n", cfg.FromName, cfg.FromEmail))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(email.To, ", ")))

	if len(email.Cc) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(email.Cc, ", ")))
	}

	if email.ReplyTo != "" {
		buf.WriteString(fmt.Sprintf("Reply-To: %s\r\n", email.ReplyTo))
	}

	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", boundary))
	buf.WriteString("\r\n")

	// Write text/HTML body
	if email.HTMLBody != "" {
		// Create alternative part for text and HTML
		altWriter := multipart.NewWriter(&buf)
		altBoundary := altWriter.Boundary()

		partHeader := make(textproto.MIMEHeader)
		partHeader.Set("Content-Type", fmt.Sprintf("multipart/alternative; boundary=%s", altBoundary))

		if _, err := writer.CreatePart(partHeader); err != nil {
			return nil, err
		}

		// Text part
		textHeader := make(textproto.MIMEHeader)
		textHeader.Set("Content-Type", "text/plain; charset=utf-8")
		textHeader.Set("Content-Transfer-Encoding", "quoted-printable")

		textPart, err := altWriter.CreatePart(textHeader)
		if err != nil {
			return nil, err
		}
		textPart.Write([]byte(email.Body))

		// HTML part
		htmlHeader := make(textproto.MIMEHeader)
		htmlHeader.Set("Content-Type", "text/html; charset=utf-8")
		htmlHeader.Set("Content-Transfer-Encoding", "quoted-printable")

		htmlPart, err := altWriter.CreatePart(htmlHeader)
		if err != nil {
			return nil, err
		}
		htmlPart.Write([]byte(email.HTMLBody))

		altWriter.Close()
	} else {
		// Plain text only
		partHeader := make(textproto.MIMEHeader)
		partHeader.Set("Content-Type", "text/plain; charset=utf-8")
		partHeader.Set("Content-Transfer-Encoding", "quoted-printable")

		part, err := writer.CreatePart(partHeader)
		if err != nil {
			return nil, err
		}
		part.Write([]byte(email.Body))
	}

	// Add attachments
	for _, filePath := range email.Attachments {
		if err := addAttachment(writer, filePath); err != nil {
			return nil, fmt.Errorf("failed to add attachment %s: %w", filePath, err)
		}
	}

	writer.Close()

	return buf.Bytes(), nil
}

// addAttachment adds a file attachment to the email
func addAttachment(writer *multipart.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	// Create attachment header
	filename := filepath.Base(filePath)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Type", "application/octet-stream")
	header.Set("Content-Transfer-Encoding", "base64")
	header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	part, err := writer.CreatePart(header)
	if err != nil {
		return err
	}

	// Read and encode file
	buf := make([]byte, fileInfo.Size())
	if _, err := file.Read(buf); err != nil && err != io.EOF {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(buf)

	// Write in 76-character lines (MIME standard)
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		part.Write([]byte(encoded[i:end] + "\r\n"))
	}

	return nil
}

// ValidateConfig checks if email configuration is valid
func ValidateConfig(cfg *Config) error {
	if cfg.SMTPHost == "" {
		return fmt.Errorf("SMTP host is required")
	}
	if cfg.SMTPPort == 0 {
		return fmt.Errorf("SMTP port is required")
	}
	if cfg.Username == "" {
		return fmt.Errorf("username is required")
	}
	if cfg.Password == "" {
		return fmt.Errorf("password is required")
	}
	if cfg.FromEmail == "" {
		return fmt.Errorf("from email is required")
	}
	return nil
}

// TestConnection tests the SMTP connection
func TestConnection(cfg *Config) error {
	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SMTPHost)

	// Try to connect and authenticate
	if cfg.UseTLS {
		host := strings.Split(addr, ":")[0]
		tlsConfig := &tls.Config{ServerName: host}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}
		defer client.Quit()

		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
	} else {
		client, err := smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}
		defer client.Quit()

		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
	}

	return nil
}
