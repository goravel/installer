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

type AgentInstallCommandTestSuite struct {
	suite.Suite
	agentInstallCommand *AgentInstallCommand
}

func TestAgentInstallCommandTestSuite(t *testing.T) {
	suite.Run(t, &AgentInstallCommandTestSuite{})
}

func (s *AgentInstallCommandTestSuite) SetupTest() {
	s.agentInstallCommand = NewAgentInstallCommand()
}

func (s *AgentInstallCommandTestSuite) TestGetDestinationDefaultPath() {
	home := s.T().TempDir()
	setHomeDir(s.T(), home)
	s.T().Setenv(opencodeConfigEnv, "")
	s.T().Setenv(xdgConfigEnv, "")

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Option("path").Return("").Once()

	destination, err := s.agentInstallCommand.getDestination(mockContext)
	s.NoError(err)
	s.Equal(filepath.Join(home, ".config", "opencode", "agents"), destination)
}

func (s *AgentInstallCommandTestSuite) TestGetDestinationOpencodeConfigDir() {
	configDir := s.T().TempDir()
	s.T().Setenv(opencodeConfigEnv, configDir)
	s.T().Setenv(xdgConfigEnv, s.T().TempDir())

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Option("path").Return("").Once()

	destination, err := s.agentInstallCommand.getDestination(mockContext)
	s.NoError(err)
	s.Equal(filepath.Join(configDir, "agents"), destination)
}

func (s *AgentInstallCommandTestSuite) TestGetDestinationXDGConfigHome() {
	configHome := s.T().TempDir()
	s.T().Setenv(opencodeConfigEnv, "")
	s.T().Setenv(xdgConfigEnv, configHome)

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Option("path").Return("").Once()

	destination, err := s.agentInstallCommand.getDestination(mockContext)
	s.NoError(err)
	s.Equal(filepath.Join(configHome, "opencode", "agents"), destination)
}

func (s *AgentInstallCommandTestSuite) TestGetDestinationCustomHomePath() {
	home := s.T().TempDir()
	setHomeDir(s.T(), home)

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Option("path").Return("~/goravel-agents").Once()

	destination, err := s.agentInstallCommand.getDestination(mockContext)
	s.NoError(err)
	s.Equal(filepath.Join(home, "goravel-agents"), destination)
}

func (s *AgentInstallCommandTestSuite) TestGetDestinationCustomWindowsHomePath() {
	home := s.T().TempDir()
	setHomeDir(s.T(), home)

	mockContext := mocksconsole.NewContext(s.T())
	mockContext.EXPECT().Option("path").Return(`~\goravel-agents`).Once()

	destination, err := s.agentInstallCommand.getDestination(mockContext)
	s.NoError(err)
	s.Equal(filepath.Join(home, "goravel-agents"), destination)
}

func (s *AgentInstallCommandTestSuite) TestHandleInstallAll() {
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(s.T().TempDir(), "agents")

	expectAgentsCloneForAgents(s.T(), mockProcess, map[string]string{
		"code.md":             "top-level agent",
		"subagents/code.md":   "subagent",
		"subagents/go.md":     "go subagent",
		"subagents/notes.txt": "ignored",
	})

	mockContext := newAgentInstallContext(s.T(), destination, false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.agentInstallCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "Installed 1 Goravel agent(s)")
	s.Equal("top-level agent", readAgentContent(s.T(), destination, "code.md"))
	s.Equal("subagent", readAgentContent(s.T(), destination, "subagents/code.md"))
	s.Equal("go subagent", readAgentContent(s.T(), destination, "subagents/go.md"))
	s.NoFileExists(filepath.Join(destination, "subagents", "notes.txt"))
}

func (s *AgentInstallCommandTestSuite) TestHandleEmptyAgentsFolder() {
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(s.T().TempDir(), "agents")

	expectAgentsCloneForAgents(s.T(), mockProcess, map[string]string{})

	mockContext := newAgentInstallContext(s.T(), destination, false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.agentInstallCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "no agents found in goravel/agents")
	s.NoDirExists(destination)
}

