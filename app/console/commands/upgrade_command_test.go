package commands

import (
	"fmt"
	"io"
	"testing"

	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksprocess "github.com/goravel/framework/mocks/process"
	"github.com/goravel/framework/support/color"
	frameworkmock "github.com/goravel/framework/testing/mock"
	"github.com/stretchr/testify/assert"
)

func TestUpgradeCommand(t *testing.T) {
	mockFactory := frameworkmock.Factory()
	mockProcess := mockFactory.Process()

	upgradeCommand := NewUpgradeCommand()
	pkg := "github.com/goravel/installer/goravel"

	t.Run("failed", func(t *testing.T) {
		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().ArgumentString("version").Return("unknown").Once()
		mockProcess.EXPECT().WithSpinner().Return(mockProcess).Once()

		mockProcessResult := mocksprocess.NewResult(t)
		mockProcessResult.EXPECT().Failed().Return(true).Once()
		mockProcessResult.EXPECT().Error().Return(assert.AnError).Once()
		mockProcess.EXPECT().Run("go", "install", fmt.Sprintf("%s@unknown", pkg)).Return(mockProcessResult).Once()

		captureOutput := color.CaptureOutput(func(w io.Writer) {
			assert.NoError(t, upgradeCommand.Handle(mockContext))
		})
		assert.Contains(t, captureOutput, fmt.Sprintf("Failed to upgrade Goravel installer: %s", assert.AnError.Error()))
	})

	t.Run("happy path", func(t *testing.T) {
		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().ArgumentString("version").Return("latest").Once()
		mockProcess.EXPECT().WithSpinner().Return(mockProcess).Once()

		mockProcessResult := mocksprocess.NewResult(t)
		mockProcessResult.EXPECT().Failed().Return(false).Once()
		mockProcess.EXPECT().Run("go", "install", fmt.Sprintf("%s@latest", pkg)).Return(mockProcessResult).Once()

		captureOutput := color.CaptureOutput(func(w io.Writer) {
			assert.NoError(t, upgradeCommand.Handle(mockContext))
		})

		assert.Contains(t, captureOutput, "Goravel installer has been upgraded successfully")
	})
}
