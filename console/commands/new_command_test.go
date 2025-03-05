package commands

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
)

func TestNewCommand(t *testing.T) {
	newCommand := &NewCommand{}

	assert.Equal(t, newCommand.Signature(), "new")
	assert.Equal(t, newCommand.Description(), "Create a new Goravel application")
	assert.Equal(t, newCommand.Extend().Flags[0], &command.BoolFlag{
		Name:    "force",
		Aliases: []string{"f"},
		Usage:   "Forces install even if the directory already exists",
	})

	mockContext := &consolemocks.Context{}
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("What is the name of your project?", mock.Anything).Return("", errors.New("the project name is required")).Once()
	mockContext.EXPECT().NewLine().Return()
	mockContext.EXPECT().Spinner(mock.Anything, mock.AnythingOfType("console.SpinnerOption")).Return(nil).
		Run(func(message string, options console.SpinnerOption) {
			assert.Nil(t, options.Action())
		}).Times(6)
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, newCommand.Handle(mockContext))
	}), "the project name is required")

	mockContext.EXPECT().Argument(0).Return("example-app").Once()
	mockContext.EXPECT().OptionBool("force").Return(true).Once()
	mockContext.EXPECT().Option("module").Return("").Once()
	mockContext.EXPECT().Ask("What is the module name?", mock.Anything).Return("github.com/example/", nil).Once()
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, newCommand.Handle(mockContext))
	})
	assert.Contains(t, captureOutput, ".env file generated successfully!")
	assert.Contains(t, captureOutput, "App key generated successfully!")
	assert.True(t, file.Exists("example-app"))
	assert.True(t, file.Exists(filepath.Join("example-app", ".env")))

	mockContext.EXPECT().Argument(0).Return("example-app").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, newCommand.Handle(mockContext))
	}), "the directory already exists. use the --force flag to overwrite")

	assert.Nil(t, file.Remove("example-app"))
	mockContext.AssertExpectations(t)
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
	defer os.RemoveAll(tmpDir)

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
