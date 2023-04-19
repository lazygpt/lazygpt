//

package api

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
)

// Memory is the interface that plugins must implement to provide
// memory functionality.
type Memory interface {
	// Memorize memorizes each data string.
	Memorize(ctx context.Context, data []string) error

	// Recall recalls the memory closest to the data string. If a count is
	// specified, it returns the count closest memories.
	Recall(ctx context.Context, data string, count ...int) ([]string, error)
}

// NewMemoryPlugin returns a new MemoryPlugin.
func NewMemoryPlugin(memory Memory) *Plugin {
	return NewPlugin(
		func(srv *grpc.Server) {
			RegisterMemoryServer(srv, NewMemoryGRPCServer(memory))
		},
		func(client *grpc.ClientConn) (interface{}, error) {
			return NewMemoryGRPCClient(NewMemoryClient(client)), nil
		},
	)
}

// MemoryGRPCServer is the gRPC server implementation of the plugin.
type MemoryGRPCServer struct {
	UnimplementedMemoryServer

	Impl Memory
}

var _ MemoryServer = (*MemoryGRPCServer)(nil)

// NewMemoryGRPCServer returns a new MemoryGRPCServer.
func NewMemoryGRPCServer(impl Memory) *MemoryGRPCServer {
	return &MemoryGRPCServer{
		Impl: impl,
	}
}

// Memorize implements the gRPC server for the memory plugin memorize method.
func (s *MemoryGRPCServer) Memorize(
	ctx context.Context,
	req *MemorizeRequest,
) (*MemorizeResponse, error) {
	ctx = InitLogging(ctx, "memorize")

	if err := s.Impl.Memorize(ctx, req.Data); err != nil {
		return nil, fmt.Errorf("memorize failed: %w", err)
	}

	return &MemorizeResponse{}, nil
}

// Recall implements the gRPC server for the memory plugin recall method.
func (s *MemoryGRPCServer) Recall(
	ctx context.Context,
	req *RecallRequest,
) (*RecallResponse, error) {
	ctx = InitLogging(ctx, "recall")

	memory, err := s.Impl.Recall(ctx, req.Data, int(req.Count))
	if err != nil {
		return nil, fmt.Errorf("recall failed: %w", err)
	}

	return &RecallResponse{
		Data: memory,
	}, nil
}

// MemoryGRPCClient is the gRPC client implementation of the plugin.
type MemoryGRPCClient struct {
	Client MemoryClient
}

var _ Memory = (*MemoryGRPCClient)(nil)

// NewMemoryGRPCClient returns a new MemoryGRPCClient.
func NewMemoryGRPCClient(client MemoryClient) *MemoryGRPCClient {
	return &MemoryGRPCClient{
		Client: client,
	}
}

// Memorize implements the gRPC client for the memory plugin.
func (c *MemoryGRPCClient) Memorize(
	ctx context.Context,
	data []string,
) error {
	req := &MemorizeRequest{
		Data: data,
	}

	_, err := c.Client.Memorize(ctx, req)
	if err != nil {
		return fmt.Errorf("memory failed: %w", err)
	}

	return nil
}

// Recall implements the gRPC client for the memory plugin.
func (c *MemoryGRPCClient) Recall(
	ctx context.Context,
	data string,
	count ...int,
) ([]string, error) {
	req := &RecallRequest{
		Data:  data,
		Count: 1,
	}

	if len(count) > 0 {
		req.Count = int32(count[0])
	}

	resp, err := c.Client.Recall(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("recall failed: %w", err)
	}

	return resp.Data, nil
}
