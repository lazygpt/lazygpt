//

package local

import (
	"context"
	"fmt"

	"github.com/lazygpt/lazygpt/plugin/api"
	"github.com/lazygpt/lazygpt/plugin/local/pkg/memory"
)

type Plugin struct {
	Memory *memory.Local
}

var (
	_ api.Memory     = (*Plugin)(nil)
	_ api.Interfaces = (*Plugin)(nil)
)

// NewPlugin creates a new Plugin instance.
func NewPlugin(datadir string) *Plugin {
	return &Plugin{
		Memory: memory.NewLocal(datadir),
	}
}

// Open opens the local memory database.
func (plugin *Plugin) Open(ctx context.Context) error {
	if err := plugin.Memory.Open(ctx); err != nil {
		return fmt.Errorf("failed to open local memory: %w", err)
	}

	return nil
}

// Close closes the local memory database.
func (plugin *Plugin) Close(ctx context.Context) error {
	if err := plugin.Memory.Close(ctx); err != nil {
		return fmt.Errorf("failed to close local memory: %w", err)
	}

	return nil
}

// Memorize implements the `api.Memory` interface.
func (plugin *Plugin) Memorize(ctx context.Context, data []string) error {
	if err := plugin.Memory.Memorize(ctx, data); err != nil {
		return fmt.Errorf("failed to memorize data: %w", err)
	}

	return nil
}

// Recall implements the `api.Memory` interface.
func (plugin *Plugin) Recall(ctx context.Context, data string, count ...int) ([]string, error) {
	memories, err := plugin.Memory.Recall(ctx, data, count...)
	if err != nil {
		return nil, fmt.Errorf("failed to recall data: %w", err)
	}

	return memories, nil
}

// Interfaces implements the `api.Interfaces` interface.
func (plugin *Plugin) Interfaces(_ context.Context) ([]string, error) {
	return []string{
		"interfaces",
		"memory",
	}, nil
}
