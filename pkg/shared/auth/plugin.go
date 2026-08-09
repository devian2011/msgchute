//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative auth.proto
package auth

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

type Provider interface {
	Allow(ctx context.Context, request *http.Request) (bool, error)
	Configure(payload []byte) error
}

var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "AUTH_PROVIDER_PLUGIN",
	MagicCookieValue: "secure_auth_handshake",
}

var PluginMap = map[string]plugin.Plugin{
	"auth_provider": &ProviderPlugin{},
}

type ProviderPlugin struct {
	plugin.Plugin
	Provider Provider
}

func (p *ProviderPlugin) GRPCServer(_ *plugin.GRPCBroker, s *grpc.Server) error {
	RegisterAuthProviderServiceServer(s, &GRPCServer{Impl: p.Provider})
	return nil
}

func (p *ProviderPlugin) GRPCClient(_ context.Context, _ *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: NewAuthProviderServiceClient(c)}, nil
}

// GRPCServer implements the gRPC server side of the plugin.
type GRPCServer struct {
	UnimplementedAuthProviderServiceServer
	Impl Provider
}

func (s *GRPCServer) Allow(ctx context.Context, req *AllowRequest) (*AllowResponse, error) {
	httpReq, err := toHTTPRequest(req)
	if err != nil {
		return nil, err // internal conversion error
	}

	allowed, err := s.Impl.Allow(ctx, httpReq)
	if err != nil {
		// Provider error – return via response field
		return &AllowResponse{Allowed: false, Error: err.Error()}, nil
	}
	return &AllowResponse{Allowed: allowed}, nil
}

func (s *GRPCServer) Configure(ctx context.Context, req *ConfigureRequest) (*ConfigureResponse, error) {
	err := s.Impl.Configure(req.Payload)
	if err != nil {
		return &ConfigureResponse{Error: err.Error()}, nil
	}
	return &ConfigureResponse{}, nil
}

// GRPCClient implements the gRPC client side of the plugin.
type GRPCClient struct {
	client AuthProviderServiceClient
}

func (c *GRPCClient) Allow(ctx context.Context, req *http.Request) (bool, error) {
	pbReq := toProtoRequest(req)
	resp, err := c.client.Allow(ctx, pbReq)
	if err != nil {
		return false, err
	}
	if resp.Error != "" {
		return false, errors.New(resp.Error)
	}
	return resp.Allowed, nil
}

func (c *GRPCClient) Configure(payload []byte) error {
	resp, err := c.client.Configure(context.Background(), &ConfigureRequest{Payload: payload})
	if err != nil {
		return err
	}
	if resp.Error != "" {
		return errors.New(resp.Error)
	}
	return nil
}

// toProtoRequest converts an http.Request to AllowRequest.
func toProtoRequest(req *http.Request) *AllowRequest {
	pbReq := &AllowRequest{
		Method:  req.Method,
		Path:    req.URL.Path,
		Headers: make(map[string]*HeaderValues),
		Cookies: make(map[string]string),
	}

	for key, values := range req.Header {
		pbReq.Headers[key] = &HeaderValues{Values: values}
	}

	for _, cookie := range req.Cookies() {
		pbReq.Cookies[cookie.Name] = cookie.Value
	}

	return pbReq
}

// toHTTPRequest converts AllowRequest to http.Request.
func toHTTPRequest(req *AllowRequest) (*http.Request, error) {
	u, err := url.Parse(req.Path)
	if err != nil {
		return nil, err
	}

	httpReq := &http.Request{
		Method: req.Method,
		URL:    u,
		Header: make(http.Header),
	}

	for key, headerVals := range req.Headers {
		httpReq.Header[key] = headerVals.Values
	}

	for name, value := range req.Cookies {
		httpReq.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	return httpReq, nil
}
