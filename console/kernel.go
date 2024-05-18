package console

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/installer/console/commands"
)

type Kernel struct {
}

func (kernel *Kernel) Commands() []console.Command {
	return []console.Command{
		&commands.MakeProjectCommand{},
	}
}
