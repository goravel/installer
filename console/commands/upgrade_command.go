package commands

import (
	"os/exec"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
	supportconsole "github.com/goravel/framework/support/console"

	"github.com/goravel/installer/support"
)

type UpgradeCommand struct {
}

// Signature The name and signature of the console command.
func (r *UpgradeCommand) Signature() string {
	return "upgrade"
}

// Description The console command description.
func (r *UpgradeCommand) Description() string {
	return "Upgrade Goravel installer"
}

// Extend The console command extend.
func (r *UpgradeCommand) Extend() command.Extend {
	return command.Extend{
		ArgsUsage: " [version]",
	}
}

// Handle Execute the console command.
func (r *UpgradeCommand) Handle(ctx console.Context) error {
	pkg := support.InstallerModuleName
	version := ctx.Argument(0)
	if version == "" {
		version = "latest"
	}

	upgrade := exec.Command("go", "install", pkg+"@"+version)
	if err := supportconsole.ExecuteCommand(ctx, upgrade); err != nil {
		color.Errorln("Failed to upgrade Goravel installer")
		color.Red().Println(err.Error())

		return nil
	}

	color.Successln("Goravel installer has been upgraded successfully")

	return nil
}