func (s *AgentInstallCommandTestSuite) TestHandleSkipExisting() {
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(s.T().TempDir(), "agents")

	writeAgentContent(s.T(), destination, "code.md", "old agent")
	expectAgentsCloneForAgents(s.T(), mockProcess, map[string]string{
		"code.md": "new agent",
	})

	mockContext := newAgentInstallContext(s.T(), destination, false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.agentInstallCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "Skipped 1 existing Goravel agent(s)")
	s.Equal("old agent", readAgentContent(s.T(), destination, "code.md"))
}

func (s *AgentInstallCommandTestSuite) TestHandleForceExisting() {
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(s.T().TempDir(), "agents")

	writeAgentContent(s.T(), destination, "code.md", "old agent")
	expectAgentsCloneForAgents(s.T(), mockProcess, map[string]string{
		"code.md": "new agent",
	})

	mockContext := newAgentInstallContext(s.T(), destination, true)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.agentInstallCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "Installed 1 Goravel agent(s)")
	s.Equal("new agent", readAgentContent(s.T(), destination, "code.md"))
}

func (s *AgentInstallCommandTestSuite) TestHandleCloneFailure() {
	mockProcess := frameworkmock.Factory().Process()
	destination := filepath.Join(s.T().TempDir(), "agents")
	cloneError := errors.New("clone failed")

	mockProcess.EXPECT().Quietly().Return(mockProcess).Once()
	mockProcess.EXPECT().WithSpinner("Downloading Goravel agents").Return(mockProcess).Once()
	mockProcessResult := mocksprocess.NewResult(s.T())
	mockProcessResult.EXPECT().Failed().Return(true).Once()
	mockProcessResult.EXPECT().Error().Return(cloneError).Once()
	mockProcess.EXPECT().Run("git", "clone", "--depth=1", agentsRepo, mock.Anything).Return(mockProcessResult).Once()

	mockContext := newAgentInstallContext(s.T(), destination, false)
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		s.NoError(s.agentInstallCommand.Handle(mockContext))
	})

	s.Contains(captureOutput, "failed to clone goravel agents: clone failed")
	s.NoDirExists(destination)
}

func newAgentInstallContext(t *testing.T, destination string, force bool) *mocksconsole.Context {
	t.Helper()

	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Option("path").Return(destination).Once()
	mockContext.EXPECT().OptionBool("force").Return(force).Once()

	return mockContext
}

func expectAgentsCloneForAgents(t *testing.T, mockProcess *mocksprocess.Process, files map[string]string) {
	t.Helper()

	mockProcess.EXPECT().Quietly().Return(mockProcess).Once()
	mockProcess.EXPECT().WithSpinner("Downloading Goravel agents").Return(mockProcess).Once()
	mockProcessResult := mocksprocess.NewResult(t)
	mockProcessResult.EXPECT().Failed().Return(false).Once()
	mockProcess.EXPECT().Run("git", "clone", "--depth=1", agentsRepo, mock.Anything).RunAndReturn(func(name string, args ...string) process.Result {
		agentsPath := filepath.Join(args[3], "agents")
		if err := os.MkdirAll(agentsPath, 0755); err != nil {
			t.Fatalf("os.MkdirAll(%q) = %v, want nil", agentsPath, err)
		}
		for rel, content := range files {
			writeAgentContent(t, agentsPath, rel, content)
		}

		return mockProcessResult
	}).Once()
}

func writeAgentContent(t *testing.T, basePath, rel, content string) {
	t.Helper()

	path := filepath.Join(basePath, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("os.MkdirAll(%q) = %v, want nil", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("os.WriteFile(%q) = %v, want nil", path, err)
	}
}

func readAgentContent(t *testing.T, basePath, rel string) string {
	t.Helper()

	path := filepath.Join(basePath, rel)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) = %v, want nil", path, err)
	}

	return string(content)
}
