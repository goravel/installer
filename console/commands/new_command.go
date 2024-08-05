package commands

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
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
	fmt.Println(pterm.NewRGB(142, 211, 249).Sprint(support.WelcomeHeading)) // color hex code: #8ED3F9
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
	err := ctx.Spinner("Creating a \"goravel/goravel\" project at \""+name+"\"", console.SpinnerOption{
		Action: func() error {
			return clone.Run()
		},
	})
	if err != nil {
		color.Errorf("error while generating the project : %s\n", err.Error())
		return nil
	}
	color.Successln("created project in " + path)

	// git cleanup
	err = ctx.Spinner("> @rm -rf "+name+"/.git "+name+"/.github", console.SpinnerOption{
		Action: func() error {
			return receiver.removeFiles(path)
		},
	})
	if err != nil {
		color.Errorf("error happend while removing the files : %s\n", err)
		return nil
	}
	color.Successln("git cleanup done")

	// install dependencies
	install := exec.Command("go", "mod", "tidy")
	install.Dir = path
	err = ctx.Spinner("> @go mod tidy", console.SpinnerOption{
		Action: func() error {
			return install.Run()
		},
	})
	if err != nil {
		color.Errorf("error while installing the dependecies : %s\n", err)
		return nil
	}
	color.Successln("Goravel installed successfully!")

	// generate .env file
	err = ctx.Spinner("> @cp .env.example .env", console.SpinnerOption{
		Action: func() error {
			inputFilePath := filepath.Join(path, ".env.example")
			outputFilePath := filepath.Join(path, ".env")
			return receiver.copyFile(inputFilePath, outputFilePath)
		},
	})
	if err != nil {
		color.Errorf("error while generating the .env file : %s\n", err)
		return nil
	}
	color.Successln(".env file generated successfully!")

	// generate app key
	initAppKey := exec.Command("go", "run", ".", "artisan", "key:generate")
	initAppKey.Dir = path
	err = ctx.Spinner("> @go run . artisan key:generate", console.SpinnerOption{
		Action: func() error {
			return initAppKey.Run()
		},
	})
	if err != nil {
		color.Errorf("error while generating the app key : %s\n", err)
		return nil
	}

	color.Successln("App key generated successfully!")
	color.Successln("Application ready in [" + bold.Sprintf("%s", name) + "]. Build something amazing!")
	color.Successln("Are you new to Goravel? Please visit https://goravel.dev to get started.")
	return nil
}

func (receiver *NewCommand) removeFiles(path string) error {
	// Remove the .git directory
	if err := file.Remove(filepath.Join(path, ".git")); err != nil {
		return err
	}

	// Remove the .GitHub directory
	return file.Remove(filepath.Join(path, ".github"))
}

func (receiver *NewCommand) copyFile(inputFilePath, outputFilePath string) (err error) {
	// Open .env.example file
	in, err := os.Open(inputFilePath)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := in.Close(); cerr != nil {
			if err == nil {
				err = cerr
			} else {
				fmt.Printf("Error closing input file: %v\n", cerr)
			}
		}
	}()

	// Create .env file
	out, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := out.Close(); cerr != nil {
			if err == nil {
				err = cerr
			} else {
				fmt.Printf("Error closing output file: %v\n", cerr)
			}
		}
	}()

	// Copy .env.example to .env file
	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	return nil
}
