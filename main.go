package main

import (
	"os"

	"github.com/goravel/framework/console"

	installerconsole "github.com/goravel/installer/console"
)

func main() {
	cliApp := console.NewApplication()
	kernel := &installerconsole.Kernel{}
	cliApp.Register(kernel.Commands())
	cliApp.Run(os.Args, false)
}
