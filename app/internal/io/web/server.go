package web

import (
	"encoding/json"
	"net/http"

	"github.com/valyala/fasthttp"
)

type Config struct {
	Addr string `config:"addr"`
	Tls  struct {
		CertFile string `config:"cer"`
		KeyFile  string `config:"key"`
	} `config:"tls"`
}

type Handler func(command []byte, body []byte) (interface{}, error)

type Server struct {
	cfg *Config
	srv *fasthttp.Server
}

func NewServer(cfg *Config, fn Handler) *Server {
	return &Server{
		cfg: cfg,
		srv: &fasthttp.Server{
			Handler: func(ctx *fasthttp.RequestCtx) {
				data, err := fn(ctx.Path(), ctx.Request.Body())

				ctx.SetStatusCode(http.StatusOK)
				ctx.Response.Header.Add("Content-Type", "application/json")
				json.NewEncoder(ctx.Response.BodyWriter()).Encode(struct {
					Data  interface{} `json:"data"`
					Error error       `json:"err"`
				}{
					Data:  data,
					Error: err,
				})
			},
		},
	}
}

func (s *Server) Run(errCh chan error) {
	go func() {
		if s.cfg.Tls.CertFile != "" && s.cfg.Tls.KeyFile != "" {
			errCh <- s.srv.ListenAndServeTLS(s.cfg.Addr, s.cfg.Tls.CertFile, s.cfg.Tls.KeyFile)
		} else {
			errCh <- s.srv.ListenAndServe(s.cfg.Addr)
		}
	}()
}

func (s *Server) Shutdown() error {
	return s.srv.Shutdown()
}
