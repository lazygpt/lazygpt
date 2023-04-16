package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/lazygpt/lazygpt/cmd/lazygpt/app/chat"
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

			pctx := chat.NewPromptContext(ctx, completion, nil)
			p := tea.NewProgram(chat.NewModel(pctx))
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("failed to run program: %w", err)
			}

			return nil
		},
	}

	app.RootCmd.AddCommand(chatCmd)
}
