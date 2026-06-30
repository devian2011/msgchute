package main

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/go-plugin"

	"github.com/devian2011/msgchute/pkg/shared/provider"
)

type SmtpConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type SmtpProvider struct{}

func (s *SmtpProvider) GetCode() string {
	return "smtp"
}

func (s *SmtpProvider) Send(params []byte, msg *provider.Message) error {
	var config SmtpConfig
	if err := json.Unmarshal(params, &config); err != nil {
		return fmt.Errorf("failed to parse smtp config: %w", err)
	}

	var meta map[string]any
	if len(msg.Meta) > 0 {
		_ = json.Unmarshal(msg.Meta, &meta)
	}

	fmt.Printf("[SMTP Plugin] Connecting to %s:%d using user %s\n", config.Host, config.Port, config.Username)
	fmt.Printf("[SMTP Plugin] Sending Email to %v, Subject: %s, Body: %s, Meta: %v\n", msg.To, msg.Subject, msg.Body, meta)
	return nil
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
