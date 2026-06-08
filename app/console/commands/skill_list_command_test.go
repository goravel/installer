package commands

import (
	"errors"
	"io"
	"testing"

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

	mockContext := mocksconsole.NewContext(s.T())
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillListCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "Available Goravel skills:")
	s.Contains(captureOutput, "  - goravel-planning")
	s.Contains(captureOutput, "  - goravel-testing")
}

func (s *SkillListCommandTestSuite) TestHandleNoSkills() {
	mockProcess := frameworkmock.Factory().Process()
	expectAgentsClone(s.T(), mockProcess, map[string]string{})

	mockContext := mocksconsole.NewContext(s.T())
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillListCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "no skills found in goravel/agents")
}

func (s *SkillListCommandTestSuite) TestHandleCloneFailure() {
	mockProcess := frameworkmock.Factory().Process()
	cloneError := errors.New("clone failed")

	mockProcess.EXPECT().Quietly().Return(mockProcess).Once()
	mockProcess.EXPECT().WithSpinner("Downloading Goravel agents").Return(mockProcess).Once()
	mockProcessResult := mocksprocess.NewResult(s.T())
	mockProcessResult.EXPECT().Failed().Return(true).Once()
	mockProcessResult.EXPECT().Error().Return(cloneError).Once()
	mockProcess.EXPECT().Run("git", "clone", "--depth=1", agentsRepo, mock.Anything).Return(mockProcessResult).Once()

	mockContext := mocksconsole.NewContext(s.T())
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.skillListCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "failed to clone goravel agents: clone failed")
}
