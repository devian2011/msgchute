package main

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/hashicorp/go-plugin"

	"github.com/devian2011/msgchute/pkg/shared/auth"
)

type Config struct {
	HeaderToken string `json:"headerToken"`
	AdminToken  string `json:"adminToken"`
	PublicToken string `json:"publicToken"`
}

type Auth struct {
	cfg Config
}

func (s *Auth) Configure(payload []byte) error {
	var cfg Config
	if err := sonic.Unmarshal(payload, &cfg); err != nil {
		return err
	}
	if cfg.HeaderToken == "" || cfg.AdminToken == "" {
		return errors.New("auth: invalid config")
	}
	s.cfg = cfg
	return nil
}

func (s *Auth) Allow(_ context.Context, req *http.Request) (bool, error) {
	token := req.Header.Get(s.cfg.HeaderToken)
	path := req.URL.Path

	if s.cfg.AdminToken == token {
		return true, nil
	}
	if strings.HasPrefix(path, "/api/admin") {
		return false, nil
	}
	return s.cfg.PublicToken == token, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: auth.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"auth_provider": &auth.ProviderPlugin{Provider: &Auth{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
