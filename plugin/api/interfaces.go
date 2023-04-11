//

package api

import (
	"context"
	"fmt"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// Interfaces is the interface that plugins must implement to provide
// discovery of implemented interfaces.
type Interfaces interface {
	// Interfaces returns a list of interfaces that the plugin implements.
	Interfaces(ctx context.Context) ([]string, error)
}

// InterfacesPlugin is the implementation of the plugin for the interfaces
// plugin.
type InterfacesPlugin struct {
	plugin.Plugin

	Interfaces Interfaces
}

var (
	_ plugin.Plugin     = (*InterfacesPlugin)(nil)
	_ plugin.GRPCPlugin = (*InterfacesPlugin)(nil)
)

// NewInterfacesPlugin returns a new InterfacesPlugin.
func NewInterfacesPlugin(interfaces Interfaces) *InterfacesPlugin {
	return &InterfacesPlugin{
		Interfaces: interfaces,
	}
}

// Server always returns an error, we only support GRPC.
func (plugin *InterfacesPlugin) Server(_ *plugin.MuxBroker) (interface{}, error) {
	return nil, ErrNotGRPC
}

// Client always returns an error, we only support GRPC.
func (plugin *InterfacesPlugin) Client(_ *plugin.MuxBroker, _ *rpc.Client) (interface{}, error) {
	return nil, ErrNotGRPC
}

// GRPCServer registers the interfaces plugin with the gRPC server.
func (plugin *InterfacesPlugin) GRPCServer(_ *plugin.GRPCBroker, srv *grpc.Server) error {
	RegisterInterfacesServer(srv, NewInterfacesGRPCServer(plugin.Interfaces))

	return nil
}

// GRPCClient returns the interfaes plugin client.
func (plugin *InterfacesPlugin) GRPCClient(
	_ context.Context,
	_ *plugin.GRPCBroker,
	client *grpc.ClientConn,
) (interface{}, error) {
	return NewInterfacesGRPCClient(NewInterfacesClient(client)), nil
}

// InterfacesGRPCServer is the gRPC server implementation of the plugin.
type InterfacesGRPCServer struct {
	UnimplementedInterfacesServer

	Impl Interfaces
}

var _ InterfacesServer = (*InterfacesGRPCServer)(nil)

// NewInterfacesGRPCServer returns a new InterfacesGRPCServer.
func NewInterfacesGRPCServer(impl Interfaces) *InterfacesGRPCServer {
	return &InterfacesGRPCServer{
		Impl: impl,
	}
}

// Interfaces implements the gRPC server for the interfaces plugin.
func (s *InterfacesGRPCServer) Interfaces(
	ctx context.Context,
	_ *InterfacesRequest,
) (*InterfacesResponse, error) {
	interfaces, err := s.Impl.Interfaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("interfaces failed: %w", err)
	}

	return &InterfacesResponse{
		Interfaces: interfaces,
	}, nil
}

// InterfacesGRPCClient is the gRPC client implementation of the plugin.
type InterfacesGRPCClient struct {
	Client InterfacesClient
}

var _ Interfaces = (*InterfacesGRPCClient)(nil)

// NewInterfacesGRPCClient returns a new InterfacesGRPCClient.
func NewInterfacesGRPCClient(client InterfacesClient) *InterfacesGRPCClient {
	return &InterfacesGRPCClient{
		Client: client,
	}
}

// Interfaces implements the gRPC client for the interfaces plugin.
func (c *InterfacesGRPCClient) Interfaces(ctx context.Context) ([]string, error) {
	req := &InterfacesRequest{}

	resp, err := c.Client.Interfaces(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("interfaces failed: %w", err)
	}

	return resp.Interfaces, nil
}
