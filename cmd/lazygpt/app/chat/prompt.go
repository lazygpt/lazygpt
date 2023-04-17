package chat

import (
	"context"
	"errors"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/lazygpt/lazygpt/plugin/api"
	"github.com/lazygpt/lazygpt/plugin/log"
)

const (
	RoleAssistant = "assistant"
	RoleUser      = "user"
)

var (
	ErrPromptNotString = errors.New("prompt must be a string")
	ErrPromptEmpty     = errors.New("prompt can not be empty")
)

type contextUpdateFn func()

type promptContext struct {
	completion api.Completion
	messages   *[]api.Message
}

type promptUpdated struct{}

type promptExecutor func() tea.Cmd

func doPromptUpdated() tea.Cmd {
	return func() tea.Msg {
		return promptUpdated{}
	}
}

func NewPromptContext(completion api.Completion, messages *[]api.Message) *promptContext {
	if messages == nil {
		messagesSlice := make([]api.Message, 0)
		messages = &messagesSlice
	}

	return &promptContext{
		completion: completion,
		messages:   messages,
	}
}

func (pc *promptContext) AddUserMessage(promptText string) api.Message {
	message := api.Message{
		Role:    RoleUser,
		Content: promptText,
	}
	*pc.messages = append(*pc.messages, message)

	return message
}

// Executor returns a function that accepts a promptText and returns a function that will send the promptText to the completion plugin.
func (pc *promptContext) Executor(ctx context.Context) promptExecutor {
	return func() tea.Cmd {
		return func() tea.Msg {
			response, reason, err := pc.completion.Complete(ctx, *pc.messages)
			if err != nil || response == nil {
				log.Error(
					ctx, "failed to complete", err,
					"response", response,
					"reason", reason,
				)

				return ""
			}

			*pc.messages = append(*pc.messages, api.Message{
				Role:    response.Role,
				Content: response.Content,
			})

			return promptUpdated{}
		}
	}
}
