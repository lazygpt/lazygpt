//

//nolint:forbidigo
package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"

	"github.com/lazygpt/lazygpt/pkg/plugin"
	"github.com/lazygpt/lazygpt/plugin/api"
)

func InitChatCmd(app *LazyGPTApp) {
	chatCmd := &cobra.Command{
		Use:   "chat",
		Short: "Start an interactive chat session with LazyGPT",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := plugin.NewManager()
			defer manager.Close()

			client, err := manager.Client(cmd.Context(), "openai")
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

			executor := func(in string) {
				input := strings.TrimSpace(in)

				if input == "" {
					return
				}

				if input == "exit" {
					os.Exit(0)
				}

				messages = append(messages, api.Message{
					Role:    "user",
					Content: input,
				})

				response, reason, err := completion.Complete(cmd.Context(), messages)
				if err != nil {
					fmt.Printf("Error: %s\n", err)

					return
				}

				if response == nil {
					fmt.Printf("Reason: %s\n", reason)

					return
				}
				fmt.Printf("Reason: %s\n", reason)
				fmt.Printf("Response: %s\n", response)

				messages = append(messages, api.Message{
					Role:    response.Role,
					Content: response.Content,
				})
			}

			prompt.New(
				executor,
				func(_ prompt.Document) []prompt.Suggest { return []prompt.Suggest{} },
				prompt.OptionPrefix("> "),
			).Run()

			return nil
		},
	}

	app.RootCmd.AddCommand(chatCmd)
}
