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
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Forces install even if the directory already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *NewCommand) Handle(ctx console.Context) (err error) {
	color.Cyan().Println(support.WelcomeHeading)
	ctx.NewLine()
	name := ctx.Argument(0)
	if name == "" {
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
			color.Errorln(err.Error())
			return nil
		}
	}

	force := ctx.OptionBool("force")
	if !force && receiver.verifyIfDirectoryExists(receiver.getPath(name)) {
		color.Errorln("the directory already exists. use the --force flag to overwrite")
		return nil
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
		color.Errorf("error while removing the directory : %s\n", err.Error())
		return nil
	}

	// clone the repository
	clone := exec.Command("git", "clone", "https://github.com/goravel/goravel.git", path)
	color.Successln("creating a \"goravel/goravel\" project at \"" + name + "\"")
	if err := clone.Run(); err != nil {
		color.Errorf("error while generating the project : %s\n", err.Error())
		return nil
	}
	color.Successln("created project in " + path)

	// git cleanup
	color.Default().Println("> @rm -rf " + name + "/.git " + name + "/.github")
	var removeFiles *exec.Cmd
	if runtime.GOOS == "windows" {
		removeFiles = exec.Command("Remove-Item", "-Path", path+"/.git", path+"/.github", "-Recursive", "-Force")
	} else {
		removeFiles = exec.Command("rm", "-rf", path+"/.git", path+"/.github")
	}
	if err := removeFiles.Run(); err != nil {
		color.Errorf("error happend while removing the files : %s\n", err)
		return nil
	}
	color.Successln("git cleanup done")

	// install dependencies
	color.Default().Println("> @go mod tidy")
	install := exec.Command("go", "mod", "tidy")
	install.Dir = path
	if err := install.Run(); err != nil {
		color.Errorf("error while installing the dependecies : %s\n", err)
		return nil
	}
	color.Successln("goravel installed successfully!")

	// generate .env file
	color.Default().Println("> @cp .env.example .env")
	copyEnv := exec.Command("cp", ".env.example", ".env")
	copyEnv.Dir = path
	if err := copyEnv.Run(); err != nil {
		color.Errorf("error while generating the .env file : %s\n", err)
		return nil
	}
	color.Successln(".env file generated successfully!")

	// generate app key
	color.Default().Println("> @go run . artisan key:generate")
	initAppKey := exec.Command("go", "run", ".", "artisan", "key:generate")
	initAppKey.Dir = path
	if err := initAppKey.Run(); err != nil {
		color.Errorf("error while generating the app key : %s\n", err)
		return nil
	}
	color.Successln("App key generated successfully!")
	color.Successln("Application ready in [" + bold.Sprintf("%s", name) + "]. Build something amazing!")
	color.Successln("Are you new to Goravel? Please visit https://goravel.dev to get started.")
	return nil
}
