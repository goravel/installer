package commands

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/goravel/framework/contracts/process"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksprocess "github.com/goravel/framework/mocks/process"
	"github.com/goravel/framework/support/color"
	frameworkmock "github.com/goravel/framework/testing/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSkillInstallCommandGetDestination(t *testing.T) {
	skillInstallCommand := NewSkillInstallCommand()

	t.Run("default path", func(t *testing.T) {
		home := t.TempDir()
		setHomeDir(t, home)

		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Option("path").Return("").Once()

		destination, err := skillInstallCommand.getDestination(mockContext)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(home, ".agents", "skills"), destination)
	})

	t.Run("custom home path", func(t *testing.T) {
		home := t.TempDir()
		setHomeDir(t, home)

		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Option("path").Return("~/goravel-skills").Once()

		destination, err := skillInstallCommand.getDestination(mockContext)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(home, "goravel-skills"), destination)
	})
}

func TestSkillInstallCommandHandleInstallAll(t *testing.T) {
	skillInstallCommand := NewSkillInstallCommand()
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(t.TempDir(), "skills")

	expectAgentsClone(t, mockProcess, map[string]string{
		"goravel-planning": "planning skill",
		"goravel-testing":  "testing skill",
	})

	mockContext := newSkillInstallContext(t, destination, nil, false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		assert.NoError(t, skillInstallCommand.Handle(mockContext))
	})

	assert.Contains(t, captureOutput, "Installed 2 Goravel skill(s)")
	assert.Equal(t, "planning skill", readSkillContent(t, destination, "goravel-planning"))
	assert.Equal(t, "testing skill", readSkillContent(t, destination, "goravel-testing"))
}

func TestSkillInstallCommandHandleInstallSelected(t *testing.T) {
	skillInstallCommand := NewSkillInstallCommand()
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(t.TempDir(), "skills")

	expectAgentsClone(t, mockProcess, map[string]string{
		"goravel-planning": "planning skill",
		"goravel-testing":  "testing skill",
	})

	mockContext := newSkillInstallContext(t, destination, []string{"goravel-testing"}, false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		assert.NoError(t, skillInstallCommand.Handle(mockContext))
	})

	assert.Contains(t, captureOutput, "Installed 1 Goravel skill(s)")
	assert.Equal(t, "testing skill", readSkillContent(t, destination, "goravel-testing"))
	assert.NoFileExists(t, filepath.Join(destination, "goravel-planning", "SKILL.md"))
}

func TestSkillInstallCommandHandleMissingSkill(t *testing.T) {
	skillInstallCommand := NewSkillInstallCommand()
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(t.TempDir(), "skills")

	expectAgentsClone(t, mockProcess, map[string]string{
		"goravel-testing": "testing skill",
	})

	mockContext := newSkillInstallContext(t, destination, []string{"missing-skill"}, false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		assert.NoError(t, skillInstallCommand.Handle(mockContext))
	})

	assert.Contains(t, captureOutput, `skill "missing-skill" does not exist`)
	assert.NoDirExists(t, destination)
}

func TestSkillInstallCommandHandleSkipExisting(t *testing.T) {
	skillInstallCommand := NewSkillInstallCommand()
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(t.TempDir(), "skills")

	writeSkillContent(t, destination, "goravel-testing", "old skill")
	expectAgentsClone(t, mockProcess, map[string]string{
		"goravel-testing": "new skill",
	})

	mockContext := newSkillInstallContext(t, destination, []string{"goravel-testing"}, false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		assert.NoError(t, skillInstallCommand.Handle(mockContext))
	})

	assert.Contains(t, captureOutput, "Skipped 1 existing Goravel skill(s)")
	assert.Equal(t, "old skill", readSkillContent(t, destination, "goravel-testing"))
}

func TestSkillInstallCommandHandleForceExisting(t *testing.T) {
	skillInstallCommand := NewSkillInstallCommand()
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(t.TempDir(), "skills")

	writeSkillContent(t, destination, "goravel-testing", "old skill")
	expectAgentsClone(t, mockProcess, map[string]string{
		"goravel-testing": "new skill",
	})

	mockContext := newSkillInstallContext(t, destination, []string{"goravel-testing"}, true)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		assert.NoError(t, skillInstallCommand.Handle(mockContext))
	})

	assert.Contains(t, captureOutput, "Installed 1 Goravel skill(s)")
	assert.Equal(t, "new skill", readSkillContent(t, destination, "goravel-testing"))
}

func TestSkillInstallCommandHandleCloneFailure(t *testing.T) {
	skillInstallCommand := NewSkillInstallCommand()
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(t.TempDir(), "skills")

	mockProcess.EXPECT().WithSpinner("Downloading Goravel agents").Return(mockProcess).Once()
	mockProcessResult := mocksprocess.NewResult(t)
	mockProcessResult.EXPECT().Failed().Return(true).Once()
	mockProcessResult.EXPECT().Error().Return(assert.AnError).Once()
	mockProcess.EXPECT().Run("git", "clone", "--depth=1", agentsRepo, mock.Anything).Return(mockProcessResult).Once()

	mockContext := newSkillInstallContext(t, destination, nil, false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		assert.NoError(t, skillInstallCommand.Handle(mockContext))
	})

	assert.Contains(t, captureOutput, "failed to clone goravel agents")
	assert.NoDirExists(t, destination)
}

func newSkillInstallContext(t *testing.T, destination string, skills []string, force bool) *mocksconsole.Context {
	t.Helper()

	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Option("path").Return(destination).Once()
	mockContext.EXPECT().ArgumentStringSlice("skills").Return(skills).Once()
	mockContext.EXPECT().OptionBool("force").Return(force).Once()

	return mockContext
}

func setHomeDir(t *testing.T, home string) {
	t.Helper()

	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
}

func expectAgentsClone(t *testing.T, mockProcess *mocksprocess.Process, skills map[string]string) {
	t.Helper()

	mockProcess.EXPECT().WithSpinner("Downloading Goravel agents").Return(mockProcess).Once()
	mockProcessResult := mocksprocess.NewResult(t)
	mockProcessResult.EXPECT().Failed().Return(false).Once()
	mockProcess.EXPECT().Run("git", "clone", "--depth=1", agentsRepo, mock.Anything).RunAndReturn(func(name string, args ...string) process.Result {
		createAgentsRepo(t, args[3], skills)

		return mockProcessResult
	}).Once()
}

func createAgentsRepo(t *testing.T, path string, skills map[string]string) {
	t.Helper()

	skillsPath := filepath.Join(path, "skills")
	assert.NoError(t, os.MkdirAll(skillsPath, 0755))
	for skill, content := range skills {
		writeSkillContent(t, skillsPath, skill, content)
	}
}

func readSkillContent(t *testing.T, skillsPath, skill string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(skillsPath, skill, "SKILL.md"))
	assert.NoError(t, err)

	return string(content)
}

func writeSkillContent(t *testing.T, skillsPath, skill, content string) {
	t.Helper()

	skillPath := filepath.Join(skillsPath, skill)
	assert.NoError(t, os.MkdirAll(skillPath, 0755))
	assert.NoError(t, os.WriteFile(filepath.Join(skillPath, "SKILL.md"), []byte(content), 0644))
}
