//

package api

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
)

type Message struct {
	Content string `json:"content"`
	Name    string `json:"name"`
	Role    string `json:"role"`
}

// Completion is the interface that plugins must implement to provide
// completion suggestions.
type Completion interface {
	// Complete returns a list of possible completions for the given input.
	Complete(ctx context.Context, messages []Message) (*Message, Reason, error)
}

// NewCompletionPlugin returns a new CompletionPlugin.
func NewCompletionPlugin(completion Completion) *Plugin {
	return NewPlugin(
		func(srv *grpc.Server) {
			RegisterCompletionServer(srv, NewCompletionGRPCServer(completion))
		},
		func(conn *grpc.ClientConn) (interface{}, error) {
			return NewCompletionGRPCClient(NewCompletionClient(conn)), nil
		},
	)
}

// CompletionGRPCServer is the gRPC server implementation of the plugin.
type CompletionGRPCServer struct {
	UnimplementedCompletionServer

	Impl Completion
}

var _ CompletionServer = (*CompletionGRPCServer)(nil)

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

	ctx = InitLogging(ctx, "completion")

	msg, reason, err := s.Impl.Complete(ctx, msgs)
	if err != nil {
		return nil, fmt.Errorf("completion failed: %w", err)
	}

	message := &CompletionMessage{
		Role:    msg.Role,
		Content: msg.Content,
	}

	if msg.Name != "" {
		message.Name = msg.Name
	}

	return &CompletionResponse{
		Message: message,
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
		Messages: make([]*CompletionMessage, len(messages)),
	}

	for idx := range messages {
		message := &CompletionMessage{
			Role:    messages[idx].Role,
			Content: messages[idx].Content,
		}

		if messages[idx].Name != "" {
			message.Name = messages[idx].Name
		}

		req.Messages[idx] = message
	}

	resp, err := c.Client.Complete(ctx, req)
	if err != nil {
		return nil, Reason_UNKNOWN, fmt.Errorf("completion failed: %w", err)
	}

	msg := &Message{
		Role:    resp.Message.Role,
		Content: resp.Message.Content,
	}

	if resp.Message.Name != "" {
		msg.Name = resp.Message.Name
	}

	return msg, resp.Reason, nil
}
