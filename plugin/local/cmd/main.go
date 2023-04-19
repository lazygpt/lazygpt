//

package main

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/go-plugin"

	"github.com/lazygpt/lazygpt/plugin/api"
	"github.com/lazygpt/lazygpt/plugin/local/pkg/local"
)

func main() {
	// NOTE(jkoelker) This is a temporary hack until we can figure out
	//                plugin configuration.
	dataDir := os.Getenv("LAZYGPT_LOCAL_DATA_DIR")
	if dataDir == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			panic(err)
		}

		dataDir = filepath.Join(configDir, "lazygpt", "local")

		if err := os.MkdirAll(dataDir, os.ModePerm); err != nil && !os.IsExist(err) {
			panic(err)
		}
	}

	localPlugin := local.NewPlugin(dataDir)

	config := &plugin.ServeConfig{
		HandshakeConfig: api.HandshakeConfig(),
		GRPCServer:      plugin.DefaultGRPCServer,

		Plugins: plugin.PluginSet{
			"memory":     api.NewMemoryPlugin(localPlugin),
			"interfaces": api.NewInterfacesPlugin(localPlugin),
		},
	}

	plugin.Serve(config)
}
