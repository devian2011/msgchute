package main

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/bytedance/sonic"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/hashicorp/go-plugin"

	"github.com/devian2011/msgchute/pkg/shared/provider"
)

// Config holds configuration for the Telegram bot.
type Config struct {
	Token string `json:"token"`
}

// Attachment represents a file attachment.
type Attachment struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"`
	ContentType string `json:"content_type,omitempty"`
}

// SendOptions are additional options for sending a message.
type SendOptions struct {
	ParseMode           string       `json:"parse_mode,omitempty"`
	DisableNotification bool         `json:"disable_notification,omitempty"`
	Attachments         []Attachment `json:"attachments,omitempty"`
}

// TgProvider implements the provider.Provider interface for Telegram.
type TgProvider struct {
	bot      *tgbotapi.BotAPI
	sendFunc func(tgbotapi.Chattable) (tgbotapi.Message, error) // for testing
}

// newBotAPI is a variable that can be overridden in tests.
var newBotAPI = func(token string) (*tgbotapi.BotAPI, error) {
	return tgbotapi.NewBotAPI(token)
}

// Configure parses the plugin configuration.
func (s *TgProvider) Configure(params []byte) error {
	var config Config
	if err := sonic.Unmarshal(params, &config); err != nil {
		return fmt.Errorf("failed to parse telegram config: %w", err)
	}
	if config.Token == "" {
		return fmt.Errorf("missing required field: token")
	}

	bot, err := newBotAPI(config.Token)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}
	s.bot = bot
	if s.sendFunc == nil {
		s.sendFunc = s.bot.Send
	}
	return nil
}

// Send sends a message via Telegram with optional attachments.
// Returns *provider.MessageResponse as required by the interface.
func (s *TgProvider) Send(msg *provider.Message) *provider.MessageResponse {
	if s.bot == nil {
		return &provider.MessageResponse{
			Err:        fmt.Errorf("bot not configured"),
			IsCritical: true,
		}
	}

	var opts SendOptions
	if len(msg.Params) > 0 {
		if err := sonic.Unmarshal(msg.Params, &opts); err != nil {
			return &provider.MessageResponse{
				Err:        fmt.Errorf("failed to parse send options: %w", err),
				IsCritical: true,
			}
		}
	}

	text := msg.Body
	if msg.Subject != "" {
		text = fmt.Sprintf("<b>%s</b>\n\n%s", msg.Subject, msg.Body)
	}

	parseMode := opts.ParseMode
	if parseMode == "" {
		parseMode = "HTML"
	}

	for _, chatIDStr := range msg.To {
		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			return &provider.MessageResponse{
				Err:        fmt.Errorf("invalid chat_id '%s': %w", chatIDStr, err),
				IsCritical: true,
			}
		}

		tgMsg := tgbotapi.NewMessage(chatID, text)
		tgMsg.ParseMode = parseMode
		tgMsg.DisableNotification = opts.DisableNotification
		if _, err := s.sendFunc(tgMsg); err != nil {
			return &provider.MessageResponse{
				Err:        fmt.Errorf("failed to send text message to %s: %w", chatIDStr, err),
				IsCritical: true,
			}
		}

		for _, att := range opts.Attachments {
			if att.Filename == "" || att.Content == "" {
				continue
			}
			fileBytes, err := base64.StdEncoding.DecodeString(att.Content)
			if err != nil {
				return &provider.MessageResponse{
					Err:        fmt.Errorf("failed to decode attachment '%s': %w", att.Filename, err),
					IsCritical: true,
				}
			}
			file := tgbotapi.FileBytes{
				Name:  att.Filename,
				Bytes: fileBytes,
			}
			doc := tgbotapi.NewDocument(chatID, file)
			doc.DisableNotification = opts.DisableNotification
			if _, err := s.sendFunc(doc); err != nil {
				return &provider.MessageResponse{
					Err:        fmt.Errorf("failed to send attachment '%s' to %s: %w", att.Filename, chatIDStr, err),
					IsCritical: true,
				}
			}
		}
	}
	// Success
	return &provider.MessageResponse{
		Err:      nil,
		Response: "Message sent successfully",
	}
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: provider.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"provider": &provider.ProviderPlugin{Provider: &TgProvider{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
