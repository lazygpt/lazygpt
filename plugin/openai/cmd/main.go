package main

import (
	"os"

	"github.com/hashicorp/go-plugin"
	"github.com/lazygpt/lazygpt/plugin/api"

	"github.com/lazygpt/lazygpt/plugin/openai/pkg/openai"
)

const OpenAIAPIKeyEnv = "OPENAI_API_KEY"

func main() {
	key, ok := os.LookupEnv(OpenAIAPIKeyEnv)
	if !ok {
		// NOTE(jkoelker) This is a hack for now until we have better config
		//                management.
		panic("OPENAI_API_KEY not set")
	}

	config := &plugin.ServeConfig{
		HandshakeConfig: api.HandshakeConfig(),
		GRPCServer: plugin.DefaultGRPCServer,

		Plugins: plugin.PluginSet{
			"completion": api.NewCompletionPlugin(openai.NewOpenAIPlugin(key)),
		},
	}

	plugin.Serve(config)
}

