package commands

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenCodeSkillsDir(t *testing.T) {
	home := t.TempDir()
	setHomeDir(t, home)

	configDir := t.TempDir()
	configHome := t.TempDir()

	tests := []struct {
		name     string
		openCode string
		xdg      string
		want     string
	}{
		{
			name:     "OpenCode config dir wins over xdg",
			openCode: configDir,
			xdg:      configHome,
			want:     filepath.Join(configDir, "skills"),
		},
		{
			name:     "home-relative OpenCode config dir is expanded",
			openCode: "~/custom",
			xdg:      configHome,
			want:     filepath.Join(home, "custom", "skills"),
		},
		{
			name: "home-relative xdg config home is expanded",
			xdg:  "~/xdg",
			want: filepath.Join(home, "xdg", "opencode", "skills"),
		},
		{
			name: "xdg config home used when OpenCode unset",
			xdg:  configHome,
			want: filepath.Join(configHome, "opencode", "skills"),
		},
		{
			name:     "relative OpenCode config dir is ignored",
			openCode: "relative/config",
			xdg:      configHome,
			want:     filepath.Join(configHome, "opencode", "skills"),
		},
		{
			name:     "unsupported ~user OpenCode value is skipped",
			openCode: "~bob/opencode",
			xdg:      configHome,
			want:     filepath.Join(configHome, "opencode", "skills"),
		},
		{
			name: "relative xdg config home is ignored",
			xdg:  "relative/config",
			want: filepath.Join(home, ".config", "opencode", "skills"),
		},
		{
			name: "unsupported ~user xdg value is skipped",
			xdg:  "~bob/opencode",
			want: filepath.Join(home, ".config", "opencode", "skills"),
		},
		{
			name:     "whitespace-only values fall back to home config",
			openCode: "   ",
			xdg:      "\t",
			want:     filepath.Join(home, ".config", "opencode", "skills"),
		},
		{
			name:     "padded absolute OpenCode value is trimmed",
			openCode: "  " + configDir + "  ",
			xdg:      configHome,
			want:     filepath.Join(configDir, "skills"),
		},
		{
			name:     "padded home-relative OpenCode value is trimmed and expanded",
			openCode: "  ~/custom  ",
			xdg:      configHome,
			want:     filepath.Join(home, "custom", "skills"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := openCodeSkillsDir(tt.openCode, tt.xdg)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("home unavailable errors on fallback", func(t *testing.T) {
		setHomeDir(t, "")

		_, err := openCodeSkillsDir("", "")
		assert.ErrorContains(t, err, "failed to get home directory")
	})
}

func TestResolveDestination(t *testing.T) {
	home := canonicalTempDir(t)
	setHomeDir(t, home)

	configDir := canonicalTempDir(t)
	envDir := canonicalTempDir(t)

	tests := []struct {
		name     string
		rawPath  string
		openCode string
		xdg      string
		want     string
		wantErr  bool
	}{
		{
			name: "empty path uses default",
			want: filepath.Join(home, ".config", "opencode", "skills"),
		},
		{
			name:     "explicit path wins over env",
			rawPath:  configDir,
			openCode: envDir,
			xdg:      t.TempDir(),
			want:     configDir,
		},
		{
			name:    "bare tilde resolves to home and is rejected",
			rawPath: "~",
			wantErr: true,
		},
		{
			name:    "home-relative path expands",
			rawPath: "~/goravel-skills",
			want:    filepath.Join(home, "goravel-skills"),
		},
		{
			name:    "filesystem root is rejected",
			rawPath: string(filepath.Separator),
			wantErr: true,
		},
		{
			name:    "whitespace-only path falls back to default",
			rawPath: "   ",
			want:    filepath.Join(home, ".config", "opencode", "skills"),
		},
		{
			name:    "padded path is trimmed and expanded",
			rawPath: "  ~/goravel-skills  ",
			want:    filepath.Join(home, "goravel-skills"),
		},
		{
			name:    "unsupported ~user path errors",
			rawPath: "~bob/skills",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveDestination(tt.rawPath, tt.openCode, tt.xdg)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveDestinationRejectsSymlinkToHome(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks are not reliable on windows")
	}

	home := t.TempDir()
	setHomeDir(t, home)

	link := filepath.Join(t.TempDir(), "home-link")
	require.NoError(t, os.Symlink(home, link))

	_, err := resolveDestination(link, "", "")
	assert.ErrorContains(t, err, "refusing to use the home directory")
}
