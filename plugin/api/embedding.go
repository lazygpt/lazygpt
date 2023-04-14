//

package api

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
)

// Embedding is the interface that plugins must implement to provide
// embedding functionality.
type Embedding interface {
	// Embedding returns the embedding for the given input.
	Embedding(ctx context.Context, input string) ([]float32, error)
}

// NewEmbeddingPlugin returns a new EmbeddingPlugin.
func NewEmbeddingPlugin(embedding Embedding) *Plugin {
	return NewPlugin(
		func(srv *grpc.Server) {
			RegisterEmbeddingServer(srv, NewEmbeddingGRPCServer(embedding))
		},
		func(client *grpc.ClientConn) (interface{}, error) {
			return NewEmbeddingGRPCClient(NewEmbeddingClient(client)), nil
		},
	)
}

// EmbeddingGRPCServer is the gRPC server implementation of the plugin.
type EmbeddingGRPCServer struct {
	UnimplementedEmbeddingServer

	Impl Embedding
}

var _ EmbeddingServer = (*EmbeddingGRPCServer)(nil)

// NewEmbeddingGRPCServer returns a new EmbeddingGRPCServer.
func NewEmbeddingGRPCServer(impl Embedding) *EmbeddingGRPCServer {
	return &EmbeddingGRPCServer{
		Impl: impl,
	}
}

// Embedding implements the gRPC server for the embedding plugin.
func (s *EmbeddingGRPCServer) Embedding(
	ctx context.Context,
	req *EmbeddingRequest,
) (*EmbeddingResponse, error) {
	ctx = InitLogging(ctx, "embedding")

	embedding, err := s.Impl.Embedding(ctx, req.Input)
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}

	return &EmbeddingResponse{
		Embedding: embedding,
	}, nil
}

// EmbeddingGRPCClient is the gRPC client implementation of the plugin.
type EmbeddingGRPCClient struct {
	Client EmbeddingClient
}

var _ Embedding = (*EmbeddingGRPCClient)(nil)

// NewEmbeddingGRPCClient returns a new EmbeddingGRPCClient.
func NewEmbeddingGRPCClient(client EmbeddingClient) *EmbeddingGRPCClient {
	return &EmbeddingGRPCClient{
		Client: client,
	}
}

// Embedding implements the gRPC client for the embedding plugin.
func (c *EmbeddingGRPCClient) Embedding(
	ctx context.Context,
	input string,
) ([]float32, error) {
	req := &EmbeddingRequest{
		Input: input,
	}

	resp, err := c.Client.Embedding(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}

	return resp.Embedding, nil
}
