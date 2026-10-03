//go:build !windows

package fsutil

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCopyDirectoryBypassesUmask proves the explicit Chmod is required: with a
// restrictive umask, only the explicit chmod can produce the source modes.
func TestCopyDirectoryBypassesUmask(t *testing.T) {
	old := syscall.Umask(0o077)
	defer syscall.Umask(old)

	source := filepath.Join(t.TempDir(), "skill")
	require.NoError(t, os.MkdirAll(source, 0o777))
	require.NoError(t, os.Chmod(source, 0o777))

	filePath := filepath.Join(source, "SKILL.md")
	require.NoError(t, os.WriteFile(filePath, []byte("content"), 0o666))
	require.NoError(t, os.Chmod(filePath, 0o666))

	target := filepath.Join(t.TempDir(), "copied")
	require.NoError(t, CopyDirectory(source, target))

	dirInfo, err := os.Stat(target)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o777), dirInfo.Mode().Perm())

	fileInfo, err := os.Stat(filepath.Join(target, "SKILL.md"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o666), fileInfo.Mode().Perm())
}
