//

package plugin

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-plugin"

	"github.com/lazygpt/lazygpt/plugin/api"
	"github.com/lazygpt/lazygpt/plugin/log"
)

// ErrUnsupportedPluginInterface is returned when an interface is requested
// that is not supported.
var ErrUnsupportedPluginInterface = fmt.Errorf("unsupported plugin interface")

// Factory is a factory for creating plugin clients for the interfaces
// requested.
func Factory(ctx context.Context, path string, interfaces []string) (*plugin.Client, error) {
	plugins := make(plugin.PluginSet)
	supported := api.Plugins()

	for _, iface := range interfaces {
		if _, ok := supported[iface]; !ok {
			return nil, fmt.Errorf("%w: %s", ErrUnsupportedPluginInterface, iface)
		}

		plugins[iface] = supported[iface]
	}

	logger := log.FromContext(ctx).Named("plugin")

	config := &plugin.ClientConfig{
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		AutoMTLS:         true,
		Cmd:              exec.Command(path),
		HandshakeConfig:  api.HandshakeConfig(),
		Managed:          true,
		Plugins:          plugins,
		Logger:           logger,
	}

	return plugin.NewClient(config), nil
}
