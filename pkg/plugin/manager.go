//

package plugin

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/hashicorp/go-plugin"
)

// ErrPluginNotFound is returned when a plugin is not found.
var ErrPluginNotFound = errors.New("plugin not found")

type Manager struct {
	Clients map[string]*plugin.Client
	Dirs    []string

	interfaces map[string][]string
	paths      map[string]string
	mu         sync.Mutex
}

func NewManager(dirs ...string) *Manager {
	return &Manager{
		Clients: make(map[string]*plugin.Client),
		Dirs:    dirs,

		interfaces: make(map[string][]string),
		paths:      make(map[string]string),
	}
}

// Close closes all plugin clients.
func (manager *Manager) Close() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	for _, client := range manager.Clients {
		client.Kill()
	}

	manager.Clients = make(map[string]*plugin.Client)
	manager.interfaces = make(map[string][]string)
}

// resolvePlugin resolves a plugin name to a path.
func (manager *Manager) resolvePlugin(_ context.Context, name string) (string, error) {
	if path, ok := manager.paths[name]; ok {
		return path, nil
	}

	paths, err := ResolvePlugins(manager.Dirs)
	if err != nil {
		return "", fmt.Errorf("failed to resolve plugin: %w", err)
	}

	manager.paths = paths

	path, ok := paths[name]
	if !ok {
		return "", fmt.Errorf("%w: %q", ErrPluginNotFound, name)
	}

	return path, nil
}

// getInterfaces returns a list of interfaces for a given plugin.
func (manager *Manager) getInterfaces(ctx context.Context, name string) ([]string, error) {
	if list, ok := manager.interfaces[name]; ok {
		return list, nil
	}

	path, err := manager.resolvePlugin(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve plugin: %w", err)
	}

	list, err := Interfaces(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %w", err)
	}

	manager.interfaces[name] = list

	return list, nil
}

// client returns a client for the given plugin.
func (manager *Manager) client(ctx context.Context, name string) (*plugin.Client, error) {
	if client, ok := manager.Clients[name]; ok {
		return client, nil
	}

	interfaces, err := manager.getInterfaces(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %w", err)
	}

	path, err := manager.resolvePlugin(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve plugin: %w", err)
	}

	client, err := Factory(ctx, path, interfaces)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	if _, err := client.Start(); err != nil {
		return nil, fmt.Errorf("failed to start client: %w", err)
	}

	manager.Clients[name] = client

	return client, nil
}

// ResolvePlugin resolves a plugin name to a path.
func (manager *Manager) ResolvePlugin(ctx context.Context, name string) (string, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.resolvePlugin(ctx, name)
}

// Interfaces returns a list of interfaces for a given plugin.
func (manager *Manager) Interfaces(ctx context.Context, name string) ([]string, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.getInterfaces(ctx, name)
}

// Client returns a client for the given plugin.
func (manager *Manager) Client(ctx context.Context, name string) (*plugin.Client, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.client(ctx, name)
}
