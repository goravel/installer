package commands

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"

	"github.com/goravel/installer/app/facades"
)

type UpgradeCommand struct{}

func NewUpgradeCommand() *UpgradeCommand {
	return &UpgradeCommand{}
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
		Arguments: []command.Argument{
			&command.ArgumentStringSlice{
				Name:  "version",
				Value: "latest",
				Usage: "The version of Goravel installer to upgrade to (default: latest)",
			},
		},
	}
}

// Handle Execute the console command.
func (r *UpgradeCommand) Handle(ctx console.Context) error {
	pkg := "github.com/goravel/installer/goravel"
	version := ctx.ArgumentString("version")

	if res := facades.Process().WithLoading().Run("go", "install", pkg+"@"+version); res.Failed() {
		color.Errorf("Failed to upgrade Goravel installer: %s\n", res.Error())

		return nil
	}

	color.Successln("Goravel installer has been upgraded successfully")

	return nil
}
