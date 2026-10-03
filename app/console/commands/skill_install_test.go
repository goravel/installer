package commands

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeSkillNames(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    []string
		wantErr bool
	}{
		{name: "keeps names and trims", input: []string{"  a ", "b"}, want: []string{"a", "b"}},
		{name: "skips blank entries", input: []string{"", "  ", "a"}, want: []string{"a"}},
		{name: "deduplicates", input: []string{"a", "a", " b ", "b"}, want: []string{"a", "b"}},
		{name: "all blank yields empty", input: []string{"", " "}, want: []string{}},
		{name: "rejects dot", input: []string{"."}, wantErr: true},
		{name: "rejects dotdot", input: []string{".."}, wantErr: true},
		{name: "rejects forward slash", input: []string{"a/b"}, wantErr: true},
		{name: "rejects backslash", input: []string{`a\b`}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeSkillNames(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveSkills(t *testing.T) {
	t.Run("rejects a regular file", func(t *testing.T) {
		skillsPath := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(skillsPath, "regular"), []byte("x"), 0o644))

		_, err := resolveSkills(skillsPath, []string{"regular"})
		assert.ErrorContains(t, err, "is not a directory")
	})

	t.Run("rejects a symlinked skill", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("symlinks are not reliable on windows")
		}

		skillsPath := filepath.Join(t.TempDir(), "skills")
		require.NoError(t, os.MkdirAll(filepath.Join(skillsPath, "real"), 0o755))
		require.NoError(t, os.Symlink(filepath.Join(skillsPath, "real"), filepath.Join(skillsPath, "link")))

		_, err := resolveSkills(skillsPath, []string{"link"})
		assert.ErrorContains(t, err, "is not a directory")
	})

	t.Run("empty names lists all skills", func(t *testing.T) {
		skillsPath := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(skillsPath, "goravel-testing"), 0o755))

		got, err := resolveSkills(skillsPath, nil)
		assert.NoError(t, err)
		assert.Equal(t, []string{"goravel-testing"}, got)
	})
}

func TestInstallSkillFailedCopyKeepsPreviousInstall(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks are not reliable on windows")
	}

	skillsPath := filepath.Join(t.TempDir(), "skills")
	writeSkillContent(t, skillsPath, "goravel-testing", "new skill")
	// A nested symlink makes the recursive copy fail partway through.
	require.NoError(t, os.Symlink(
		filepath.Join(skillsPath, "goravel-testing", "SKILL.md"),
		filepath.Join(skillsPath, "goravel-testing", "link"),
	))

	destination := filepath.Join(t.TempDir(), "dest")
	writeSkillContent(t, destination, "goravel-testing", "old skill")

	_, err := installSkill(skillsPath, destination, "goravel-testing", true, os.Rename)
	require.Error(t, err)
	assert.ErrorContains(t, err, "unsupported file type")

	assert.Equal(t, "old skill", readSkillContent(t, destination, "goravel-testing"))

	entries, err := os.ReadDir(destination)
	require.NoError(t, err)
	assert.Equal(t, []string{"goravel-testing"}, entryNames(entries))
}

func TestInstallSkillRestoresBackupOnRenameFailure(t *testing.T) {
	skillsPath := filepath.Join(t.TempDir(), "skills")
	writeSkillContent(t, skillsPath, "goravel-testing", "new skill")

	destination := filepath.Join(t.TempDir(), "dest")
	writeSkillContent(t, destination, "goravel-testing", "old skill")

	// The rename is injected rather than a package global, so this test needs no
	// cleanup and is safe to run alongside other tests (no shared mutable state).
	rename := func(oldpath, newpath string) error {
		if strings.Contains(oldpath, "-staging-") {
			return errors.New("forced rename failure")
		}

		return os.Rename(oldpath, newpath)
	}

	_, err := installSkill(skillsPath, destination, "goravel-testing", true, rename)
	require.Error(t, err)
	assert.ErrorContains(t, err, "forced rename failure")

	assert.Equal(t, "old skill", readSkillContent(t, destination, "goravel-testing"))

	entries, err := os.ReadDir(destination)
	require.NoError(t, err)
	assert.Equal(t, []string{"goravel-testing"}, entryNames(entries))
}

func TestInstallSkillReplaceFailureKeepsTarget(t *testing.T) {
	skillsPath := filepath.Join(t.TempDir(), "skills")
	writeSkillContent(t, skillsPath, "goravel-testing", "new skill")

	destination := filepath.Join(t.TempDir(), "dest")
	writeSkillContent(t, destination, "goravel-testing", "old skill")

	// Fail the rename that moves the existing target aside.
	rename := func(oldpath, newpath string) error {
		if strings.Contains(newpath, "-backup-") {
			return errors.New("forced backup rename failure")
		}

		return os.Rename(oldpath, newpath)
	}

	_, err := installSkill(skillsPath, destination, "goravel-testing", true, rename)
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to replace existing skill")

	assert.Equal(t, "old skill", readSkillContent(t, destination, "goravel-testing"))

	entries, err := os.ReadDir(destination)
	require.NoError(t, err)
	assert.Equal(t, []string{"goravel-testing"}, entryNames(entries), "staging must be removed")
}

func TestInstallSkillReportsRestoreFailure(t *testing.T) {
	skillsPath := filepath.Join(t.TempDir(), "skills")
	writeSkillContent(t, skillsPath, "goravel-testing", "new skill")

	destination := filepath.Join(t.TempDir(), "dest")
	writeSkillContent(t, destination, "goravel-testing", "old skill")

	// Fail both the staging->target swap and the backup->target restore so the
	// joined error is reported and the backup is left in place.
	rename := func(oldpath, newpath string) error {
		if strings.Contains(oldpath, "-staging-") || strings.Contains(oldpath, "-backup-") {
			return errors.New("forced rename failure")
		}

		return os.Rename(oldpath, newpath)
	}

	_, err := installSkill(skillsPath, destination, "goravel-testing", true, rename)
	require.Error(t, err)
	assert.ErrorContains(t, err, "forced rename failure")
	assert.ErrorContains(t, err, "failed to restore previous skill")

	entries, err := os.ReadDir(destination)
	require.NoError(t, err)
	require.Len(t, entries, 1)

	backup := filepath.Join(destination, entries[0].Name())
	require.Contains(t, entries[0].Name(), "-backup-", "backup should survive")
	assert.Equal(t, "old skill", readFile(t, filepath.Join(backup, "SKILL.md")))
}
