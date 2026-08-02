//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative provider.proto

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

type Message struct {
	To      []string
	Params  []byte
	Subject string
	Body    string
}

type MessageResponse struct {
	IsCritical bool
	Err        error
	Response   string
}

type Provider interface {
	Configure(params []byte) error
	Send(msg *Message) *MessageResponse
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

func (s *GRPCServer) Configure(ctx context.Context, req *ConfigureRequest) (*Response, error) {
	err := s.Impl.Configure(req.Params)
	if err != nil {
		return &Response{Error: err.Error()}, nil
	}
	return &Response{Error: ""}, nil
}

func (s *GRPCServer) Send(ctx context.Context, req *MessageRequest) (*MsgResponse, error) {
	msg := &Message{
		To:      req.To,
		Subject: req.Subject,
		Body:    req.Body,
		Params:  req.Params,
	}

	resp := s.Impl.Send(msg)

	var errStr string
	if resp.Err != nil {
		errStr = resp.Err.Error()
	}

	return &MsgResponse{
		Response:   resp.Response,
		Error:      errStr,
		IsCritical: resp.IsCritical,
	}, nil
}

type GRPCClient struct {
	client ProviderServiceClient
}

func (c *GRPCClient) Configure(params []byte) error {
	resp, err := c.client.Configure(context.Background(), &ConfigureRequest{Params: params})
	if err != nil {
		return err
	}
	if resp.Error != "" {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func (c *GRPCClient) Send(msg *Message) *MessageResponse {
	gMsg := &MessageRequest{
		To:      msg.To,
		Subject: msg.Subject,
		Body:    msg.Body,
		Params:  msg.Params,
	}

	resp, err := c.client.Send(context.Background(), gMsg)
	if err != nil {
		return &MessageResponse{
			IsCritical: true,
			Err:        err,
		}
	}

	var respErr error
	if resp.Error != "" {
		respErr = fmt.Errorf(resp.Error)
	}

	return &MessageResponse{
		Response:   resp.Response,
		IsCritical: resp.IsCritical,
		Err:        respErr,
	}
}
