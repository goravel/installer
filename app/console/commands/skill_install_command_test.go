package commands

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/goravel/framework/contracts/process"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksprocess "github.com/goravel/framework/mocks/process"
	"github.com/goravel/framework/support/color"
	frameworkmock "github.com/goravel/framework/testing/mock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type SkillInstallCommandTestSuite struct {
	suite.Suite
	skillInstallCommand *SkillInstallCommand
}

func TestSkillInstallCommandTestSuite(t *testing.T) {
	suite.Run(t, &SkillInstallCommandTestSuite{})
}

func (s *SkillInstallCommandTestSuite) SetupTest() {
	s.skillInstallCommand = NewSkillInstallCommand()
}

func (s *SkillInstallCommandTestSuite) TestGetDestinationDefaultPath() {
	home := s.T().TempDir()
	setHomeDir(s.T(), home)

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Option("path").Return("").Once()

	destination, err := s.skillInstallCommand.getDestination(mockContext)
	s.NoError(err)
	s.Equal(filepath.Join(home, ".agents", "skills"), destination)
}

func (s *SkillInstallCommandTestSuite) TestGetDestinationCustomHomePath() {
	home := s.T().TempDir()
	setHomeDir(s.T(), home)

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Option("path").Return("~/goravel-skills").Once()

	destination, err := s.skillInstallCommand.getDestination(mockContext)
	s.NoError(err)
	s.Equal(filepath.Join(home, "goravel-skills"), destination)
}

func (s *SkillInstallCommandTestSuite) TestGetDestinationCustomWindowsHomePath() {
	home := s.T().TempDir()
	setHomeDir(s.T(), home)

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Option("path").Return(`~\goravel-skills`).Once()

	destination, err := s.skillInstallCommand.getDestination(mockContext)
	s.NoError(err)
	s.Equal(filepath.Join(home, "goravel-skills"), destination)
}

func (s *SkillInstallCommandTestSuite) TestHandleInstallAll() {
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(s.T().TempDir(), "skills")

	expectAgentsClone(s.T(), mockProcess, map[string]string{
		"goravel-planning": "planning skill",
		"goravel-testing":  "testing skill",
	})

	mockContext := newSkillInstallContext(s.T(), destination, nil, false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillInstallCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "Installed 2 Goravel skill(s)")
	s.Equal("planning skill", readSkillContent(s.T(), destination, "goravel-planning"))
	s.Equal("testing skill", readSkillContent(s.T(), destination, "goravel-testing"))
}

func (s *SkillInstallCommandTestSuite) TestHandleInstallSelected() {
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(s.T().TempDir(), "skills")

	expectAgentsClone(s.T(), mockProcess, map[string]string{
		"goravel-planning": "planning skill",
		"goravel-testing":  "testing skill",
	})

	mockContext := newSkillInstallContext(s.T(), destination, []string{"goravel-testing"}, false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillInstallCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "Installed 1 Goravel skill(s)")
	s.Equal("testing skill", readSkillContent(s.T(), destination, "goravel-testing"))
	s.NoFileExists(filepath.Join(destination, "goravel-planning", "SKILL.md"))
}

func (s *SkillInstallCommandTestSuite) TestHandleMissingSkill() {
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(s.T().TempDir(), "skills")

	expectAgentsClone(s.T(), mockProcess, map[string]string{
		"goravel-testing": "testing skill",
	})

	mockContext := newSkillInstallContext(s.T(), destination, []string{"missing-skill"}, false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillInstallCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, `skill "missing-skill" does not exist`)
	s.NoDirExists(destination)
}

func (s *SkillInstallCommandTestSuite) TestHandleSkipExisting() {
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(s.T().TempDir(), "skills")

	writeSkillContent(s.T(), destination, "goravel-testing", "old skill")
	expectAgentsClone(s.T(), mockProcess, map[string]string{
		"goravel-testing": "new skill",
	})

	mockContext := newSkillInstallContext(s.T(), destination, []string{"goravel-testing"}, false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillInstallCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "Skipped 1 existing Goravel skill(s)")
	s.Equal("old skill", readSkillContent(s.T(), destination, "goravel-testing"))
}

func (s *SkillInstallCommandTestSuite) TestHandleForceExisting() {
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(s.T().TempDir(), "skills")

	writeSkillContent(s.T(), destination, "goravel-testing", "old skill")
	expectAgentsClone(s.T(), mockProcess, map[string]string{
		"goravel-testing": "new skill",
	})

	mockContext := newSkillInstallContext(s.T(), destination, []string{"goravel-testing"}, true)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillInstallCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "Installed 1 Goravel skill(s)")
	s.Equal("new skill", readSkillContent(s.T(), destination, "goravel-testing"))
}

func (s *SkillInstallCommandTestSuite) TestHandleCloneFailure() {
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(s.T().TempDir(), "skills")
	cloneError := errors.New("clone failed")

	mockProcess.EXPECT().Quietly().Return(mockProcess).Once()
	mockProcess.EXPECT().WithSpinner("Downloading Goravel agents").Return(mockProcess).Once()
	mockProcessResult := mocksprocess.NewResult(s.T())
	mockProcessResult.EXPECT().Failed().Return(true).Once()
	mockProcessResult.EXPECT().Error().Return(cloneError).Once()
	mockProcess.EXPECT().Run("git", "clone", "--depth=1", agentsRepo, mock.Anything).Return(mockProcessResult).Once()

	mockContext := newSkillInstallContext(s.T(), destination, nil, false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillInstallCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "failed to clone goravel agents: clone failed")
	s.NoDirExists(destination)
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

	mockProcess.EXPECT().Quietly().Return(mockProcess).Once()
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
	if err := os.MkdirAll(skillsPath, 0755); err != nil {
		t.Fatalf("os.MkdirAll(%q) = %v, want nil", skillsPath, err)
	}
	for skill, content := range skills {
		writeSkillContent(t, skillsPath, skill, content)
	}
}

func readSkillContent(t *testing.T, skillsPath, skill string) string {
	t.Helper()

	path := filepath.Join(skillsPath, skill, "SKILL.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) = %v, want nil", path, err)
	}

	return string(content)
}

func writeSkillContent(t *testing.T, skillsPath, skill, content string) {
	t.Helper()

	skillPath := filepath.Join(skillsPath, skill)
	if err := os.MkdirAll(skillPath, 0755); err != nil {
		t.Fatalf("os.MkdirAll(%q) = %v, want nil", skillPath, err)
	}
	path := filepath.Join(skillPath, "SKILL.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("os.WriteFile(%q) = %v, want nil", path, err)
	}
}
