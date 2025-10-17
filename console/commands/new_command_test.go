package commands

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/mocks/process"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewCommand(t *testing.T) {
	mockProcess := process.NewProcess(t)
	newCommand := &NewCommand{process: mockProcess}

	assert.Equal(t, newCommand.Signature(), "new")
	assert.Equal(t, newCommand.Description(), "Create a new Goravel application")
	assert.Contains(t, newCommand.Extend().Flags, &command.BoolFlag{
		Name:               "force",
		Aliases:            []string{"f"},
		Usage:              "Forces install even if the directory already exists",
		DisableDefaultText: true,
	})

	mockContext := consolemocks.NewContext(t)
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("What is the name of your project?", mock.Anything).Return("", errors.New("the project name is required")).Once()
	mockContext.EXPECT().NewLine().Return()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, newCommand.Handle(mockContext))
	}), "the project name is required")

	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("What is the name of your project?", mock.Anything).Return("invalid:name", nil).Once()
	mockContext.EXPECT().NewLine().Return()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, newCommand.Handle(mockContext))
	}), "the name only supports letters, numbers, dashes, underscores, and periods")

	mockContext.EXPECT().Argument(0).Return("invalid:name").Once()
	mockContext.EXPECT().NewLine().Return()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, newCommand.Handle(mockContext))
	}), "the name only supports letters, numbers, dashes, underscores, and periods")

	mockContext.EXPECT().Argument(0).Return("example-app").Once()
	mockContext.EXPECT().OptionBool("force").Return(true).Once()
	mockContext.EXPECT().Option("module").Return("").Once()
	mockContext.EXPECT().Ask("What is the module name?", mock.Anything).Return("invalid:module", nil).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, newCommand.Handle(mockContext))
	}), "invalid module name format. Use only letters, numbers, dots (.), slashes (/), underscores (_), hyphens (-), and tildes (~). Example: [github.com/yourusername/yourproject] or [yourproject]")

	mockContext.EXPECT().Argument(0).Return("example-app").Once()
	mockContext.EXPECT().OptionBool("force").Return(true).Once()
	mockContext.EXPECT().Option("module").Return("invalid:module").Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, newCommand.Handle(mockContext))
	}), "invalid module name format. Use only letters, numbers, dots (.), slashes (/), underscores (_), hyphens (-), and tildes (~). Example: [github.com/yourusername/yourproject] or [yourproject]")

	mockContext.EXPECT().Argument(0).Return("example-app").Once()
	mockContext.EXPECT().OptionBool("force").Return(true).Once()
	mockContext.EXPECT().Option("module").Return("").Once()
	mockContext.EXPECT().Ask("What is the module name?", mock.Anything).Return("github.com/example/", nil).Once()
	mockContext.EXPECT().OptionBool("dev").Return(true).Once()
	mockContext.EXPECT().Spinner(`Creating a "goravel/goravel-lite" project at "example-app"`, mock.Anything).Return(nil).
		Run(func(_ string, option console.SpinnerOption) {
			assert.NoError(t, option.Action())
		}).Once()

	mockContext.EXPECT().Spinner("> @rm -rf example-app/.git example-app/.github", mock.Anything).Return(nil).
		Run(func(_ string, option console.SpinnerOption) {
			assert.NoError(t, option.Action())
		}).Once()
	mockContext.EXPECT().Spinner(`Updating module name to "github.com/example/"`, mock.Anything).Return(nil).
		Run(func(_ string, option console.SpinnerOption) {
			assert.NoError(t, option.Action())
		}).Once()
	mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(nil).
		Run(func(_ string, option console.SpinnerOption) {
			assert.NoError(t, option.Action())
		}).Once()
	mockContext.EXPECT().Spinner("> @cp .env.example .env", mock.Anything).Return(nil).
		Run(func(_ string, option console.SpinnerOption) {
			assert.NoError(t, option.Action())
		}).Once()
	mockContext.EXPECT().Spinner("> @go run . artisan key:generate", mock.Anything).Return(nil).
		Run(func(_ string, option console.SpinnerOption) {
			assert.NoError(t, option.Action())
		}).Once()

	mockProcess.EXPECT().Path(mock.MatchedBy(func(path string) bool {
		return strings.Contains(path, "example-app")
	})).Return(mockProcess).Once()
	mockProcess.EXPECT().TapCmd(mock.AnythingOfType("func(*exec.Cmd)")).Return(mockProcess).Once()
	mockResult := process.NewResult(t)
	mockProcess.EXPECT().Run("go", "run", ".", "artisan", "package:install").Return(mockResult, nil).Once()
	mockResult.EXPECT().Error().Return(nil).Once()
	mockResult.EXPECT().ErrorOutput().Return("").Once()

	captureOutput := color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, newCommand.Handle(mockContext))
	})

	assert.Contains(t, captureOutput, ".env file generated successfully!")
	assert.Contains(t, captureOutput, "App key generated successfully!")
	assert.True(t, file.Exists("example-app"))
	assert.True(t, file.Exists(filepath.Join("example-app", ".env")))
	if !env.IsWindows() {
		artisan := filepath.Join("example-app", "artisan")
		info, err := os.Stat(artisan)
		assert.Nil(t, err)
		assert.Equal(t, info.Mode().Perm(), os.FileMode(0755))
	}

	mockContext.EXPECT().Argument(0).Return("example-app").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, newCommand.Handle(mockContext))
	}), "the directory already exists. use the --force flag to overwrite")

	assert.Nil(t, file.Remove("example-app"))
}

