//

package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

func InitServeCmd(app *LazyGPTApp) {
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start an HTTP server to serve the LazyGPT web UI",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Starting HTTP server...") //nolint:forbidigo
		},
	}

	app.RootCmd.AddCommand(serveCmd)
}
