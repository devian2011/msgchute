package main

import (
	"encoding/base64"
	"encoding/json"
	"net/smtp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/devian2011/msgchute/pkg/shared/provider"
)

func TestSmtpProvider_Configure(t *testing.T) {
	tests := []struct {
		name    string
		params  []byte
		wantErr bool
	}{
		{
			name:    "valid config",
			params:  []byte(`{"host":"smtp.example.com","port":587,"username":"user","password":"pass"}`),
			wantErr: false,
		},
		{
			name:    "valid config with from",
			params:  []byte(`{"host":"smtp.example.com","port":587,"username":"user","password":"pass","from":"sender@example.com"}`),
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			params:  []byte(`{"host":"smtp.example.com","port":587,"username":"user","password":"pass"`),
			wantErr: true,
		},
		{
			name:    "missing host",
			params:  []byte(`{"port":587,"username":"user","password":"pass"}`),
			wantErr: true,
		},
		{
			name:    "missing port",
			params:  []byte(`{"host":"smtp.example.com","username":"user","password":"pass"}`),
			wantErr: true,
		},
		{
			name:    "missing username",
			params:  []byte(`{"host":"smtp.example.com","port":587,"password":"pass"}`),
			wantErr: true,
		},
		{
			name:    "missing password",
			params:  []byte(`{"host":"smtp.example.com","port":587,"username":"user"}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &SmtpProvider{}
			err := p.Configure(tt.params)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSmtpProvider_Send(t *testing.T) {
	// Save original sendMailFunc and restore after test.
	origSendMail := sendMailFunc
	defer func() { sendMailFunc = origSendMail }()

	var capturedFrom string
	var capturedTo []string
	var capturedMsg []byte

	// Mock sendMailFunc to capture arguments and succeed.
	sendMailFunc = func(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
		capturedFrom = from
		capturedTo = to
		capturedMsg = msg
		return nil
	}

	// Setup provider with valid config.
	p := &SmtpProvider{}
	err := p.Configure([]byte(`{"host":"smtp.example.com","port":587,"username":"user","password":"pass","from":"default@example.com"}`))
	require.NoError(t, err)

	tests := []struct {
		name       string
		msg        *provider.Message
		wantErr    bool
		errContain string
		checkFunc  func(t *testing.T, from string, to []string, msg []byte)
	}{
		{
			name: "simple send",
			msg: &provider.Message{
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				Body:    "Test Body",
				Params:  []byte(`{}`),
			},
			wantErr: false,
			checkFunc: func(t *testing.T, from string, to []string, msg []byte) {
				assert.Equal(t, "default@example.com", from)
				assert.Equal(t, []string{"recipient@example.com"}, to)
				body := string(msg)
				assert.Contains(t, body, "From: default@example.com")
				assert.Contains(t, body, "To: recipient@example.com")
				assert.Contains(t, body, "Subject: Test Subject")
				assert.Contains(t, body, "Test Body")
			},
		},
		{
			name: "override from",
			msg: &provider.Message{
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				Body:    "Test Body",
				Params:  []byte(`{"from":"sender@example.com"}`),
			},
			wantErr: false,
			checkFunc: func(t *testing.T, from string, to []string, msg []byte) {
				assert.Equal(t, "sender@example.com", from)
				body := string(msg)
				assert.Contains(t, body, "From: sender@example.com")
			},
		},
		{
			name: "with CC and BCC",
			msg: &provider.Message{
				To:      []string{"to@example.com"},
				Subject: "Test Subject",
				Body:    "Test Body",
				Params:  []byte(`{"cc":["cc@example.com"],"bcc":["bcc@example.com"]}`),
			},
			wantErr: false,
			checkFunc: func(t *testing.T, from string, to []string, msg []byte) {
				assert.Equal(t, "default@example.com", from)
				// All recipients: To + CC + BCC
				assert.ElementsMatch(t, []string{"to@example.com", "cc@example.com", "bcc@example.com"}, to)
				body := string(msg)
				assert.Contains(t, body, "To: to@example.com")
				assert.Contains(t, body, "Cc: cc@example.com")
				assert.NotContains(t, body, "Bcc") // BCC should not appear in headers
			},
		},
		{
			name: "no recipients",
			msg: &provider.Message{
				To:      []string{},
				Subject: "Test",
				Body:    "Test",
				Params:  []byte(`{}`),
			},
			wantErr:    true,
			errContain: "no recipients specified",
		},
		{
			name: "invalid params JSON",
			msg: &provider.Message{
				To:      []string{"to@example.com"},
				Subject: "Test",
				Body:    "Test",
				Params:  []byte(`{"from":"sender@example.com","cc":["cc@example.com"`), // malformed
			},
			wantErr:    true,
			errContain: "failed to parse send options",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := p.Send(tt.msg)
			if tt.wantErr {
				assert.Error(t, resp.Err)
				if tt.errContain != "" {
					assert.Contains(t, resp.Err.Error(), tt.errContain)
				}
				return
			}
			assert.NoError(t, resp.Err)
			if tt.checkFunc != nil {
				tt.checkFunc(t, capturedFrom, capturedTo, capturedMsg)
			}
		})
	}
}

func TestSmtpProvider_buildMessage(t *testing.T) {
	p := &SmtpProvider{}

	baseMsg := &provider.Message{
		To:      []string{"to@example.com"},
		Subject: "Test Subject",
		Body:    "Test Body",
	}

	tests := []struct {
		name      string
		from      string
		msg       *provider.Message
		opts      SendOptions
		wantErr   bool
		checkFunc func(t *testing.T, data []byte)
	}{
		{
			name: "simple text",
			from: "sender@example.com",
			msg:  baseMsg,
			opts: SendOptions{},
			checkFunc: func(t *testing.T, data []byte) {
				body := string(data)
				assert.Contains(t, body, "From: sender@example.com")
				assert.Contains(t, body, "To: to@example.com")
				assert.Contains(t, body, "Subject: Test Subject")
				assert.Contains(t, body, `Content-Type: text/plain; charset="utf-8"`)
				assert.Contains(t, body, "\r\n\r\nTest Body")
				assert.NotContains(t, body, "multipart/mixed")
			},
		},
		{
			name: "with CC",
			from: "sender@example.com",
			msg:  baseMsg,
			opts: SendOptions{CC: []string{"cc@example.com"}},
			checkFunc: func(t *testing.T, data []byte) {
				body := string(data)
				assert.Contains(t, body, "Cc: cc@example.com")
				assert.Contains(t, body, "To: to@example.com")
			},
		},
		{
			name: "with Reply-To",
			from: "sender@example.com",
			msg:  baseMsg,
			opts: SendOptions{ReplyTo: "reply@example.com"},
			checkFunc: func(t *testing.T, data []byte) {
				body := string(data)
				assert.Contains(t, body, "Reply-To: reply@example.com")
			},
		},
		{
			name: "with attachments",
			from: "sender@example.com",
			msg:  baseMsg,
			opts: SendOptions{
				Attachments: []Attachment{
					{
						Filename:    "test.txt",
						Content:     base64.StdEncoding.EncodeToString([]byte("Hello attachment")),
						ContentType: "text/plain",
					},
				},
			},
			checkFunc: func(t *testing.T, data []byte) {
				body := string(data)
				assert.Contains(t, body, `Content-Type: multipart/mixed; boundary="msgchute-boundary"`)
				assert.Contains(t, body, `Content-Disposition: attachment; filename="test.txt"`)
				assert.Contains(t, body, "SGVsbG8gYXR0YWNobWVudA==") // base64 of "Hello attachment"
				assert.Contains(t, body, `Content-Type: text/plain; charset="utf-8"`)
				assert.Contains(t, body, "Test Body")
			},
		},
		{
			name: "with attachment but missing content (skipped)",
			from: "sender@example.com",
			msg:  baseMsg,
			opts: SendOptions{
				Attachments: []Attachment{
					{
						Filename:    "empty.txt",
						Content:     "", // empty content
						ContentType: "text/plain",
					},
				},
			},
			checkFunc: func(t *testing.T, data []byte) {
				body := string(data)
				assert.Contains(t, body, `Content-Type: multipart/mixed; boundary="msgchute-boundary"`)
				assert.Contains(t, body, `Content-Type: text/plain; charset="utf-8"`)
				assert.Contains(t, body, "Test Body")
				assert.NotContains(t, body, "Content-Disposition: attachment")
			},
		},
		{
			name: "with attachment but missing filename (skipped)",
			from: "sender@example.com",
			msg:  baseMsg,
			opts: SendOptions{
				Attachments: []Attachment{
					{
						Filename:    "",
						Content:     base64.StdEncoding.EncodeToString([]byte("content")),
						ContentType: "text/plain",
					},
				},
			},
			checkFunc: func(t *testing.T, data []byte) {
				body := string(data)
				assert.Contains(t, body, `Content-Type: multipart/mixed; boundary="msgchute-boundary"`)
				assert.NotContains(t, body, "Content-Disposition: attachment")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := p.buildMessage(tt.from, tt.msg, tt.opts)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.checkFunc != nil {
				tt.checkFunc(t, data)
			}
		})
	}
}

func TestSmtpProvider_Send_WithAttachment(t *testing.T) {
	origSendMail := sendMailFunc
	defer func() { sendMailFunc = origSendMail }()

	var capturedMsg []byte
	sendMailFunc = func(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
		capturedMsg = msg
		return nil
	}

	p := &SmtpProvider{}
	err := p.Configure([]byte(`{"host":"smtp.example.com","port":587,"username":"user","password":"pass","from":"default@example.com"}`))
	require.NoError(t, err)

	attachmentContent := base64.StdEncoding.EncodeToString([]byte("Hello attachment"))
	paramsJSON, _ := json.Marshal(map[string]interface{}{
		"attachments": []map[string]string{
			{
				"filename":     "file.txt",
				"content":      attachmentContent,
				"content_type": "text/plain",
			},
		},
	})

	msg := &provider.Message{
		To:      []string{"to@example.com"},
		Subject: "Test with attachment",
		Body:    "Email body with attachment",
		Params:  paramsJSON,
	}

	resp := p.Send(msg)
	assert.NoError(t, resp.Err)

	body := string(capturedMsg)
	assert.Contains(t, body, `Content-Type: multipart/mixed; boundary="msgchute-boundary"`)
	assert.Contains(t, body, `Content-Disposition: attachment; filename="file.txt"`)
	assert.Contains(t, body, attachmentContent)
}
