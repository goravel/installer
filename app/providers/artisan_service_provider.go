package providers

import (
	"io"
	"os"
	"slices"

	"github.com/goravel/framework/console"
	"github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/support/color"

	"github.com/goravel/installer/app/console/commands"
	"github.com/goravel/installer/support"
)

type ArtisanServiceProvider struct {
}

func (r *ArtisanServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings: []string{
			binding.Artisan,
		},
		Dependencies: binding.Bindings[binding.Artisan].Dependencies,
		ProvideFor:   []string{},
	}
}

func (r *ArtisanServiceProvider) Register(app foundation.Application) {
	app.Singleton(binding.Artisan, func(app foundation.Application) (any, error) {
		return NewApplication(), nil
	})
}

func (r *ArtisanServiceProvider) Boot(app foundation.Application) {
	artisanFacade := app.MakeArtisan()
	if artisanFacade == nil {
		color.Warningln("Artisan Facade is not initialized. Skipping command registration.")
		return
	}

	artisanFacade.Register([]contractsconsole.Command{
		commands.NewNewCommand(),
		commands.NewUpgradeCommand(),
	})
}

type Application struct {
	*console.Application
	commands   []contractsconsole.Command
	name       string
	usage      string
	usageText  string
	useArtisan bool
	version    string
	writer     io.Writer
}

func NewApplication() contractsconsole.Artisan {
	name := "goravel"
	usage := "Goravel Installer"
	usageText := "goravel [global options] command [command options] [arguments...]"
	version := support.Version
	useArtisan := false

	return &Application{
		Application: console.NewApplication(name, usage, usageText, version, useArtisan),
		name:        name,
		usage:       usage,
		usageText:   usageText,
		useArtisan:  useArtisan,
		version:     version,
		writer:      os.Stdout,
	}
}

func (r *Application) Register(commands []contractsconsole.Command) {
	for _, item := range commands {
		if slices.Contains([]string{"list", "new", "upgrade"}, item.Signature()) {
			r.commands = append(r.commands, item)
		}
	}

	r.SetCommands(r.commands)
}
