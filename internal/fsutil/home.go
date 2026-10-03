package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExpandHomePath expands a leading "~", "~/", or "~\" in path to the user's home
// directory. A bare "~" resolves to the home directory itself, and any other
// leading "~" (for example "~user/...") is rejected as unsupported.
func ExpandHomePath(path string) (string, error) {
	switch {
	case path == "~":
		return HomeDir()
	case strings.HasPrefix(path, "~/"), strings.HasPrefix(path, `~\`):
		home, err := HomeDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(home, path[2:]), nil
	case strings.HasPrefix(path, "~"):
		return "", fmt.Errorf("unsupported home path %q", path)
	default:
		return path, nil
	}
}

// HomeDir returns the current user's home directory, ensuring it is absolute so
// callers never resolve against the working directory.
func HomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	if !filepath.IsAbs(home) {
		return "", fmt.Errorf("home directory %q is not absolute", home)
	}

	return home, nil
}
