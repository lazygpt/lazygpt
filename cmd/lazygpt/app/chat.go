//

package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

func InitChatCmd(app *LazyGPTApp) {
	chatCmd := &cobra.Command{
		Use:   "chat",
		Short: "Start an interactive chat session with LazyGPT",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: Implement the chat functionality
			fmt.Println("Starting interactive chat session...")
		},
	}

	app.RootCmd.AddCommand(chatCmd)
}
