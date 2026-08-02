package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/devian2011/msgchute/pkg/shared/provider"
)

func TestTgProvider_Configure(t *testing.T) {
	// Save original newBotAPI and restore after test
	origNewBotAPI := newBotAPI
	defer func() { newBotAPI = origNewBotAPI }()

	tests := []struct {
		name    string
		params  []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config",
			params:  []byte(`{"token":"123456:ABC-DEF"}`),
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			params:  []byte(`{"token":"123456:ABC-DEF"`),
			wantErr: true,
			errMsg:  "failed to parse telegram config",
		},
		{
			name:    "missing token",
			params:  []byte(`{}`),
			wantErr: true,
			errMsg:  "missing required field: token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock for each test
			newBotAPI = func(token string) (*tgbotapi.BotAPI, error) {
				return &tgbotapi.BotAPI{}, nil
			}
			if tt.name == "valid config" {
				// for valid config we want to return a mock bot
				newBotAPI = func(token string) (*tgbotapi.BotAPI, error) {
					return &tgbotapi.BotAPI{}, nil
				}
			}
			p := &TgProvider{}
			err := p.Configure(tt.params)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, p.bot)
				assert.NotNil(t, p.sendFunc)
			}
		})
	}
}

func TestTgProvider_Send(t *testing.T) {
	// Save original newBotAPI and restore
	origNewBotAPI := newBotAPI
	defer func() { newBotAPI = origNewBotAPI }()

	// Mock bot creation to succeed
	newBotAPI = func(token string) (*tgbotapi.BotAPI, error) {
		return &tgbotapi.BotAPI{}, nil
	}

	var sentMessages []tgbotapi.Chattable
	var lastError error
	mockSend := func(msg tgbotapi.Chattable) (tgbotapi.Message, error) {
		sentMessages = append(sentMessages, msg)
		return tgbotapi.Message{}, lastError
	}

	p := &TgProvider{
		sendFunc: mockSend,
	}
	err := p.Configure([]byte(`{"token":"123456:ABC-DEF"}`))
	require.NoError(t, err)

	tests := []struct {
		name       string
		msg        *provider.Message
		mockError  error
		wantErr    bool
		errContain string
		checkFunc  func(t *testing.T, sent []tgbotapi.Chattable)
	}{
		{
			name: "simple send",
			msg: &provider.Message{
				To:      []string{"123456789"},
				Subject: "Test Subject",
				Body:    "Test Body",
				Params:  []byte(`{}`),
			},
			wantErr: false,
			checkFunc: func(t *testing.T, sent []tgbotapi.Chattable) {
				assert.Len(t, sent, 1)
				msg, ok := sent[0].(tgbotapi.MessageConfig)
				assert.True(t, ok)
				assert.Equal(t, int64(123456789), msg.ChatID)
				assert.Contains(t, msg.Text, "<b>Test Subject</b>")
				assert.Contains(t, msg.Text, "Test Body")
				assert.Equal(t, "HTML", msg.ParseMode)
			},
		},
		{
			name: "with attachment",
			msg: &provider.Message{
				To:      []string{"123456789"},
				Subject: "Test",
				Body:    "Body",
				Params: func() []byte {
					att := Attachment{
						Filename:    "file.txt",
						Content:     base64.StdEncoding.EncodeToString([]byte("Hello")),
						ContentType: "text/plain",
					}
					opts := SendOptions{Attachments: []Attachment{att}}
					b, _ := json.Marshal(opts)
					return b
				}(),
			},
			wantErr: false,
			checkFunc: func(t *testing.T, sent []tgbotapi.Chattable) {
				assert.Len(t, sent, 2)
				textMsg, ok := sent[0].(tgbotapi.MessageConfig)
				assert.True(t, ok)
				assert.Contains(t, textMsg.Text, "Body")
				docMsg, ok := sent[1].(tgbotapi.DocumentConfig)
				assert.True(t, ok)
				fileBytes, ok := docMsg.File.(tgbotapi.FileBytes)
				assert.True(t, ok)
				assert.Equal(t, "file.txt", fileBytes.Name)
			},
		},
		{
			name: "with parse_mode",
			msg: &provider.Message{
				To:      []string{"123456789"},
				Subject: "Test",
				Body:    "Body",
				Params:  []byte(`{"parse_mode":"Markdown"}`),
			},
			wantErr: false,
			checkFunc: func(t *testing.T, sent []tgbotapi.Chattable) {
				msg, ok := sent[0].(tgbotapi.MessageConfig)
				assert.True(t, ok)
				assert.Equal(t, "Markdown", msg.ParseMode)
			},
		},
		{
			name: "invalid chat_id",
			msg: &provider.Message{
				To:      []string{"invalid"},
				Subject: "Test",
				Body:    "Body",
				Params:  []byte(`{}`),
			},
			wantErr:    true,
			errContain: "invalid chat_id",
		},
		{
			name: "send error",
			msg: &provider.Message{
				To:      []string{"123456789"},
				Subject: "Test",
				Body:    "Body",
				Params:  []byte(`{}`),
			},
			mockError:  errors.New("telegram api error"),
			wantErr:    true,
			errContain: "failed to send text message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sentMessages = []tgbotapi.Chattable{}
			lastError = tt.mockError

			resp := p.Send(tt.msg)
			if tt.wantErr {
				assert.Error(t, resp.Err)
				if tt.errContain != "" {
					assert.Contains(t, resp.Err.Error(), tt.errContain)
				}
			} else {
				assert.NoError(t, resp.Err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, sentMessages)
				}
			}
		})
	}
}
