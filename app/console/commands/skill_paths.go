package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/goravel/installer/internal/fsutil"
)

// Environment variables OpenCode honors when locating its configuration directory.
const (
	envOpenCodeConfigDir = "OPENCODE_CONFIG_DIR"
	envXDGConfigHome     = "XDG_CONFIG_HOME"
)

// resolveDestination turns a raw --path value into an absolute destination
// directory. An empty value defers to openCodeSkillsDir; an explicit value wins
// over OpenCode's environment and is home-expanded and made absolute. The caller
// reads the environment once at the command boundary and passes it in. The
// result is symlink-resolved and the user's home directory and the filesystem
// root are rejected.
func resolveDestination(rawPath, openCodeConfig, xdgConfig string) (string, error) {
	destination := strings.TrimSpace(rawPath)
	if destination == "" {
		resolved, err := openCodeSkillsDir(openCodeConfig, xdgConfig)
		if err != nil {
			return "", err
		}

		return rejectUnsafeDestination(resolved)
	}

	expanded, err := fsutil.ExpandHomePath(destination)
	if err != nil {
		return "", err
	}

	destination, err = filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("failed to resolve skills path: %w", err)
	}

	return rejectUnsafeDestination(destination)
}

// rejectUnsafeDestination refuses the user's home directory and the filesystem
// root, which would scatter skill folders across sensitive locations. It returns
// the symlink-resolved (canonical) path so the validated path and the written
// path are the same.
func rejectUnsafeDestination(destination string) (string, error) {
	canonical := canonicalDestination(destination)

	if home, err := fsutil.HomeDir(); err == nil {
		if resolved, rerr := filepath.EvalSymlinks(home); rerr == nil {
			home = resolved
		}
		if strings.EqualFold(filepath.Clean(canonical), filepath.Clean(home)) {
			return "", fmt.Errorf("refusing to use the home directory %q as the skills destination", destination)
		}
	}

	root := filepath.VolumeName(canonical) + string(filepath.Separator)
	if strings.EqualFold(filepath.Clean(canonical), filepath.Clean(root)) {
		return "", fmt.Errorf("refusing to use the filesystem root %q as the skills destination", destination)
	}

	return canonical, nil
}

// canonicalDestination resolves symlinks on destination, falling back to the
// deepest existing ancestor when destination itself does not exist yet.
func canonicalDestination(destination string) string {
	if resolved, err := filepath.EvalSymlinks(destination); err == nil {
		return resolved
	}

	dir := filepath.Dir(destination)
	for {
		if resolved, err := filepath.EvalSymlinks(dir); err == nil {
			rel, relErr := filepath.Rel(dir, destination)
			if relErr != nil {
				return destination
			}

			return filepath.Join(resolved, rel)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return destination
		}
		dir = parent
	}
}

// openCodeSkillsDir resolves the default directory where OpenCode discovers agent
// skills: the first usable candidate of $OPENCODE_CONFIG_DIR, then
// $XDG_CONFIG_HOME/opencode, and finally ~/.config/opencode, with a "skills"
// folder appended. Each value is trimmed and home-expanded; only absolute (or
// "~"-prefixed) values are honored. A non-empty but unusable value (relative or
// "~user/...") is intentionally skipped so it cannot resolve against the working
// directory, falling through to the next candidate with no error.
func openCodeSkillsDir(openCodeConfig, xdgConfig string) (string, error) {
	if configDir, err := fsutil.ExpandHomePath(strings.TrimSpace(openCodeConfig)); err == nil && filepath.IsAbs(configDir) {
		return filepath.Join(configDir, "skills"), nil
	}

	if configHome, err := fsutil.ExpandHomePath(strings.TrimSpace(xdgConfig)); err == nil && filepath.IsAbs(configHome) {
		return filepath.Join(configHome, "opencode", "skills"), nil
	}

	home, err := fsutil.HomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".config", "opencode", "skills"), nil
}
