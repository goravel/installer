package commands

import (
	"errors"
	"io"
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
	mockContext.On("Argument", 0).Return("").Once()
	mockContext.On("Ask", "What is the name of your project?", mock.Anything).Return("", errors.New("the project name is required")).Once()
	mockContext.On("NewLine").Return()
	mockContext.On("Spinner", mock.Anything, mock.AnythingOfType("console.SpinnerOption")).Return(nil).
		Run(func(args mock.Arguments) {
			options := args.Get(1).(console.SpinnerOption)
			assert.Nil(t, options.Action())
		}).Times(5)
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, newCommand.Handle(mockContext))
	}), "the project name is required")

	mockContext.On("Argument", 0).Return("example-app").Once()
	mockContext.On("OptionBool", "force").Return(true).Once()
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, newCommand.Handle(mockContext))
	})
	assert.Contains(t, captureOutput, ".env file generated successfully!")
	assert.Contains(t, captureOutput, "App key generated successfully!")
	assert.True(t, file.Exists("example-app"))

	// Test copyFile
	src := filepath.Join("example-app", ".env.example")
	dst := filepath.Join("example-app", ".env")
	assert.Nil(t, newCommand.copyFile(src, dst))
	assert.True(t, file.Exists(dst))

	mockContext.On("Argument", 0).Return("example-app").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, newCommand.Handle(mockContext))
	}), "the directory already exists. use the --force flag to overwrite")

	assert.Nil(t, file.Remove("example-app"))
	mockContext.AssertExpectations(t)
}
