package auth

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-plugin"

	"github.com/devian2011/msgchute/pkg/file"
	"github.com/devian2011/msgchute/pkg/shared/auth"
	"github.com/devian2011/msgchute/pkg/shared/provider"
)

type Provider struct {
	ctx      context.Context
	provider auth.Provider
	client   *plugin.Client
}

func NewProvider(ctx context.Context, pluginPath string) (*Provider, error) {
	c, p, initErr := initProvider(pluginPath)
	if initErr != nil {
		return nil, initErr
	}

	return &Provider{
		ctx:      ctx,
		provider: p,
		client:   c,
	}, nil
}

func initProvider(pluginPath string) (*plugin.Client, auth.Provider, error) {
	if !file.Exists(pluginPath) {
		return nil, nil, fmt.Errorf("auth plugin binary not found at %s", pluginPath)
	}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  provider.HandshakeConfig,
		Plugins:          provider.PluginMap,
		Cmd:              exec.Command(pluginPath),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, nil, fmt.Errorf("handshake failed for auth provider: %s", err)
	}

	raw, err := rpcClient.Dispense("provider")
	if err != nil {
		client.Kill()
		return nil, nil, errors.New("plugin auth does not implement 'provider': auth.Provider")
	}

	p, ok := raw.(auth.Provider)
	if !ok {
		client.Kill()
		return nil, nil, errors.New("type assertion failed for auth plugin")
	}

	return client, p, nil
}

func (m *Provider) Close() {
	m.client.Kill()
}

func (m *Provider) GetProvider() auth.Provider {
	return m.provider
}