func TestCopyFile(t *testing.T) {
	newCommand := &NewCommand{}

	tmpDir, err := os.MkdirTemp("", "test-copy-file")
	assert.Nil(t, err)
	src := filepath.Join(tmpDir, ".env.example")
	dst := filepath.Join(tmpDir, ".env")

	// Create a mock .env.example file for testing
	err = os.WriteFile(src, []byte("example env"), os.ModePerm)
	assert.Nil(t, err)
	assert.True(t, file.Exists(src))
	assert.Nil(t, newCommand.copyFile(src, dst))
	assert.True(t, file.Exists(dst))
	assert.Nil(t, os.Remove(src))
	assert.Nil(t, os.Remove(dst))
}

func TestReplaceModule(t *testing.T) {
	newCommand := &NewCommand{}

	tmpDir, err := os.MkdirTemp("", "test-replace-module")
	assert.Nil(t, err)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	goFile := filepath.Join(tmpDir, "main.go")
	modFile := filepath.Join(tmpDir, "go.mod")
	invalidFile := filepath.Join(tmpDir, "ignore.txt")

	newModule := "github.com/example/project"

	// Write content to files
	err = os.WriteFile(goFile, []byte(`package main
import "goravel/utils"
func main() {}`), os.ModePerm)
	assert.Nil(t, err)

	err = os.WriteFile(modFile, []byte("module goravel\nrequire example.com v1.0.0"), os.ModePerm)
	assert.Nil(t, err)

	err = os.WriteFile(invalidFile, []byte("This is a test file."), os.ModePerm)
	assert.Nil(t, err)

	err = newCommand.replaceModule(tmpDir, newModule)
	assert.Nil(t, err)

	modContent, err := os.ReadFile(modFile)
	assert.Nil(t, err)
	assert.Contains(t, string(modContent), "module github.com/example/project")
	assert.NotContains(t, string(modContent), "module goravel")

	goContent, err := os.ReadFile(goFile)
	assert.Nil(t, err)
	assert.Contains(t, string(goContent), "import \"github.com/example/project/utils\"")
	assert.NotContains(t, string(goContent), "import \"goravel/utils\"")

	invalidContent, err := os.ReadFile(invalidFile)
	assert.Nil(t, err)
	assert.Equal(t, "This is a test file.", string(invalidContent))
}
