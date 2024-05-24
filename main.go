package main

import (
	"os"

	frameworkconsole "github.com/goravel/framework/console"

	"github.com/goravel/installer/console"
	"github.com/goravel/installer/support"
)

func main() {
	cliApp := frameworkconsole.NewApplication("Goravel Installer", support.Version, "go run . [global options] command [command options] [arguments...]", support.Version, false)
	kernel := &console.Kernel{}
	cliApp.Register(kernel.Commands())
	cliApp.Run(os.Args, false)
}
