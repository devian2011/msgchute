package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/hashicorp/go-plugin"

	"github.com/devian2011/msgchute/pkg/shared/provider"
)

// SmtpConfig holds the SMTP server settings. Provided during Configure.
type SmtpConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from,omitempty"` // can be overridden in Send
}

// Attachment represents a file attachment.
type Attachment struct {
	Filename    string `json:"filename"`               // name of the attached file
	Content     string `json:"content"`                // base64-encoded content
	ContentType string `json:"content_type,omitempty"` // MIME type, optional
}

// SendOptions are additional options for sending a message, passed in params.
type SendOptions struct {
	From        string       `json:"from,omitempty"`
	CC          []string     `json:"cc,omitempty"`
	BCC         []string     `json:"bcc,omitempty"`
	ReplyTo     string       `json:"reply_to,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// SmtpProvider implements the provider.Provider interface for SMTP.
type SmtpProvider struct {
	config SmtpConfig
}

// Configure parses the plugin configuration.
func (s *SmtpProvider) Configure(params []byte) error {
	var config SmtpConfig
	if err := sonic.Unmarshal(params, &config); err != nil {
		return fmt.Errorf("failed to parse smtp config: %w", err)
	}
	// Validate required fields
	if len(config.Host) == 0 || len(config.Port) == 0 || len(config.Username) == 0 || len(config.Password) == 0 {
		return fmt.Errorf("missing required smtp config fields (host, port, username, password)")
	}
	s.config = config
	return nil
}

// Send sends an email via SMTP with optional attachments.
// Returns *provider.MessageResponse as required by the interface.
func (s *SmtpProvider) Send(msg *provider.Message) *provider.MessageResponse {
	// Parse additional options from params
	var opts SendOptions
	if len(msg.Params) > 0 {
		if err := sonic.Unmarshal(msg.Params, &opts); err != nil {
			return &provider.MessageResponse{
				Err:        fmt.Errorf("failed to parse send options: %w", err),
				IsCritical: true,
			}
		}
	}

	// Determine sender
	from := s.config.From
	if opts.From != "" {
		from = opts.From
	}
	if from == "" {
		from = s.config.Username // fallback
	}

	// Collect all recipients (To + CC + BCC)
	allRecipients := append([]string{}, msg.To...)
	if len(opts.CC) > 0 {
		allRecipients = append(allRecipients, opts.CC...)
	}
	if len(opts.BCC) > 0 {
		allRecipients = append(allRecipients, opts.BCC...)
	}
	if len(allRecipients) == 0 {
		return &provider.MessageResponse{
			Err:        fmt.Errorf("no recipients specified"),
			IsCritical: true,
		}
	}

	// Build the email message (plain text or multipart)
	emailBody, err := s.buildMessage(from, msg, opts)
	if err != nil {
		return &provider.MessageResponse{
			Err:        fmt.Errorf("failed to build email message: %w", err),
			IsCritical: true,
		}
	}

	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)

	tlsConfig := &tls.Config{
		ServerName: s.config.Host,
	}
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return &provider.MessageResponse{
			Err:        fmt.Errorf("TLS connection failed: %w", err),
			IsCritical: true,
		}
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return &provider.MessageResponse{
			Err:        fmt.Errorf("SMTP client creation failed: %w", err),
			IsCritical: true,
		}
	}
	defer client.Quit()

	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	if err := client.Auth(auth); err != nil {
		return &provider.MessageResponse{
			Err:        fmt.Errorf("SMTP auth failed: %w", err),
			IsCritical: true,
		}
	}

	if err := client.Mail(from); err != nil {
		return &provider.MessageResponse{
			Err:        fmt.Errorf("MAIL FROM failed: %w", err),
			IsCritical: true,
		}
	}
	for _, rcpt := range allRecipients {
		if err := client.Rcpt(rcpt); err != nil {
			return &provider.MessageResponse{
				Err:        fmt.Errorf("RCPT TO %s failed: %w", rcpt, err),
				IsCritical: true,
			}
		}
	}

	w, err := client.Data()
	if err != nil {
		return &provider.MessageResponse{
			Err:        fmt.Errorf("DATA command failed: %w", err),
			IsCritical: true,
		}
	}
	if _, err = w.Write(emailBody); err != nil {
		return &provider.MessageResponse{
			Err:        fmt.Errorf("failed to write email body: %w", err),
			IsCritical: true,
		}
	}
	if err = w.Close(); err != nil {
		return &provider.MessageResponse{
			Err:        fmt.Errorf("failed to close email body writer: %w", err),
			IsCritical: true,
		}
	}

	// Success
	return &provider.MessageResponse{
		Err:      nil,
		Response: "Email sent successfully",
	}
}

// buildMessage constructs the full email message (headers + body) as a byte slice.
// If there are attachments, it builds a multipart/mixed message.
func (s *SmtpProvider) buildMessage(from string, msg *provider.Message, opts SendOptions) ([]byte, error) {
	var buf bytes.Buffer

	// Prepare headers
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = strings.Join(msg.To, ", ")
	if len(opts.CC) > 0 {
		headers["Cc"] = strings.Join(opts.CC, ", ")
	}
	if opts.ReplyTo != "" {
		headers["Reply-To"] = opts.ReplyTo
	}
	headers["Subject"] = msg.Subject
	headers["MIME-Version"] = "1.0"

	hasAttachments := len(opts.Attachments) > 0
	boundary := "msgchute-boundary"

	if hasAttachments {
		headers["Content-Type"] = `multipart/mixed; boundary="` + boundary + `"`

		for k, v := range headers {
			fmt.Fprintf(&buf, "%s: %s\r\n", k, v)
		}
		buf.WriteString("\r\n")

		mw := multipart.NewWriter(&buf)
		mw.SetBoundary(boundary)

		textPart, err := mw.CreatePart(textPlainHeader())
		if err != nil {
			return nil, err
		}
		if _, err := textPart.Write([]byte(msg.Body)); err != nil {
			return nil, err
		}

		for _, att := range opts.Attachments {
			if att.Filename == "" || att.Content == "" {
				continue
			}
			contentType := att.ContentType
			if contentType == "" {
				contentType = "application/octet-stream"
			}
			decoded, err := base64.StdEncoding.DecodeString(att.Content)
			if err != nil {
				return nil, fmt.Errorf("failed to decode attachment '%s': %w", att.Filename, err)
			}

			partHeaders := textproto.MIMEHeader{}
			partHeaders.Set("Content-Type", contentType)
			// Исправлено: явные кавычки и кодировка имени файла
			encodedFilename := mime.QEncoding.Encode("utf-8", att.Filename)
			partHeaders.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, encodedFilename))
			partHeaders.Set("Content-Transfer-Encoding", "base64")

			part, err := mw.CreatePart(partHeaders)
			if err != nil {
				return nil, err
			}
			encoder := base64.NewEncoder(base64.StdEncoding, part)
			if _, err := encoder.Write(decoded); err != nil {
				return nil, err
			}
			if err := encoder.Close(); err != nil {
				return nil, err
			}
		}

		if err := mw.Close(); err != nil {
			return nil, err
		}
	} else {
		headers["Content-Type"] = `text/plain; charset="utf-8"`
		for k, v := range headers {
			fmt.Fprintf(&buf, "%s: %s\r\n", k, v)
		}
		buf.WriteString("\r\n")
		buf.WriteString(msg.Body)
	}

	return buf.Bytes(), nil
}

// textPlainHeader returns MIME headers for a plain text part.
func textPlainHeader() textproto.MIMEHeader {
	h := textproto.MIMEHeader{}
	h.Set("Content-Type", `text/plain; charset="utf-8"`)
	h.Set("Content-Transfer-Encoding", "quoted-printable")
	return h
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: provider.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"provider": &provider.ProviderPlugin{Provider: &SmtpProvider{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
