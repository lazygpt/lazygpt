//

package openai

import (
	"context"
	"errors"
	"fmt"

	"github.com/sashabaranov/go-openai"

	"github.com/lazygpt/lazygpt/plugin/api"
	"github.com/lazygpt/lazygpt/plugin/log"
)

var (
	// ErrNoCompletions is returned when no completions are returned from the OpenAI API.
	ErrNoCompletions = errors.New("no completions returned")

	// ErrNoEmbeddings is returned when no embeddings are returned from the OpenAI API.
	ErrNoEmbeddings = errors.New("no embeddings returned")
)

type Plugin struct {
	Client *openai.Client
}

var (
	_ api.Completion = (*Plugin)(nil)
	_ api.Embedding  = (*Plugin)(nil)
	_ api.Interfaces = (*Plugin)(nil)
)

// NewPlugin creates a new Plugin instance.
func NewPlugin(key string) *Plugin {
	return &Plugin{
		Client: openai.NewClient(key),
	}
}

// Complete implements the `Completion` interface.
func (plugin *Plugin) Complete(
	ctx context.Context,
	messages []api.Message,
) (*api.Message, api.Reason, error) {
	msgs := make([]openai.ChatCompletionMessage, len(messages))
	for i := range messages {
		msgs[i] = openai.ChatCompletionMessage{
			Role:    messages[i].Role,
			Content: messages[i].Content,
		}
	}

	req := openai.ChatCompletionRequest{
		Model:    openai.GPT3Dot5Turbo,
		Messages: msgs,
		N:        1,
	}

	resp, err := plugin.Client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, api.Reason_UNKNOWN, fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, api.Reason_UNKNOWN, ErrNoCompletions
	}

	response := &api.Message{
		Role:    resp.Choices[0].Message.Role,
		Content: resp.Choices[0].Message.Content,
	}

	log.Info(
		ctx, "OpenAI response",
		"content", response.Content,
		"role", response.Role,
		"reason", resp.Choices[0].FinishReason,
	)

	return response, api.StringToReason(resp.Choices[0].FinishReason), nil
}

// Embedding implements the `api.Embedding` interface.
func (plugin *Plugin) Embedding(ctx context.Context, input string) ([]float32, error) {
	req := openai.EmbeddingRequest{
		Model: openai.AdaEmbeddingV2,
		Input: []string{input},
	}

	resp, err := plugin.Client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, ErrNoEmbeddings
	}

	return resp.Data[0].Embedding, nil
}

// Interfaces implements the `api.Interfaces` interface.
func (plugin *Plugin) Interfaces(_ context.Context) ([]string, error) {
	return []string{
		"completion",
		"embedding",
		"interfaces",
	}, nil
}
