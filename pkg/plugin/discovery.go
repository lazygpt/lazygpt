//

package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DiscoverPlugins looks for available plugins in the provided directories.
func DiscoverPlugins(dirs []string) ([]string, error) {
	plugins := []string{}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory %q: %w", dir, err)
		}

		for _, entry := range entries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), "lazygpt-plugin-") {
				plugins = append(plugins, filepath.Join(dir, entry.Name()))
			}
		}
	}

	return plugins, nil
}

// FindPlugins looks for available plugins in the same directory as the exeuatable
// or in the provided directories.
func FindPlugins(dirs []string) ([]string, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	exeDir := filepath.Dir(exe)
	dirs = append([]string{exeDir}, dirs...)

	return DiscoverPlugins(dirs)
}

// ResolvePlugins resolves all plugins found in the provided directories and
// returns a `map[string]string` containing all plugins names and their path.
func ResolvePlugins(dirs []string) (map[string]string, error) {
	plugins := make(map[string]string)

	paths, err := FindPlugins(dirs)
	if err != nil {
		return nil, fmt.Errorf("failed to find plugins: %w", err)
	}

	for _, path := range paths {
		name := strings.TrimPrefix(filepath.Base(path), "lazygpt-plugin-")
		plugins[name] = path
	}

	return plugins, nil
}
