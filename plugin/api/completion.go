//

package api

import (
	"context"
	"fmt"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// Completion is the interface that plugins must implement to provide
// completion suggestions.
type Completion interface {
	// Complete returns a list of possible completions for the given input.
	Complete(ctx context.Context, messages []Message) (*Message, Reason, error)
}

// CompletionPlugin is the implementation of the plugin for the completion
// plugin.
type CompletionPlugin struct {
	plugin.Plugin

	Completion Completion
}

var (
	_ plugin.Plugin = (*CompletionPlugin)(nil)
	_ plugin.GRPCPlugin = (*CompletionPlugin)(nil)
)

// NewCompletionPlugin returns a new CompletionPlugin.
func NewCompletionPlugin(completion Completion) *CompletionPlugin {
	return &CompletionPlugin{
		Completion: completion,
	}
}

// Server always returns an error, we only support GRPC.
func (plugin *CompletionPlugin) Server(_ *plugin.MuxBroker) (interface{}, error) {
	return nil, ErrNotGRPC
}

// Client always returns an error, we only support GRPC.
func (plugin *CompletionPlugin) Client(_ *plugin.MuxBroker, _ *rpc.Client) (interface{}, error) {
	return nil, ErrNotGRPC
}

// GRPCServer registers the completion plugin with the gRPC server.
func (plugin *CompletionPlugin) GRPCServer(broker *plugin.GRPCBroker, srv *grpc.Server) error {
	RegisterCompletionServer(srv, NewCompletionGRPCServer(plugin.Completion))
	return nil
}

// GRPCClient returns the completion plugin client.
func (plugin *CompletionPlugin) GRPCClient(
	ctx context.Context,
	broker *plugin.GRPCBroker,
	client *grpc.ClientConn,
) (interface{}, error) {
	return NewCompletionGRPCClient(NewCompletionClient(client)), nil
}

// CompletionGRPCServer is the gRPC server implementation of the plugin.
type CompletionGRPCServer struct {
	UnimplementedCompletionServer

	Impl Completion
}

// NewCompletionGRPCServer returns a new CompletionGRPCServer.
func NewCompletionGRPCServer(impl Completion) *CompletionGRPCServer {
	return &CompletionGRPCServer{
		Impl: impl,
	}
}

// Complete implements the gRPC server for the completion plugin.
func (s *CompletionGRPCServer) Complete(
	ctx context.Context,
	req *CompletionRequest,
) (*CompletionResponse, error) {
	msgs := make([]Message, len(req.Messages))
	for i := range req.Messages {
		msgs[i] = Message{
			Role:    req.Messages[i].Role,
			Content: req.Messages[i].Content,
		}
	}

	msg, reason, err := s.Impl.Complete(ctx, msgs)
	if err != nil {
		return nil, fmt.Errorf("completion failed: %w", err)
	}

	return &CompletionResponse{
		Message: msg,
		Reason:  reason,
	}, nil
}

// CompletionGRPCClient is the gRPC client implementation of the plugin.
type CompletionGRPCClient struct {
	Client CompletionClient
}

var _ Completion = (*CompletionGRPCClient)(nil)

// NewCompletionGRPCClient returns a new CompletionGRPCClient.
func NewCompletionGRPCClient(client CompletionClient) *CompletionGRPCClient {
	return &CompletionGRPCClient{
		Client: client,
	}
}

// Complete implements the gRPC client for the completion plugin.
func (c *CompletionGRPCClient) Complete(
	ctx context.Context,
	messages []Message,
) (*Message, Reason, error) {
	req := &CompletionRequest{
		Messages: make([]*Message, len(messages)),
	}

	for i := range messages {
		req.Messages[i] = &messages[i]
	}

	resp, err := c.Client.Complete(ctx, req)
	if err != nil {
		return nil, Reason_UNKNOWN, fmt.Errorf("completion failed: %w", err)
	}

	return resp.Message, resp.Reason, nil
}
