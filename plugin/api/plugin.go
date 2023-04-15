//

package api

import (
	"context"
	"fmt"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

type Plugin struct {
	plugin.Plugin

	// client is the function to return a gRPC client.
	client func(*grpc.ClientConn) (interface{}, error)

	// register is the function to register a gRPC server.
	register func(*grpc.Server)
}

var (
	_ plugin.Plugin     = (*Plugin)(nil)
	_ plugin.GRPCPlugin = (*Plugin)(nil)
)

// NewPlugin return a new plugin instance for the given plugin implementation.
func NewPlugin(
	register func(*grpc.Server),
	client func(*grpc.ClientConn) (interface{}, error),
) *Plugin {
	return &Plugin{
		client:   client,
		register: register,
	}
}

// Server always returns an error, we only support GRPC.
func (plugin *Plugin) Server(_ *plugin.MuxBroker) (interface{}, error) {
	return nil, ErrNotGRPC
}

// Client always returns an error, we only support GRPC.
func (plugin *Plugin) Client(_ *plugin.MuxBroker, _ *rpc.Client) (interface{}, error) {
	return nil, ErrNotGRPC
}

// GRPCServer registers the plugin with the gRPC server.
func (plugin *Plugin) GRPCServer(_ *plugin.GRPCBroker, srv *grpc.Server) error {
	plugin.register(srv)

	return nil
}

// GRPCClient returns the plugin client.
func (plugin *Plugin) GRPCClient(
	_ context.Context,
	_ *plugin.GRPCBroker,
	client *grpc.ClientConn,
) (interface{}, error) {
	raw, err := plugin.client(client)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return raw, nil
}
