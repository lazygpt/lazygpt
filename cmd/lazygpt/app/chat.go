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

	MaxTokens      = 4096
	ResponseTokens = 1024
	SendTokens     = MaxTokens - ResponseTokens

	// NOTE(jkoelker) Allow 80% of the tokens to be used for the memory.
	//                Perform this calculation in integer space to avoid
	//                floating point errors.
	MemoryTokens  = SendTokens * 80 / 100
	HistoryTokens = SendTokens - MemoryTokens
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

			execute := Executor(ctx, completion, memory)

			if err := execute(Prompt(), "system"); err != nil {
				return fmt.Errorf("failed to set initial prompt: %w", err)
			}

			prompt.New(
				func(in string) {
					input := strings.TrimSpace(in)

					if input == "" {
						return
					}

					if input == "exit" {
						os.Exit(0)
					}

					if err := execute(input, "user"); err != nil {
						log.Error(ctx, "failed to execute", err)

						return
					}
				},
				func(_ prompt.Document) []prompt.Suggest { return []prompt.Suggest{} },
				prompt.OptionPrefix("> "),
			).Run()

			return nil
		},
	}

	app.RootCmd.AddCommand(chatCmd)
}

func Executor(
	ctx context.Context,
	completion api.Completion,
	memory api.Memory,
) func(string, string) error {
	var history []api.Message

	return func(input string, role string) error {
		memories := make([]string, 0, MaxMemory)

		for i := len(history) - 1; i >= 0 && len(memories) < MaxMemory; i-- {
			memories = append(memories, Memorize(&history[i], "", ""))
		}

		recollection, err := Recollection(ctx, memory, memories)
		if err != nil {
			return fmt.Errorf("failed to recollect: %w", err)
		}

		history = append(history, api.Message{
			Role:    role,
			Content: input,
		})

		context, tokens, err := AIContext(
			ctx,
			recollection,
			history,
			"gpt-3.5-turbo",
			MemoryTokens,
			HistoryTokens,
		)
		if err != nil {
			return fmt.Errorf("failed to create context: %w", err)
		}

		log.Info(ctx, "Thinking...", "context", context, "tokens", tokens)

		response, reason, err := completion.Complete(ctx, context)
		if err != nil || response == nil {
			log.Error(
				ctx, "failed to complete", err,
				"response", response,
				"reason", reason,
			)

			return fmt.Errorf("failed to complete: %w", err)
		}

		fmt.Println(response.Content) //nolint:forbidigo // this is a CLI app

		history = append(history, api.Message{
			Role:    response.Role,
			Content: response.Content,
		})

		if err := memory.Memorize(
			ctx,
			[]string{Memorize(response, "", input)},
		); err != nil {
			return fmt.Errorf("failed to memorize: %w", err)
		}

		return nil
	}
}

func Prompt() string {
	constraints := `
1. ~4000 word limit for short term memory. Your short term memory is short, so
immediately save important information to files.
2. If you are unsure how you previously did something or want to recall past
events, thinking about similar events will help you remember
`
	commands := ""
	resources := ""
	performance := `
1. Continuously review and analyze your actions to ensure you are performing to
the best of your abilities.
2. Constructively self-criticize your big-picture behavior constantly.
3. Reflect on past decisions and strategies to refine your approach.
4. Every command has a cost, so be smart and efficient. Aim to complete tasks in
the least number of steps.
`

	return strings.Join(
		[]string{
			"Constraints:", constraints,
			"Commands:", commands,
			"Resources:", resources,
			"Performance Evaluation:", performance,
		},
		"\n",
	)
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
	_ context.Context,
	memories []string,
	history []api.Message,
	model string,
	memoriesTokens int,
	historyTokens int,
) ([]api.Message, int, error) {
	counter, err := tokens.NewCounter(model)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create counter: %w", err)
	}

	now := time.Now()
	messages := []api.Message{
		{
			Role:    "system",
			Content: fmt.Sprintf("The current time and date is %s", now.Format(time.RFC3339)),
		},
	}

	for _, message := range messages {
		counter.Add(message)
	}

	for _, memory := range memories {
		message := api.Message{
			Role:    "system",
			Content: fmt.Sprintf("This reminds you of this event from your past: %s", memory),
		}

		counter.Add(message)

		if counter.Tokens >= memoriesTokens {
			break
		}

		messages = append(messages, message)
	}

	maxHistoryIdx := len(history) - 1

	// NOTE(jkoelker) Walk history backwards adding messages until we reach the
	//                max number of history tokens.
	for idx := len(history) - 1; idx >= 0; idx-- {
		counter.Add(history[idx])

		if counter.Tokens >= historyTokens {
			maxHistoryIdx = idx - 1

			break
		}
	}

	// NOTE(jkoelker) Add history messages in reverse order so they are in the
	//                correct order.
	for idx := len(history) - 1; idx >= maxHistoryIdx; idx-- {
		messages = append(messages, history[idx])
	}

	return messages, counter.Tokens, nil
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
