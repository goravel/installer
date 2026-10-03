package commands

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/goravel/framework/contracts/process"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksprocess "github.com/goravel/framework/mocks/process"
	"github.com/goravel/framework/support/color"
	frameworkmock "github.com/goravel/framework/testing/mock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func (s *SkillInstallCommandTestSuite) TestHandleInstallAll() {
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(s.T().TempDir(), "skills")

	repoPath := expectAgentsClone(s.T(), mockProcess, map[string]string{
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
	s.NoDirExists(filepath.Dir(repoPath()))
}

func (s *SkillInstallCommandTestSuite) TestHandleInstallSelected() {
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(s.T().TempDir(), "skills")

	repoPath := expectAgentsClone(s.T(), mockProcess, map[string]string{
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
	s.NoDirExists(filepath.Dir(repoPath()))
}

func (s *SkillInstallCommandTestSuite) TestHandleMissingSkill() {
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(s.T().TempDir(), "skills")

	repoPath := expectAgentsClone(s.T(), mockProcess, map[string]string{
		"goravel-testing": "testing skill",
	})

	mockContext := newSkillInstallContext(s.T(), destination, []string{"missing-skill"}, false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillInstallCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, `skill "missing-skill" does not exist`)
	s.NoDirExists(destination)
	s.NoDirExists(filepath.Dir(repoPath()))
}

func (s *SkillInstallCommandTestSuite) TestHandleBlankSkillNameWithoutClone() {
	destination := filepath.Join(s.T().TempDir(), "skills")

	mockContext := newSkillInstallContext(s.T(), destination, []string{"   "}, false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillInstallCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "no valid skill names provided")
	s.NoDirExists(destination)
}

func (s *SkillInstallCommandTestSuite) TestHandlePartialInstallSkipsExisting() {
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(s.T().TempDir(), "skills")
	writeSkillContent(s.T(), destination, "goravel-testing", "old skill")

	repoPath := expectAgentsClone(s.T(), mockProcess, map[string]string{
		"goravel-testing":  "new skill",
		"goravel-planning": "planning skill",
	})

	mockContext := newSkillInstallContext(s.T(), destination, []string{"goravel-testing", "goravel-planning"}, false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillInstallCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "Installed 1 Goravel skill(s)")
	s.Contains(captureOutput, "Skipped 1 existing Goravel skill(s)")
	s.Equal("old skill", readSkillContent(s.T(), destination, "goravel-testing"))
	s.Equal("planning skill", readSkillContent(s.T(), destination, "goravel-planning"))
	s.NoDirExists(filepath.Dir(repoPath()))
}

func (s *SkillInstallCommandTestSuite) TestHandleMissingSkillsFolder() {
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(s.T().TempDir(), "skills")

	// The clone succeeds but creates nothing, so the skills folder is absent.
	repoPath := expectCloneWithoutRepo(s.T(), mockProcess)

	mockContext := newSkillInstallContext(s.T(), destination, nil, false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillInstallCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, noSkillsFoundMessage)
	s.NoDirExists(destination)
	s.NoDirExists(filepath.Dir(repoPath()))
}

func (s *SkillInstallCommandTestSuite) TestHandleEmptySkillsFolder() {
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(s.T().TempDir(), "skills")

	// The clone creates an empty skills folder.
	repoPath := expectAgentsClone(s.T(), mockProcess, map[string]string{})

	mockContext := newSkillInstallContext(s.T(), destination, nil, false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillInstallCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, noSkillsFoundMessage)
	s.NoDirExists(destination)
	s.NoDirExists(filepath.Dir(repoPath()))
}

func (s *SkillInstallCommandTestSuite) TestHandleSymlinkedSkillsRoot() {
	if runtime.GOOS == "windows" {
		s.T().Skip("symlinks are not reliable on windows")
	}

	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(s.T().TempDir(), "skills")
	outside := s.T().TempDir()

	var repoPath string
	mockProcess.EXPECT().Quietly().Return(mockProcess).Once()
	mockProcess.EXPECT().WithSpinner("Downloading Goravel agents").Return(mockProcess).Once()
	mockProcess.EXPECT().Timeout(cloneTimeout).Return(mockProcess).Once()
	mockProcessResult := mocksprocess.NewResult(s.T())
	mockProcessResult.EXPECT().Failed().Return(false).Once()
	mockProcess.EXPECT().Run("git", "clone", "--depth=1", agentsRepo, mock.Anything).RunAndReturn(func(name string, args ...string) process.Result {
		repoPath = args[3]
		require.NoError(s.T(), os.MkdirAll(repoPath, 0o755))
		require.NoError(s.T(), os.Symlink(outside, filepath.Join(repoPath, "skills")))

		return mockProcessResult
	}).Once()

	mockContext := newSkillInstallContext(s.T(), destination, nil, false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillInstallCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "is not a directory")
	s.NoDirExists(destination)
	s.NoDirExists(filepath.Dir(repoPath))
}

func (s *SkillInstallCommandTestSuite) TestHandleSkipExistingWithoutClone() {
	// Without --force, an already-installed skill is skipped without cloning.
	destination := filepath.Join(s.T().TempDir(), "skills")
	writeSkillContent(s.T(), destination, "goravel-testing", "old skill")

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
	repoPath := expectAgentsClone(s.T(), mockProcess, map[string]string{
		"goravel-testing": "new skill",
	})

	mockContext := newSkillInstallContext(s.T(), destination, []string{"goravel-testing"}, true)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillInstallCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "Installed 1 Goravel skill(s)")
	s.Equal("new skill", readSkillContent(s.T(), destination, "goravel-testing"))
	s.NoDirExists(filepath.Dir(repoPath()))
}

func (s *SkillInstallCommandTestSuite) TestHandleRemovesTempDir() {
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(s.T().TempDir(), "skills")

	repoPath := expectAgentsClone(s.T(), mockProcess, map[string]string{
		"goravel-testing": "testing skill",
	})

	mockContext := newSkillInstallContext(s.T(), destination, []string{"goravel-testing"}, false)
	_ = color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillInstallCommand.Handle(mockContext))
	})

	s.NotEmpty(repoPath())
	s.NoDirExists(filepath.Dir(repoPath()))
}

