//

package openai

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"

	"github.com/lazygpt/lazygpt/plugin/api"
)

type OpenAIPlugin struct {
	Client *openai.Client
}

var (
	_ api.Completion = (*OpenAIPlugin)(nil)
)

// NewOpenAIPlugin creates a new OpenAIPlugin instance
func NewOpenAIPlugin(key string) *OpenAIPlugin {
	return &OpenAIPlugin{
		Client: openai.NewClient(key),
	}
}

// Complete implements the CompletionServer interface.
func (plugin *OpenAIPlugin) Complete(
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
		Model: openai.GPT3Dot5Turbo,
		Messages: msgs,
		N: 1,
	}

	resp, err := plugin.Client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, api.Reason_UNKNOWN, fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, api.Reason_UNKNOWN, fmt.Errorf("no completions returned")
	}

	response := &api.Message{
		Role:    resp.Choices[0].Message.Role,
		Content: resp.Choices[0].Message.Content,
	}

	return response, api.StringToReason(resp.Choices[0].FinishReason), nil
}
