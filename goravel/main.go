package main

import (
	"os"

	frameworkconsole "github.com/goravel/framework/console"

	"github.com/goravel/installer/console"
	"github.com/goravel/installer/support"
)

func main() {
	name := "goravel"
	usage := "Goravel Installer"
	usageText := "goravel [global options] command [command options] [arguments...]"

	cliApp := frameworkconsole.NewApplication(name, usage, usageText, support.Version, false)

	kernel := &console.Kernel{}

	cliApp.Register(kernel.Commands())
	cliApp.Run(os.Args, false)
}
