//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative auth.proto
package auth

import (
	"context"
	"net/http"
	"net/url"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

type Provider interface {
	Allow(ctx context.Context, request *http.Request) (bool, error)
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

type GRPCServer struct {
	UnimplementedAuthProviderServiceServer
	Impl Provider
}

func (s *GRPCServer) Allow(ctx context.Context, req *AllowRequest) (*AllowResponse, error) {
	httpReq, err := toHTTPRequest(req)
	if err != nil {
		return nil, err
	}

	allowed, err := s.Impl.Allow(ctx, httpReq)
	if err != nil {
		return nil, err
	}

	return &AllowResponse{Allowed: allowed}, nil
}

type GRPCClient struct {
	client AuthProviderServiceClient
}

func (c *GRPCClient) Allow(ctx context.Context, req *http.Request) (bool, error) {
	pbReq := toProtoRequest(req)

	resp, err := c.client.Allow(ctx, pbReq)
	if err != nil {
		return false, err
	}

	return resp.Allowed, nil
}

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
