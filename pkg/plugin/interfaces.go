//

package plugin

import (
	"context"
	"errors"
	"fmt"

	"github.com/lazygpt/lazygpt/plugin/api"
)

// ErrUnexpectedInterface is returned when the plugin does not support the
// expected interface.
var ErrUnexpectedInterface = errors.New("unexpected interface")

// Intrerfaces discovers the interfaces supported by the plugin.
func Interfaces(ctx context.Context, path string) ([]string, error) {
	client, err := Factory(ctx, path, []string{"interfaces"})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	if _, err := client.Start(); err != nil {
		return nil, fmt.Errorf("failed to start client: %w", err)
	}

	defer client.Kill()

	protocol, err := client.Client()
	if err != nil {
		return nil, fmt.Errorf("failed to get client protocol: %w", err)
	}

	defer protocol.Close()

	raw, err := protocol.Dispense("interfaces")
	if err != nil {
		return nil, fmt.Errorf("failed to dispense interfaces: %w", err)
	}

	plugin, ok := raw.(api.Interfaces)
	if !ok {
		return nil, ErrUnexpectedInterface
	}

	interfaces, err := plugin.Interfaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %w", err)
	}

	return interfaces, nil
}
