package commands

import (
	"errors"
	"io"
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

type SkillListCommandTestSuite struct {
	suite.Suite
	skillListCommand *SkillListCommand
}

func TestSkillListCommandTestSuite(t *testing.T) {
	suite.Run(t, &SkillListCommandTestSuite{})
}

func (s *SkillListCommandTestSuite) SetupTest() {
	s.skillListCommand = NewSkillListCommand()
}

func (s *SkillListCommandTestSuite) TestHandleListSkills() {
	mockProcess := frameworkmock.Factory().Process()
	expectAgentsClone(s.T(), mockProcess, map[string]string{
		"goravel-planning": "planning skill",
		"goravel-testing":  "testing skill",
	})

	mockContext := newSkillListContext(s.T(), false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillListCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "Available Goravel skills:")
	s.Contains(captureOutput, "1. goravel-planning")
	s.Contains(captureOutput, "2. goravel-testing")
	s.NotContains(captureOutput, "planning skill")
}

func (s *SkillListCommandTestSuite) TestHandleListSkillDetails() {
	mockProcess := frameworkmock.Factory().Process()
	expectAgentsClone(s.T(), mockProcess, map[string]string{
		"goravel-testing": "---\nname: goravel-testing\ndescription: >\n  Goravel test-writing and test-running conventions.\n  Use this skill when adding tests.\n---\n\n# Testing",
	})

	mockContext := newSkillListContext(s.T(), true)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillListCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "1. goravel-testing")
	s.Contains(captureOutput, "   Description: Goravel test-writing and test-running conventions. Use this skill when adding tests.")
}

func (s *SkillListCommandTestSuite) TestHandleListSingleLineDescription() {
	mockProcess := frameworkmock.Factory().Process()
	expectAgentsClone(s.T(), mockProcess, map[string]string{
		"goravel-testing": "---\nname: goravel-testing\ndescription: Goravel testing skill\n---\n\n# Testing",
	})

	mockContext := newSkillListContext(s.T(), true)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillListCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "   Description: Goravel testing skill")
}

func (s *SkillListCommandTestSuite) TestHandleListQuotedSingleLineDescription() {
	mockProcess := frameworkmock.Factory().Process()
	expectAgentsClone(s.T(), mockProcess, map[string]string{
		"goravel-testing": "---\nname: goravel-testing\ndescription: \"Goravel quoted skill\"\n---\n\n# Testing",
	})

	mockContext := newSkillListContext(s.T(), true)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillListCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "   Description: Goravel quoted skill")
}

func (s *SkillListCommandTestSuite) TestHandleListSkillWithoutDescription() {
	mockProcess := frameworkmock.Factory().Process()
	expectAgentsClone(s.T(), mockProcess, map[string]string{
		"goravel-testing": "---\nname: goravel-testing\n---\n\n# Testing",
	})

	mockContext := newSkillListContext(s.T(), true)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillListCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "1. goravel-testing")
	s.NotContains(captureOutput, "Description:")
}

func (s *SkillListCommandTestSuite) TestHandleNoSkills() {
	mockProcess := frameworkmock.Factory().Process()
	repoPath := expectAgentsClone(s.T(), mockProcess, map[string]string{})

	mockContext := newSkillListContext(s.T(), false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillListCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, noSkillsFoundMessage)
	s.NoDirExists(filepath.Dir(repoPath()))
}

func (s *SkillListCommandTestSuite) TestHandleRemovesTempDir() {
	mockProcess := frameworkmock.Factory().Process()
	repoPath := expectAgentsClone(s.T(), mockProcess, map[string]string{
		"goravel-testing": "testing skill",
	})

	mockContext := newSkillListContext(s.T(), false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillListCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "goravel-testing")
	s.NotEmpty(repoPath())
	s.NoDirExists(filepath.Dir(repoPath()))
}

func (s *SkillListCommandTestSuite) TestHandleCloneFailure() {
	mockProcess := frameworkmock.Factory().Process()
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

	mockContext := newSkillListContext(s.T(), false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillListCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "failed to clone goravel agents: clone failed")
	s.NotEmpty(repoPath)
	s.NoDirExists(filepath.Dir(repoPath))
}

func newSkillListContext(t *testing.T, detail bool) *mocksconsole.Context {
	t.Helper()

	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().OptionBool("detail").Return(detail).Once()

	return mockContext
}
