package commands

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
