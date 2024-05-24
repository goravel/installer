package commands

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
	"github.com/pterm/pterm"

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
	color.Default().Println(ui.LogoStyle.Render(support.WelcomeHeading))
	ctx.NewLine()
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
			ctx.NewLine()
			color.Errorln(err.Error())
			ctx.NewLine()
			return nil
		}
	}

	force := ctx.OptionBool("force")
	if !force {
		if receiver.verifyIfDirectoryExists(receiver.getPath(name)) {
			ctx.NewLine()
			color.Errorln("the directory already exists. use the --force flag to overwrite")
			ctx.NewLine()
			return nil
		}
	}

	return receiver.generate(ctx, name)
}

// verifyIfDirectoryExists Verify if the directory already exists.
func (receiver *NewCommand) verifyIfDirectoryExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// getPath Get the full path to the command.
func (receiver *NewCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	return filepath.Clean(filepath.Join(pwd, name))
}

// generate Generate the project.
func (receiver *NewCommand) generate(ctx console.Context, name string) error {
	path := receiver.getPath(name)
	name = filepath.Clean(name)
	bold := pterm.NewStyle(pterm.Bold)

	// remove the directory if it already exists
	if err := os.RemoveAll(path); err != nil {
		ctx.NewLine()
		color.Errorf("error while removing the directory : %s\n", err.Error())
		ctx.NewLine()
		return nil
	}

	// clone the repository
	clone := exec.Command("git", "clone", "https://github.com/goravel/goravel.git", path)
	color.Green().Println("creating a \"goravel/goravel\" project at \"" + name + "\"")
	if err := clone.Run(); err != nil {
		ctx.NewLine()
		color.Errorf("error while generating the project : %s\n", err.Error())
		ctx.NewLine()
		return nil
	}
	color.Green().Println("created project in " + path)

	// git cleanup
	color.Default().Println("> @rm -rf " + name + "/.git " + name + "/.github")
	var removeFiles *exec.Cmd
	if runtime.GOOS == "windows" {
		removeFiles = exec.Command("Remove-Item", "-Path", path+"/.git", path+"/.github", "-Recursive", "-Force")
	} else {
		removeFiles = exec.Command("rm", "-rf", path+"/.git", path+"/.github")
	}
	if err := removeFiles.Run(); err != nil {
		ctx.NewLine()
		color.Errorf("error happend while removing the files : %s\n", err)
		ctx.NewLine()
		return nil
	}
	ctx.NewLine()
	color.Infoln("git cleanup done")
	ctx.NewLine()

	// install dependencies
	color.Default().Println("> @go mod tidy")
	install := exec.Command("go", "mod", "tidy")
	install.Dir = path
	if err := install.Run(); err != nil {
		ctx.NewLine()
		color.Errorf("error while installing the dependecies : %s\n", err)
		ctx.NewLine()
		return nil
	}
	ctx.NewLine()
	color.Infoln("goravel installed successfully!")
	ctx.NewLine()

	// generate .env file
	color.Default().Println("> @cp .env.example .env")
	copyEnv := exec.Command("cp", ".env.example", ".env")
	copyEnv.Dir = path
	if err := copyEnv.Run(); err != nil {
		ctx.NewLine()
		color.Errorf("error while generating the .env file : %s\n", err)
		ctx.NewLine()
		return nil
	}
	ctx.NewLine()
	color.Infoln(".env file generated successfully!")
	ctx.NewLine()

	// generate app key
	color.Default().Println("> @go run . artisan key:generate")
	initAppKey := exec.Command("go", "run", ".", "artisan", "key:generate")
	initAppKey.Dir = path
	if err := initAppKey.Run(); err != nil {
		ctx.NewLine()
		color.Errorf("error while generating the app key : %s\n", err)
		ctx.NewLine()
		return nil
	}
	ctx.NewLine()
	color.Infoln("App key generated successfully!")
	ctx.NewLine()

	color.Infoln("Application ready in [" + bold.Sprintf("%s", name) + "]. Build something amazing!")
	ctx.NewLine()
	color.Infoln("Are you new to Goravel? Please visit https://goravel.dev to get started. Build something amazing!")
	ctx.NewLine()
	return nil
}
