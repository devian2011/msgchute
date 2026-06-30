//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative provider.proto

package provider

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

type Message struct {
	To      []string
	Meta    []byte
	Subject string
	Body    string
}

type Provider interface {
	GetCode() string
	Send(params []byte, msg *Message) error
}

var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "MSGCHUTE_PROVIDER_PLUGIN",
	MagicCookieValue: "secure_handshake",
}

var PluginMap = map[string]plugin.Plugin{
	"provider": &ProviderPlugin{},
}

type ProviderPlugin struct {
	plugin.Plugin
	Provider Provider
}

func (p *ProviderPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	RegisterProviderServiceServer(s, &GRPCServer{Impl: p.Provider})
	return nil
}

func (p *ProviderPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: NewProviderServiceClient(c)}, nil
}

type GRPCServer struct {
	UnimplementedProviderServiceServer
	Impl Provider
}

func (s *GRPCServer) GetCode(ctx context.Context, req *Empty) (*GetCodeResponse, error) {
	return &GetCodeResponse{Code: s.Impl.GetCode()}, nil
}

func (s *GRPCServer) Send(ctx context.Context, req *MessageRequest) (*Empty, error) {
	msg := &Message{
		To:      req.To,
		Subject: req.Subject,
		Body:    req.Body,
		Meta:    req.Meta,
	}

	err := s.Impl.Send(req.Params, msg)
	return &Empty{}, err
}

type GRPCClient struct {
	client ProviderServiceClient
}

func (c *GRPCClient) GetCode() string {
	resp, err := c.client.GetCode(context.Background(), &Empty{})
	if err != nil {
		return ""
	}
	return resp.Code
}

func (c *GRPCClient) Send(params []byte, msg *Message) error {
	gMsg := &MessageRequest{
		Params:  params,
		To:      msg.To,
		Subject: msg.Subject,
		Body:    msg.Body,
		Meta:    msg.Meta,
	}

	_, err := c.client.Send(context.Background(), gMsg)
	return err
}
