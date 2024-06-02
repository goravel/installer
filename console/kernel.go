package console

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/console/commands"
)

type Kernel struct {
}

func (kernel *Kernel) Commands() []console.Command {
	return []console.Command{
		&commands.NewCommand{},
	}
}
