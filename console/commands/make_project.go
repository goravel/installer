package commands

import (
	"errors"
	"regexp"

	"github.com/gookit/color"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"

	"github.com/goravel/installer/support"
	"github.com/goravel/installer/ui"
)

type NewCommand struct {
}

// Signature The name and signature of the console command.
func (receiver *NewCommand) Signature() string {
	return "new"
}

// Description The console command description.
func (receiver *NewCommand) Description() string {
	return "Create a new Goravel application"
}

// Extend The console command extend.
func (receiver *NewCommand) Extend() command.Extend {
	return command.Extend{
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:  "dev",
				Usage: "Installs the latest 'development' release",
			},
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Forces install even if the directory already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *NewCommand) Handle(ctx console.Context) (err error) {
	color.Println(ui.LogoStyle.Render(support.WelcomeHeading))
	name := ctx.Argument(0)
	if name == "" {
		ctx.NewLine()
		name, err = ctx.Ask("What is the name of your project?", console.AskOption{
			Placeholder: "E.g example-app",
			Prompt:      ">",
			Validate: func(value string) error {
				if value == "" {
					return errors.New("the project name is required")
				}

				if !regexp.MustCompile(`^[\w.-]+$`).MatchString(value) {
					return errors.New("the name may only contain letters, numbers, dashes, underscores, and periods")
				}

				return nil
			},
		})
		if err != nil {
			return err
		}
	}

	ctx.Info("Creating project " + name)

	return nil
}
