//

package api

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
)

// Interfaces is the interface that plugins must implement to provide
// discovery of implemented interfaces.
type Interfaces interface {
	// Interfaces returns a list of interfaces that the plugin implements.
	Interfaces(ctx context.Context) ([]string, error)
}

// NewInterfacesPlugin returns a new InterfacesPlugin.
func NewInterfacesPlugin(interfaces Interfaces) *Plugin {
	return NewPlugin(
		func(srv *grpc.Server) {
			RegisterInterfacesServer(srv, NewInterfacesGRPCServer(interfaces))
		},
		func(client *grpc.ClientConn) (interface{}, error) {
			return NewInterfacesGRPCClient(NewInterfacesClient(client)), nil
		},
	)
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
	ctx = InitLogging(ctx, "interfaces")

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
