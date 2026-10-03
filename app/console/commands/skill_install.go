package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/support/file"

	"github.com/goravel/installer/internal/fsutil"
)

// installSkills installs the requested skills (or all of them when none are
// named) into destination and returns the number of skills installed and the
// number skipped because they already existed without --force.
func installSkills(destination string, skillNames []string, force bool) (installed, skipped int, err error) {
	normalized, err := normalizeSkillNames(skillNames)
	if err != nil {
		return 0, 0, err
	}
	// Fail before cloning when the caller supplied only blank/invalid names.
	if len(normalized) == 0 && len(skillNames) > 0 {
		return 0, 0, errors.New("no valid skill names provided")
	}

	// Without --force, skip the network clone entirely when every requested
	// skill is already installed. The install-all path still clones.
	if !force && len(normalized) > 0 {
		allInstalled := true
		for _, skill := range normalized {
			if !skillInstalled(destination, skill) {
				allInstalled = false

				break
			}
		}
		if allInstalled {
			return 0, len(normalized), nil
		}
	}

	err = withAgentsRepo(func(skillsPath string) error {
		skills, err := resolveSkills(skillsPath, normalized)
		if err != nil {
			return err
		}
		if len(skills) == 0 {
			return errNoSkillsFound
		}

		if err := os.MkdirAll(destination, 0o755); err != nil {
			return fmt.Errorf("failed to create skills directory: %w", err)
		}

		for _, skill := range skills {
			wasInstalled, err := installSkill(skillsPath, destination, skill, force, os.Rename)
			if err != nil {
				return err
			}
			if wasInstalled {
				installed++
			} else {
				skipped++
			}
		}

		return nil
	})
	if err != nil {
		return installed, skipped, err
	}

	return installed, skipped, nil
}

// skillInstalled reports whether skill is already present at destination.
func skillInstalled(destination, skill string) bool {
	return file.Exists(filepath.Join(destination, skill))
}

// resolveSkills validates that each supplied (already-normalized) name is a real
// directory, rejecting symlinks and files. An empty slice lists every skill in
// the repo. Callers normalize names and reject blank-only input before cloning.
func resolveSkills(skillsPath string, skillNames []string) ([]string, error) {
	if len(skillNames) == 0 {
		return listSkills(skillsPath)
	}

	for _, skill := range skillNames {
		info, err := os.Lstat(filepath.Join(skillsPath, skill))
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("skill %q does not exist", skill)
			}

			return nil, fmt.Errorf("failed to inspect skill %q: %w", skill, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("skill %q is not a directory", skill)
		}
	}

	return skillNames, nil
}

// installSkill installs one skill atomically. It copies the source into a unique
// staging directory inside destination, moves any existing target aside to a
// unique backup directory, renames staging into place, and restores the backup
// if the rename fails. A failed copy therefore leaves the previous install
// untouched and removes the staging directory. The rename operation is injected
// so tests can force a failure; callers normally pass os.Rename.
func installSkill(skillsPath, destination, skill string, force bool, rename func(oldpath, newpath string) error) (bool, error) {
	source := filepath.Join(skillsPath, skill)
	target := filepath.Join(destination, skill)

	targetExists := skillInstalled(destination, skill)
	if targetExists && !force {
		return false, nil
	}

	// Dot-prefixed staging/backup names keep any leftovers from being listed as
	// skills.
	staging, err := os.MkdirTemp(destination, "."+skill+"-staging-*")
	if err != nil {
		return false, fmt.Errorf("failed to create staging directory for skill %q: %w", skill, err)
	}

	if err := fsutil.CopyDirectory(source, staging); err != nil {
		if rmErr := fsutil.RemoveAll(staging); rmErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to remove staging directory: %w", rmErr))
		}

		return false, fmt.Errorf("failed to install skill %q: %w", skill, err)
	}

	backup := ""
	if targetExists {
		backup, err = os.MkdirTemp(destination, "."+skill+"-backup-*")
		if err != nil {
			_ = fsutil.RemoveAll(staging)

			return false, fmt.Errorf("failed to prepare backup for skill %q: %w", skill, err)
		}
		if err := fsutil.RemoveAll(backup); err != nil {
			_ = fsutil.RemoveAll(staging)

			return false, fmt.Errorf("failed to prepare backup for skill %q: %w", skill, err)
		}
		if err := rename(target, backup); err != nil {
			_ = fsutil.RemoveAll(staging)

			return false, fmt.Errorf("failed to replace existing skill %q: %w", skill, err)
		}
	}

	if err := rename(staging, target); err != nil {
		if backup != "" {
			if restoreErr := rename(backup, target); restoreErr != nil {
				err = errors.Join(err, fmt.Errorf("failed to restore previous skill %q: %w", skill, restoreErr))
			}
		}
		_ = fsutil.RemoveAll(staging)

		return false, fmt.Errorf("failed to install skill %q: %w", skill, err)
	}

	// The install succeeded; removing the backup is best-effort so a cleanup
	// failure cannot turn a successful install into an error.
	_ = fsutil.RemoveAll(backup)

	return true, nil
}

// normalizeSkillNames trims the supplied names, drops blanks, removes duplicates,
// and rejects ".", "..", or names containing path separators. The result is empty
// when no names survive, which callers use to distinguish "install all" from
// "the user supplied only invalid names".
func normalizeSkillNames(skillNames []string) ([]string, error) {
	seen := make(map[string]struct{}, len(skillNames))
	unique := make([]string, 0, len(skillNames))

	for _, skill := range skillNames {
		skill = strings.TrimSpace(skill)
		if skill == "" {
			continue
		}
		if skill == "." || skill == ".." || strings.ContainsAny(skill, `/\\`) {
			return nil, fmt.Errorf("invalid skill name %q", skill)
		}
		if _, ok := seen[skill]; ok {
			continue
		}

		seen[skill] = struct{}{}
		unique = append(unique, skill)
	}

	return unique, nil
}
