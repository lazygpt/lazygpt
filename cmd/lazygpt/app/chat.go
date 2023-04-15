//

package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/spf13/cobra"

	"github.com/lazygpt/lazygpt/pkg/plugin"
	"github.com/lazygpt/lazygpt/plugin/api"
	"github.com/lazygpt/lazygpt/plugin/log"
)

var (
	ErrFailedSurvey    = errors.New("failed to ask survey")
	ErrPromptNotString = errors.New("prompt must be a string")
	ErrPromptEmpty     = errors.New("prompt can not be empty")
	ErrMessagesNil     = errors.New("messages can not be nil")
)

type promptExecutor func(context.Context, string) string

// promptValidator is a basic validator for the survey prompt.
func promptValidator(in interface{}) error {
	input, ok := in.(string)
	if !ok {
		return ErrPromptNotString
	}

	input = strings.TrimSpace(input)

	if input == "" {
		return ErrPromptEmpty
	}

	if input == "exit" {
		os.Exit(0)
	}

	return nil
}

// promptExecutorFactory creates a prompt executor that will send the user's prompt to the given completion plugin.
// The executor will also append the user's prompt and the plugin's response to the given messages slice.
func promptExecutorFactory(completion api.Completion, messages *[]api.Message) (promptExecutor, error) {
	if messages == nil {
		return nil, ErrMessagesNil
	}

	return func(ctx context.Context, in string) string {
		input := strings.TrimSpace(in)

		*messages = append(*messages, api.Message{
			Role:    "user",
			Content: input,
		})

		response, reason, err := completion.Complete(ctx, *messages)
		if err != nil || response == nil {
			log.Error(
				ctx, "failed to complete", err,
				"response", response,
				"reason", reason,
			)

			return ""
		}

		*messages = append(*messages, api.Message{
			Role:    response.Role,
			Content: response.Content,
		})

		return response.Content
	}, nil
}

func InitChatCmd(app *LazyGPTApp) {
	chatCmd := &cobra.Command{
		Use:   "chat",
		Short: "Start an interactive chat session with LazyGPT",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := plugin.NewManager()
			defer manager.Close()

			ctx := cmd.Context()

			client, err := manager.Client(ctx, "openai")
			if err != nil {
				return fmt.Errorf("failed to get client: %w", err)
			}

			protocol, err := client.Client()
			if err != nil {
				return fmt.Errorf("failed to get protocol: %w", err)
			}

			defer protocol.Close()

			raw, err := protocol.Dispense("completion")
			if err != nil {
				return fmt.Errorf("failed to dispense: %w", err)
			}

			completion, ok := raw.(api.Completion)
			if !ok {
				return fmt.Errorf("failed to cast completion: %w", plugin.ErrUnexpectedInterface)
			}

			var messages []api.Message
			executor, err := promptExecutorFactory(completion, &messages)
			if err != nil {
				return fmt.Errorf("failed to create prompt executor: %w", err)
			}

			prompt := &survey.Multiline{
				Message: "You: ",
				Default: "",
				Help:    "Type your message and press Enter to send it. Type 'exit' to quit.",
			}

			for {
				var input string
				err := survey.AskOne(
					prompt,
					&input,
					survey.WithValidator(survey.Required),
					survey.WithValidator(promptValidator))
				if err != nil {
					if errors.Is(err, terminal.InterruptErr) {
						log.Debug(ctx, "got ctrl+c, exiting")

						break
					}

					return fmt.Errorf("%w: %w", ErrFailedSurvey, err)
				}

				promptResponse := executor(ctx, input)
				if promptResponse == "" {
					log.Warn(ctx, "empty response from plugin")
				}

				fmt.Printf("LazyGPT: %s", promptResponse)
			}

			return nil
		},
	}

	app.RootCmd.AddCommand(chatCmd)
}
