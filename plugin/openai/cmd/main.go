//

package main

import (
	"os"

	"github.com/hashicorp/go-plugin"

	"github.com/lazygpt/lazygpt/plugin/api"
	"github.com/lazygpt/lazygpt/plugin/openai/pkg/openai"
)

const OpenAIAPIKeyEnv = "OPENAI_API_KEY" //nolint:gosec // this is the env var name.

func main() {
	key, ok := os.LookupEnv(OpenAIAPIKeyEnv)
	if !ok {
		// NOTE(jkoelker) This is a hack for now until we have better config
		//                management.
		panic("OPENAI_API_KEY not set")
	}

	openaiPlugin := openai.NewPlugin(key)

	config := &plugin.ServeConfig{
		HandshakeConfig: api.HandshakeConfig(),
		GRPCServer:      plugin.DefaultGRPCServer,

		Plugins: plugin.PluginSet{
			"completion": api.NewCompletionPlugin(openaiPlugin),
			"interfaces": api.NewInterfacesPlugin(openaiPlugin),
		},
	}

	plugin.Serve(config)
}
