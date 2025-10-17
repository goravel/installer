package console

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/process"

	"github.com/goravel/installer/console/commands"
)

type Kernel struct {
}

func (kernel *Kernel) Commands() []console.Command {
	return []console.Command{
		commands.NewNewCommand(process.New()),
		&commands.UpgradeCommand{},
	}
}
