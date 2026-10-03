// Package fsutil provides filesystem helpers for the installer, including a
// recursive copy that preserves permission bits and a removal that can delete
// read-only trees.
package fsutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyDirectory recursively copies the tree at source into target. Regular
// read/write/execute permission bits are preserved exactly (special setuid,
// setgid, and sticky bits are dropped). Directories are created writable during
// the walk and their exact permission bits are applied in a second pass after
// all children are written, so a read-only source directory (for example 0o555)
// does not block the copy.
func CopyDirectory(source, target string) error {
	type dirPerm struct {
		path string
		perm os.FileMode
	}

	var dirs []dirPerm

	err := filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(target, relativePath)

		if entry.IsDir() {
			if err := makeDir(targetPath); err != nil {
				return err
			}

			dirs = append(dirs, dirPerm{path: targetPath, perm: info.Mode().Perm()})

			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported file type %q", path)
		}

		return copyFile(path, targetPath, info.Mode().Perm())
	})
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		if err := os.Chmod(dir.path, dir.perm); err != nil {
			return err
		}
	}

	return nil
}

// RemoveAll removes path and any children, restoring owner write/execute
// permission on directories in the tree first so a tree copied with read-only
// permission bits (for example 0o555) can still be deleted. A missing path is
// not an error.
func RemoveAll(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	if info.IsDir() {
		_ = filepath.WalkDir(path, func(p string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return nil
			}
			if !entry.IsDir() {
				return nil
			}

			if dirInfo, infoErr := entry.Info(); infoErr == nil {
				_ = os.Chmod(p, dirInfo.Mode().Perm()|0o700)
			} else {
				_ = os.Chmod(p, 0o700)
			}

			return nil
		})
	}

	return os.RemoveAll(path)
}

// makeDir creates path with a writable mode so children can be copied into it
// before the exact source permission bits are applied.
func makeDir(path string) error {
	return os.MkdirAll(path, 0o700)
}

// copyFile streams source into an exclusive temporary file in the target
// directory and renames it over target only after the copy succeeds, so an
// existing target is never truncated. The final permission bits are forced with
// Chmod to bypass the umask.
func copyFile(source, target string, perm os.FileMode) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	tmpFile, err := os.CreateTemp(filepath.Dir(target), ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	if _, err := io.Copy(tmpFile, sourceFile); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)

		return err
	}
	if err := tmpFile.Chmod(perm); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)

		return err
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)

		return err
	}
	if err := os.Rename(tmpPath, target); err != nil {
		_ = os.Remove(tmpPath)

		return err
	}

	return nil
}
