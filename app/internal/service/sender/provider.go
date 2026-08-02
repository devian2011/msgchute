package sender

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"sync"

	"github.com/hashicorp/go-plugin"

	"github.com/devian2011/msgchute/pkg/file"
	"github.com/devian2011/msgchute/pkg/shared/provider"
)

type ProviderManager struct {
	ctx         context.Context
	providerMtx sync.RWMutex
	// pluginMap Key is a transport code, Value is a binary path
	pluginMap map[string]string
	providers map[string]provider.Provider
	clients   []*plugin.Client
}

func NewProviderManager(ctx context.Context, pluginMap map[string]string) *ProviderManager {
	return &ProviderManager{
		ctx:       ctx,
		pluginMap: pluginMap,
		providers: make(map[string]provider.Provider),
		clients:   make([]*plugin.Client, 0),
	}
}

func (m *ProviderManager) Close() {
	for _, client := range m.clients {
		client.Kill()
	}
	m.providerMtx.Lock()
	m.providers = make(map[string]provider.Provider)
	m.clients = make([]*plugin.Client, 0)
	m.providerMtx.Unlock()
}

func (m *ProviderManager) BuildPlugin(code string, params []byte) error {
	if _, exists := m.pluginMap[code]; exists {
		return fmt.Errorf("unknown plugin %s", code)
	}
	if !file.Exists(m.pluginMap[code]) {
		return fmt.Errorf("plugin %s binary not found at %s", code, m.pluginMap[code])
	}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  provider.HandshakeConfig,
		Plugins:          provider.PluginMap,
		Cmd:              exec.Command(m.pluginMap[code]),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return fmt.Errorf("handshake failed for '%s': %s", code, err)
	}

	raw, err := rpcClient.Dispense("provider")
	if err != nil {
		client.Kill()
		return fmt.Errorf("plugin '%s' does not implement 'provider': %s", code, err)
	}

	p, ok := raw.(provider.Provider)
	if !ok {
		client.Kill()
		return fmt.Errorf("type assertion failed for '%s'", code)
	}

	configureErr := p.Configure(params)
	if code == "" {
		client.Kill()
		return fmt.Errorf("plugin '%s' configure error: %s", code, configureErr)
	}

	m.providerMtx.Lock()
	m.providers[code] = p
	m.clients = append(m.clients, client)
	m.providerMtx.Unlock()

	slog.Info("successfully mounted channel: [%s] from '%s'", code, m.pluginMap[code])

	return nil
}

func (m *ProviderManager) GetProvider(code string) (provider.Provider, error) {
	m.providerMtx.RLock()
	defer m.providerMtx.RUnlock()

	p, exists := m.providers[code]
	if !exists {
		return nil, fmt.Errorf("provider channel '%s' is not active", code)
	}
	return p, nil
}

func (m *ProviderManager) GetProviders() []string {
	m.providerMtx.RLock()
	defer m.providerMtx.RUnlock()

	providers := make([]string, len(m.providers))
	for n := range m.providers {
		providers = append(providers, n)
	}

	return providers
}
