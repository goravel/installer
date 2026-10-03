package fsutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopyDirectoryTree(t *testing.T) {
	source := filepath.Join(t.TempDir(), "src")
	require.NoError(t, os.MkdirAll(filepath.Join(source, "nested", "deep"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(source, "root.txt"), []byte("root"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(source, "nested", "file.txt"), []byte("nested"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(source, "nested", "deep", "leaf.txt"), []byte("leaf"), 0o644))

	target := filepath.Join(t.TempDir(), "dst")
	require.NoError(t, CopyDirectory(source, target))

	assertFileContent(t, filepath.Join(target, "root.txt"), "root")
	assertFileContent(t, filepath.Join(target, "nested", "file.txt"), "nested")
	assertFileContent(t, filepath.Join(target, "nested", "deep", "leaf.txt"), "leaf")
}

func TestCopyDirectoryOverwritesExistingTarget(t *testing.T) {
	source := filepath.Join(t.TempDir(), "src")
	require.NoError(t, os.MkdirAll(source, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(source, "SKILL.md"), []byte("new"), 0o644))

	target := filepath.Join(t.TempDir(), "dst")
	require.NoError(t, os.MkdirAll(target, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(target, "SKILL.md"), []byte("old"), 0o644))

	require.NoError(t, CopyDirectory(source, target))

	assertFileContent(t, filepath.Join(target, "SKILL.md"), "new")
}

func TestCopyDirectoryReadOnlySource(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix read-only directory semantics are not representable on windows")
	}

	source := filepath.Join(t.TempDir(), "src")
	require.NoError(t, os.MkdirAll(source, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(source, "SKILL.md"), []byte("content"), 0o644))
	require.NoError(t, os.Chmod(source, 0o555))

	target := filepath.Join(t.TempDir(), "dst")
	// Allow cleanup to remove the read-only trees.
	t.Cleanup(func() {
		_ = os.Chmod(source, 0o700)
		_ = os.Chmod(target, 0o700)
	})

	require.NoError(t, CopyDirectory(source, target))

	targetDir, err := os.Stat(target)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o555), targetDir.Mode().Perm())
	assertFileContent(t, filepath.Join(target, "SKILL.md"), "content")
}

func TestCopyDirectoryRejectsNonRegularEntry(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks are not reliable on windows")
	}

	source := filepath.Join(t.TempDir(), "src")
	require.NoError(t, os.MkdirAll(source, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(source, "real.txt"), []byte("x"), 0o644))
	require.NoError(t, os.Symlink(filepath.Join(source, "real.txt"), filepath.Join(source, "link")))

	err := CopyDirectory(source, filepath.Join(t.TempDir(), "dst"))
	require.Error(t, err)
	assert.ErrorContains(t, err, "unsupported file type")
}

func TestCopyDirectoryFailedCopyPreservesTarget(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks are not reliable on windows")
	}

	source := filepath.Join(t.TempDir(), "src")
	require.NoError(t, os.MkdirAll(source, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(source, "a.txt"), []byte("a"), 0o644))
	require.NoError(t, os.Symlink(filepath.Join(source, "a.txt"), filepath.Join(source, "link")))

	target := filepath.Join(t.TempDir(), "dst")
	require.NoError(t, os.MkdirAll(target, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(target, "SKILL.md"), []byte("old"), 0o644))

	err := CopyDirectory(source, target)
	require.Error(t, err)

	assertFileContent(t, filepath.Join(target, "SKILL.md"), "old")
}

func TestRemoveAllMissingPath(t *testing.T) {
	require.NoError(t, RemoveAll(filepath.Join(t.TempDir(), "missing")))
}

func TestRemoveAllReadOnlyTree(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix read-only directory semantics are not representable on windows")
	}

	tree := filepath.Join(t.TempDir(), "readonly")
	require.NoError(t, os.MkdirAll(filepath.Join(tree, "nested"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tree, "nested", "file.txt"), []byte("x"), 0o644))
	require.NoError(t, os.Chmod(filepath.Join(tree, "nested"), 0o555))
	require.NoError(t, os.Chmod(tree, 0o555))

	require.NoError(t, RemoveAll(tree))
	assert.NoDirExists(t, tree)
}

func TestCopyDirectoryMissingSource(t *testing.T) {
	err := CopyDirectory(filepath.Join(t.TempDir(), "missing"), filepath.Join(t.TempDir(), "dst"))
	require.Error(t, err)
}

func TestCopyFileRenameFailurePreservesTarget(t *testing.T) {
	// A directory already sitting at the target path makes the final rename fail.
	source := filepath.Join(t.TempDir(), "source.txt")
	require.NoError(t, os.WriteFile(source, []byte("new"), 0o644))

	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	require.NoError(t, os.MkdirAll(target, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(target, "keep.txt"), []byte("keep"), 0o644))

	err := copyFile(source, target, 0o644)
	require.Error(t, err)

	assertFileContent(t, filepath.Join(target, "keep.txt"), "keep")

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Equal(t, []string{"target"}, entryNames(entries), "no temp file should remain")
}

func TestCopyDirectoryPreservesPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix permission and setuid bits are not representable on windows")
	}

	// 0o777/0o666 are intentionally umask-stripped: only the explicit Chmod can
	// produce these exact modes on the copy.
	source := filepath.Join(t.TempDir(), "skill")
	require.NoError(t, os.MkdirAll(source, 0o777))
	require.NoError(t, os.Chmod(source, 0o777))

	require.NoError(t, os.MkdirAll(filepath.Join(source, "nested"), 0o777))
	require.NoError(t, os.Chmod(filepath.Join(source, "nested"), 0o777))

	filePath := filepath.Join(source, "SKILL.md")
	require.NoError(t, os.WriteFile(filePath, []byte("content"), 0o666))
	require.NoError(t, os.Chmod(filePath, 0o666|os.ModeSetuid))

	sourceFile, err := os.Stat(filePath)
	require.NoError(t, err)
	if sourceFile.Mode()&os.ModeSetuid == 0 {
		t.Skip("filesystem does not preserve the setuid bit; skipping special-bit assertion")
	}

	target := filepath.Join(t.TempDir(), "copied")
	require.NoError(t, CopyDirectory(source, target))

	targetDir, err := os.Stat(target)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o777), targetDir.Mode().Perm())

	targetNested, err := os.Stat(filepath.Join(target, "nested"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o777), targetNested.Mode().Perm())

	targetFile, err := os.Stat(filepath.Join(target, "SKILL.md"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o666), targetFile.Mode().Perm())
	assert.Zero(t, targetFile.Mode()&os.ModeSetuid, "setuid bit must not be copied")
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, want, string(content))
}

func entryNames(entries []os.DirEntry) []string {
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	return names
}
