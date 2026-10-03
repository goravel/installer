package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/goravel/installer/app/facades"
)

const (
	agentsRepo   = "https://github.com/goravel/agents.git"
	cloneTimeout = 2 * time.Minute
)

// cloneAgents shallow-clones the default branch of goravel/agents into path,
// showing a download spinner and failing after cloneTimeout.
func cloneAgents(path string) error {
	res := facades.Process().Quietly().WithSpinner("Downloading Goravel agents").
		Timeout(cloneTimeout).
		Run("git", "clone", "--depth=1", agentsRepo, path)
	if res.Failed() {
		return fmt.Errorf("failed to clone goravel agents: %w", res.Error())
	}

	return nil
}

// withAgentsRepo clones goravel/agents into a temporary directory, invokes fn
// with the repo's skills path, and always removes the temp directory afterwards.
// The skills path is rejected with an error unless it is a real directory, so a
// symlinked skills root cannot redirect install/list outside the cloned repo.
func withAgentsRepo(fn func(skillsPath string) error) error {
	tmpDir, err := os.MkdirTemp("", "goravel-agents-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	repoPath := filepath.Join(tmpDir, "agents")
	if err := cloneAgents(repoPath); err != nil {
		return err
	}

	skillsPath := filepath.Join(repoPath, "skills")
	info, err := os.Lstat(skillsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errNoSkillsFound
		}

		return fmt.Errorf("failed to inspect skills folder: %w", err)
	}
	if !info.IsDir() {
		return errors.New("skills path in goravel/agents is not a directory")
	}

	return fn(skillsPath)
}
