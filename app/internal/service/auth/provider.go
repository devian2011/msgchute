package auth

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/bytedance/sonic"
	"github.com/hashicorp/go-plugin"

	"github.com/devian2011/msgchute/pkg/file"
	"github.com/devian2011/msgchute/pkg/shared/auth"
)

type Provider struct {
	ctx      context.Context
	provider auth.Provider
	client   *plugin.Client
}

func NewProvider(ctx context.Context, cfg *Config) (*Provider, error) {
	c, p, initErr := initProvider(cfg)
	if initErr != nil {
		return nil, initErr
	}

	return &Provider{
		ctx:      ctx,
		provider: p,
		client:   c,
	}, nil
}

func initProvider(cfg *Config) (*plugin.Client, auth.Provider, error) {
	if !file.Exists(cfg.Plugin) {
		return nil, nil, fmt.Errorf("auth plugin binary not found at %s", cfg.Plugin)
	}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  auth.HandshakeConfig,
		Plugins:          auth.PluginMap,
		Cmd:              exec.Command(cfg.Plugin),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, nil, fmt.Errorf("handshake failed for auth provider: %w", err)
	}

	raw, err := rpcClient.Dispense("auth_provider") // исправлен ключ
	if err != nil {
		client.Kill()
		return nil, nil, fmt.Errorf("plugin does not implement auth_provider: %w", err)
	}

	p, ok := raw.(auth.Provider)
	if !ok {
		client.Kill()
		return nil, nil, errors.New("type assertion failed for auth plugin")
	}

	paramsJSON, err := sonic.ConfigDefault.Marshal(cfg.Params)
	if err != nil {
		client.Kill()
		return nil, nil, fmt.Errorf("failed to marshal auth config: %w", err)
	}

	if err := p.Configure(paramsJSON); err != nil {
		client.Kill()
		return nil, nil, fmt.Errorf("failed to configure auth plugin: %w", err)
	}

	return client, p, nil
}

func (m *Provider) Close() {
	m.client.Kill()
}

func (m *Provider) GetProvider() auth.Provider {
	return m.provider
}
