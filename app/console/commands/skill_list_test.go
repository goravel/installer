package commands

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListSkills(t *testing.T) {
	t.Run("returns only directories", func(t *testing.T) {
		skillsPath := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(skillsPath, "goravel-planning"), 0o755))
		require.NoError(t, os.MkdirAll(filepath.Join(skillsPath, "goravel-testing"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(skillsPath, "README.md"), []byte("x"), 0o644))

		got, err := listSkills(skillsPath)
		require.NoError(t, err)
		assert.Equal(t, []string{"goravel-planning", "goravel-testing"}, got)
	})

	t.Run("missing directory yields empty result", func(t *testing.T) {
		got, err := listSkills(filepath.Join(t.TempDir(), "missing"))
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("read error is surfaced", func(t *testing.T) {
		notDir := filepath.Join(t.TempDir(), "not-a-dir")
		require.NoError(t, os.WriteFile(notDir, []byte("x"), 0o644))

		_, err := listSkills(notDir)
		assert.Error(t, err)
	})

	t.Run("symlinked entry is rejected", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("symlinks are not reliable on windows")
		}

		skillsPath := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(skillsPath, "real"), 0o755))
		require.NoError(t, os.Symlink(filepath.Join(skillsPath, "real"), filepath.Join(skillsPath, "link")))

		_, err := listSkills(skillsPath)
		assert.ErrorContains(t, err, "is not a directory")
	})
}

func TestListSkillDetailsToleratesUnreadableDescription(t *testing.T) {
	skillsPath := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(skillsPath, "with-desc"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(skillsPath, "with-desc", "SKILL.md"),
		[]byte("---\ndescription: hello\n---\n"),
		0o644,
	))
	// A skill directory without a SKILL.md must not abort the listing.
	require.NoError(t, os.MkdirAll(filepath.Join(skillsPath, "no-desc"), 0o755))

	got, err := listSkillDetails(skillsPath, true)
	require.NoError(t, err)
	assert.Equal(t, []skillDetail{
		{Name: "no-desc"},
		{Name: "with-desc", Description: "hello"},
	}, got)
}
