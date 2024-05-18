package commands

import (
	"github.com/gookit/color"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"

	"github.com/goravel/installer/support"
	"github.com/goravel/installer/ui"
)

type MakeProjectCommand struct {
}

// Signature The name and signature of the console command.
func (receiver *MakeProjectCommand) Signature() string {
	return "make:project"
}

// Description The console command description.
func (receiver *MakeProjectCommand) Description() string {
	return "Create a new project"
}

// Extend The console command extend.
func (receiver *MakeProjectCommand) Extend() command.Extend {
	return command.Extend{}
}

// Handle Execute the console command.
func (receiver *MakeProjectCommand) Handle(ctx console.Context) (err error) {
	color.Println(ui.LogoStyle.Render(support.WelcomeHeading))
	name := ctx.Argument(0)
	if name == "" {
		ctx.NewLine()
		name, err = ctx.Ask("What is the name of your project?", console.AskOption{
			Placeholder: "E.g my-new-app",
			Prompt:      ">",
		})
		if err != nil {
			return err
		}
	}

	ctx.Info("Creating project " + name)

	return nil
}
