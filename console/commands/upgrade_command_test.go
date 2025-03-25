package commands

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/console"
	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/installer/support"
)

func TestUpgradeCommand(t *testing.T) {
	upgradeCommand := &UpgradeCommand{}

	assert.Equal(t, upgradeCommand.Signature(), "upgrade")
	assert.Equal(t, upgradeCommand.Description(), "Upgrade Goravel installer")
	assert.Equal(t, upgradeCommand.Extend().ArgsUsage, " [version]")
	assert.Len(t, upgradeCommand.Extend().Flags, 0)

	mockContext := &consolemocks.Context{}

	// upgrade failed
	mockContext.EXPECT().Argument(0).Return("unknown").Once()
	mockContext.EXPECT().Spinner(
		fmt.Sprintf("> @go install %s@unknown", support.InstallerModuleName),
		mock.AnythingOfType("console.SpinnerOption")).
			RunAndReturn(func(_ string, option console.SpinnerOption) error {
				return option.Action()
			}).Once()
	captureOutput := color.CaptureOutput(func(w io.Writer) {
		assert.NoError(t, upgradeCommand.Handle(mockContext))
	})
	assert.Contains(t, captureOutput, "Failed to upgrade Goravel installer")
	assert.Contains(t, captureOutput, "invalid version: unknown revision unknown")

	//upgrade success
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Spinner(
		fmt.Sprintf("> @go install %s", support.InstallerModuleName),
		mock.AnythingOfType("console.SpinnerOption")).Return(nil).Once()
	captureOutput = color.CaptureOutput(func(w io.Writer) {
		assert.NoError(t, upgradeCommand.Handle(mockContext))
	})
	assert.Contains(t, captureOutput, "Goravel installer has been upgraded successfully")

}