func (s *SkillInstallCommandTestSuite) TestHandleCloneFailure() {
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(s.T().TempDir(), "skills")
	cloneError := errors.New("clone failed")

	mockProcess.EXPECT().Quietly().Return(mockProcess).Once()
	mockProcess.EXPECT().WithSpinner("Downloading Goravel agents").Return(mockProcess).Once()
	mockProcess.EXPECT().Timeout(cloneTimeout).Return(mockProcess).Once()
	mockProcessResult := mocksprocess.NewResult(s.T())
	mockProcessResult.EXPECT().Failed().Return(true).Once()
	mockProcessResult.EXPECT().Error().Return(cloneError).Once()

	var repoPath string
	mockProcess.EXPECT().Run("git", "clone", "--depth=1", agentsRepo, mock.Anything).RunAndReturn(func(name string, args ...string) process.Result {
		repoPath = args[3]

		return mockProcessResult
	}).Once()

	mockContext := newSkillInstallContext(s.T(), destination, nil, false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillInstallCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "failed to clone goravel agents: clone failed")
	s.NoDirExists(destination)
	s.NotEmpty(repoPath)
	s.NoDirExists(filepath.Dir(repoPath))
}

func (s *SkillInstallCommandTestSuite) TestHandleDefaultEnvDestination() {
	mockProcess := frameworkmock.Factory().Process()
	envDir := s.T().TempDir()
	s.T().Setenv(envOpenCodeConfigDir, envDir)
	s.T().Setenv(envXDGConfigHome, "")

	repoPath := expectAgentsClone(s.T(), mockProcess, map[string]string{
		"goravel-testing": "testing skill",
	})

	mockContext := newSkillInstallContext(s.T(), "", nil, false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillInstallCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "Installed 1 Goravel skill(s)")
	s.Equal("testing skill", readSkillContent(s.T(), filepath.Join(envDir, "skills"), "goravel-testing"))
	s.NoDirExists(filepath.Dir(repoPath()))
}

func (s *SkillInstallCommandTestSuite) TestHandleRejectsUnsafePath() {
	home := s.T().TempDir()
	setHomeDir(s.T(), home)

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Option("path").Return("~").Once()

	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillInstallCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "refusing to use the home directory")
}

func newSkillInstallContext(t *testing.T, destination string, skills []string, force bool) *mocksconsole.Context {
	t.Helper()

	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Option("path").Return(destination).Once()
	mockContext.EXPECT().ArgumentStringSlice("skills").Return(skills).Once()
	mockContext.EXPECT().OptionBool("force").Return(force).Once()

	return mockContext
}
