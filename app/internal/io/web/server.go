package web

import (
	"context"
	"net/http"
	"time"
)

type Config struct {
	Addr   string `yaml:"addr" env:"APP_HTTP_ADDR"`
	Config struct {
		ReadTimeout       time.Duration `env:"APP_HTTP_READ_TIMEOUT" yaml:"readTimeout"`
		ReadHeaderTimeout time.Duration `env:"APP_HTTP_READ_HEAD_TIMEOUT" yaml:"readHeaderTimeout"`
		WriteTimeout      time.Duration `env:"APP_HTTP_WRITE_TIMEOUT" yaml:"writeTimeout"`
	} `yaml:"config"`
	TLS struct {
		CertFile string `env:"APP_HTTP_CERT_FILE" yaml:"certFile"`
		KeyFile  string `env:"APP_HTTP_KEY_FILE" yaml:"keyFile"`
	} `yaml:"tls"`
	WithSwagger bool   `env:"APP_HTTP_WITH_SWAGGER" yaml:"withSwagger"`
	Dashboard   string `env:"APP_HTTP_WITH_DASHBOARD" yaml:"dashboard"`
}

type Server struct {
	cfg *Config
	srv *http.Server
}

func NewServer(cfg *Config) *Server {
	return &Server{
		cfg: cfg,
		srv: &http.Server{
			Addr:              cfg.Addr,
			ReadTimeout:       cfg.Config.ReadTimeout,
			ReadHeaderTimeout: cfg.Config.ReadHeaderTimeout,
			WriteTimeout:      cfg.Config.WriteTimeout,
		},
	}
}

func (s *Server) WithSwagger() bool {
	return s.cfg.WithSwagger
}

func (s *Server) Dashboard() string {
	return s.cfg.Dashboard
}

func (s *Server) SetHandler(h http.Handler) {
	s.srv.Handler = h
}

func (s *Server) Run(err chan error) {
	if s.cfg.TLS.KeyFile != "" && s.cfg.TLS.CertFile != "" {
		err <- s.srv.ListenAndServeTLS(s.cfg.TLS.CertFile, s.cfg.TLS.KeyFile)
	} else {
		err <- s.srv.ListenAndServe()
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
