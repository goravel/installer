package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goravel/framework/contracts/process"
	mocksprocess "github.com/goravel/framework/mocks/process"
	"github.com/stretchr/testify/mock"
)

func setHomeDir(t *testing.T, home string) {
	t.Helper()

	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
}

// canonicalTempDir returns a temp dir with symlinks resolved, matching the
// canonical path returned by resolveDestination.
func canonicalTempDir(t *testing.T) string {
	t.Helper()

	resolved, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("filepath.EvalSymlinks() = %v, want nil", err)
	}

	return resolved
}

// expectAgentsClone sets up a successful clone that creates the given skills in
// the repo and returns a getter for the repo path, so callers can assert the
// temp directory is removed after the command runs.
func expectAgentsClone(t *testing.T, mockProcess *mocksprocess.Process, skills map[string]string) func() string {
	t.Helper()

	var repoPath string

	mockProcess.EXPECT().Quietly().Return(mockProcess).Once()
	mockProcess.EXPECT().WithSpinner("Downloading Goravel agents").Return(mockProcess).Once()
	mockProcess.EXPECT().Timeout(cloneTimeout).Return(mockProcess).Once()
	mockProcessResult := mocksprocess.NewResult(t)
	mockProcessResult.EXPECT().Failed().Return(false).Once()
	mockProcess.EXPECT().Run("git", "clone", "--depth=1", agentsRepo, mock.Anything).RunAndReturn(func(name string, args ...string) process.Result {
		repoPath = args[3]
		createAgentsRepo(t, repoPath, skills)

		return mockProcessResult
	}).Once()

	return func() string { return repoPath }
}

// expectCloneWithoutRepo sets up a successful clone that creates no repo
// contents and returns a getter for the repo path.
func expectCloneWithoutRepo(t *testing.T, mockProcess *mocksprocess.Process) func() string {
	t.Helper()

	var repoPath string

	mockProcess.EXPECT().Quietly().Return(mockProcess).Once()
	mockProcess.EXPECT().WithSpinner("Downloading Goravel agents").Return(mockProcess).Once()
	mockProcess.EXPECT().Timeout(cloneTimeout).Return(mockProcess).Once()
	mockProcessResult := mocksprocess.NewResult(t)
	mockProcessResult.EXPECT().Failed().Return(false).Once()
	mockProcess.EXPECT().Run("git", "clone", "--depth=1", agentsRepo, mock.Anything).RunAndReturn(func(name string, args ...string) process.Result {
		repoPath = args[3]

		return mockProcessResult
	}).Once()

	return func() string { return repoPath }
}

func entryNames(entries []os.DirEntry) []string {
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	return names
}

func createAgentsRepo(t *testing.T, path string, skills map[string]string) {
	t.Helper()

	skillsPath := filepath.Join(path, "skills")
	if err := os.MkdirAll(skillsPath, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q) = %v, want nil", skillsPath, err)
	}
	for skill, content := range skills {
		writeSkillContent(t, skillsPath, skill, content)
	}
}

func readSkillContent(t *testing.T, skillsPath, skill string) string {
	t.Helper()

	return readFile(t, filepath.Join(skillsPath, skill, "SKILL.md"))
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) = %v, want nil", path, err)
	}

	return string(content)
}

func writeSkillContent(t *testing.T, skillsPath, skill, content string) {
	t.Helper()

	skillPath := filepath.Join(skillsPath, skill)
	if err := os.MkdirAll(skillPath, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q) = %v, want nil", skillPath, err)
	}
	path := filepath.Join(skillPath, "SKILL.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) = %v, want nil", path, err)
	}
}
