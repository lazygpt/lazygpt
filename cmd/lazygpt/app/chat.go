//

package app

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"

	"github.com/lazygpt/lazygpt/pkg/plugin"
	"github.com/lazygpt/lazygpt/pkg/tokens"
	"github.com/lazygpt/lazygpt/plugin/api"
	"github.com/lazygpt/lazygpt/plugin/log"
)

const (
	MaxMemory   = 10
	MemoryCount = 10
	MaxTokens   = 3000
)

func InitChatCmd(app *LazyGPTApp) {
	chatCmd := &cobra.Command{
		Use:   "chat",
		Short: "Start an interactive chat session with LazyGPT",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := plugin.NewManager()
			defer manager.Close()

			ctx := cmd.Context()

			completion, closeCompletion, err := Completion(ctx, manager, "openai")
			if err != nil {
				return fmt.Errorf("failed to get completion: %w", err)
			}
			defer func() {
				if err := closeCompletion(); err != nil {
					log.Error(ctx, "failed to close completion", err)
				}
			}()

			memory, closeMemory, err := Memory(ctx, manager, "local")
			if err != nil {
				return fmt.Errorf("failed to get memory: %w", err)
			}
			defer func() {
				if err := closeMemory(); err != nil {
					log.Error(ctx, "failed to close memory", err)
				}
			}()

			prompt.New(
				Executor(ctx, completion, memory),
				func(_ prompt.Document) []prompt.Suggest { return []prompt.Suggest{} },
				prompt.OptionPrefix("> "),
			).Run()

			return nil
		},
	}

	app.RootCmd.AddCommand(chatCmd)
}

func Executor(ctx context.Context, completion api.Completion, memory api.Memory) func(string) {
	var responses []api.Message

	return func(in string) {
		input := strings.TrimSpace(in)

		if input == "" {
			return
		}

		if input == "exit" {
			os.Exit(0)
		}

		memories := make([]string, 0, MaxMemory)

		for i := len(responses) - 1; i >= 0 && len(memories) < MaxMemory; i-- {
			memories = append(memories, Memorize(&responses[i], "", ""))
		}

		recollection, err := Recollection(ctx, memory, memories)
		if err != nil {
			log.Error(ctx, "failed to recollect", err)

			return
		}

		context, err := AIContext(recollection, responses, "gpt-3.5-turbo", MaxTokens)
		if err != nil {
			log.Error(ctx, "failed to create context", err)

			return
		}

		context = append(context, api.Message{
			Role:    "user",
			Content: input,
		})

		log.Debug(ctx, "context", context)

		response, reason, err := completion.Complete(ctx, context)
		if err != nil || response == nil {
			log.Error(
				ctx, "failed to complete", err,
				"response", response,
				"reason", reason,
			)

			return
		}

		fmt.Println(response.Content) //nolint:forbidigo // this is a CLI app

		responses = append(responses, api.Message{
			Role:    response.Role,
			Content: response.Content,
		})

		if err := memory.Memorize(
			ctx,
			[]string{Memorize(response, "", input)},
		); err != nil {
			log.Error(ctx, "failed to memorize", err)

			return
		}
	}
}

func Recollection(ctx context.Context, memory api.Memory, memories []string) ([]string, error) {
	var recollection []string

	for idx := range memories {
		recall, err := memory.Recall(ctx, memories[idx], MemoryCount)
		if err != nil {
			return nil, fmt.Errorf("failed to recall: %w", err)
		}

		recollection = append(recollection, recall...)
	}

	return recollection, nil
}

func Memorize(msg *api.Message, result string, feedback string) string {
	return fmt.Sprintf(
		"Assistant Reply: %s\nResult: %s\nHuman Feedback: %s",
		msg.Content,
		result,
		feedback,
	)
}

func AIContext(
	memories []string,
	history []api.Message,
	model string,
	maxTokens int,
) ([]api.Message, error) {
	counter, err := tokens.NewCounter(model)
	if err != nil {
		return nil, fmt.Errorf("failed to create counter: %w", err)
	}

	now := time.Now()
	messages := []api.Message{
		{
			Role:    "system",
			Content: fmt.Sprintf("The current time and date is %s", now.Format(time.RFC3339)),
		},
		{
			Role: "system",
			Content: fmt.Sprintf(
				"This reminds you of these events from your past: %s",
				strings.Join(memories, "\n\n"),
			),
		},
	}

	for _, message := range messages {
		counter.Add(message)
	}

	// walk history backwards adding messages until we reach maxTokens
	for idx := len(history) - 1; idx >= 0; idx-- {
		counter.Add(history[idx])

		if counter.Tokens >= maxTokens {
			break
		}

		messages = append(messages, history[idx])
	}

	return messages, nil
}

func Completion( //nolint:ireturn
	ctx context.Context,
	manager *plugin.Manager,
	name string,
) (api.Completion, func() error, error) {
	client, err := manager.Client(ctx, name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get client: %w", err)
	}

	protocol, err := client.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get protocol: %w", err)
	}

	raw, err := protocol.Dispense("completion")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dispense: %w", err)
	}

	completion, ok := raw.(api.Completion)
	if !ok {
		return nil, nil, fmt.Errorf("failed to cast completion: %w", plugin.ErrUnexpectedInterface)
	}

	return completion, protocol.Close, nil
}

func Memory( //nolint:ireturn
	ctx context.Context,
	manager *plugin.Manager,
	name string,
) (api.Memory, func() error, error) {
	client, err := manager.Client(ctx, name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get client: %w", err)
	}

	protocol, err := client.Client()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get protocol: %w", err)
	}

	raw, err := protocol.Dispense("memory")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dispense: %w", err)
	}

	memory, ok := raw.(api.Memory)
	if !ok {
		return nil, nil, fmt.Errorf("failed to cast memory: %w", plugin.ErrUnexpectedInterface)
	}

	return memory, protocol.Close, nil
}
