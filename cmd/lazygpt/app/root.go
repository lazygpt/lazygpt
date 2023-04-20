//

package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/lazygpt/lazygpt/plugin/log"
)

type LazyGPTApp struct {
	ConfigFile string
	RootCmd    *cobra.Command
}

func NewLazyGPTApp() *LazyGPTApp {
	app := &LazyGPTApp{}

	app.RootCmd = &cobra.Command{
		Use:   "lazygpt",
		Short: "LazyGPT is an autonomous agent using GPT and plugins to achieve goals",
		Long: `LazyGPT is an autonomous agent that, given a name, role, and goals,
uses GPT or other language models to develop a plan and implement it. Everything
from the language model to the various commands exposed to the model is
implemented with plugins. The program can either run on the CLI or start a web
server and serve a web UI.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logLevel, err := cmd.Flags().GetString("log-level")
			if err != nil {
				return fmt.Errorf("can't get log-level: %w", err)
			}

			ctx := log.NewContext(cmd.Context(), log.NewLogger("lazygpt", logLevel))
			cmd.SetContext(ctx)

			return nil
		},
	}

	app.RootCmd.PersistentFlags().StringVar(
		&app.ConfigFile,
		"config",
		"",
		"config file (default is lazygpt.yaml in user's config directory)",
	)

	app.RootCmd.PersistentFlags().StringP(
		"log-level",
		"l",
		"info",
		"log level (trace, debug, info, warn, error)",
	)

	InitChatCmd(app)
	InitServeCmd(app)

	return app
}

func (app *LazyGPTApp) Execute() {
	ctx := app.InitConfig()

	args := os.Args[1:]
	cmd, _, err := app.RootCmd.Find(args)

	// NOTE(jkoelker) Default to chat if no command is specified.
	if err == nil && cmd.Use == app.RootCmd.Use && !errors.Is(cmd.Flags().Parse(args), pflag.ErrHelp) {
		app.RootCmd.SetArgs(append([]string{"chat"}, args...))
	}

	if err := app.RootCmd.ExecuteContext(ctx); err != nil {
		log.Error(ctx, "Error executing command", err)
		os.Exit(1)
	}
}

func (app *LazyGPTApp) InitConfig() context.Context {
	ctx := log.NewContext(context.Background(), log.NewLogger("bootstrap"))

	if app.ConfigFile != "" {
		viper.SetConfigFile(app.ConfigFile)
	} else {
		configDir, err := os.UserConfigDir()
		if err != nil {
			log.Error(ctx, "Can't get user config directory", err)
			os.Exit(1)
		}

		configPath := filepath.Join(configDir, "lazygpt")
		viper.AddConfigPath(configPath)
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		var cmp viper.ConfigFileNotFoundError
		if !errors.As(err, &cmp) {
			log.Error(ctx, "Can't read config", err)
			os.Exit(1)
		}
	}

	return ctx
}
