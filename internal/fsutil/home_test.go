package fsutil

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setHomeDir(t *testing.T, home string) {
	t.Helper()

	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
}

func TestHomeDir(t *testing.T) {
	t.Run("absolute home", func(t *testing.T) {
		home := t.TempDir()
		setHomeDir(t, home)

		got, err := HomeDir()
		assert.NoError(t, err)
		assert.Equal(t, home, got)
	})

	t.Run("relative home is rejected", func(t *testing.T) {
		setHomeDir(t, "relative/home")

		_, err := HomeDir()
		assert.ErrorContains(t, err, "is not absolute")
	})

	t.Run("missing home is rejected", func(t *testing.T) {
		setHomeDir(t, "")

		_, err := HomeDir()
		assert.ErrorContains(t, err, "failed to get home directory")
	})
}

func TestExpandHomePath(t *testing.T) {
	home := t.TempDir()

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{name: "absolute path unchanged", path: "/tmp/skills", want: "/tmp/skills"},
		{name: "relative path unchanged", path: "relative/skills", want: "relative/skills"},
		{name: "bare tilde expands to home", path: "~", want: home},
		{name: "home-relative expands", path: "~/goravel", want: filepath.Join(home, "goravel")},
		{name: "windows home-relative expands", path: `~\goravel`, want: filepath.Join(home, "goravel")},
		{name: "unsupported ~user rejects", path: "~bob/skills", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setHomeDir(t, home)

			got, err := ExpandHomePath(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("bare tilde fails when home unavailable", func(t *testing.T) {
		setHomeDir(t, "")

		_, err := ExpandHomePath("~")
		assert.ErrorContains(t, err, "failed to get home directory")
	})

	t.Run("home-relative fails when home unavailable", func(t *testing.T) {
		setHomeDir(t, "")

		_, err := ExpandHomePath("~/goravel")
		assert.ErrorContains(t, err, "failed to get home directory")
	})
}
