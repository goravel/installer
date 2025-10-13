package commands

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"

	"github.com/goravel/installer/support"
)

type NewCommand struct {
}

// Signature The name and signature of the console command.
func (r *NewCommand) Signature() string {
	return "new"
}

// Description The console command description.
func (r *NewCommand) Description() string {
	return "Create a new Goravel application"
}

// Extend The console command extend.
func (r *NewCommand) Extend() command.Extend {
	return command.Extend{
		ArgsUsage: " [--] <name>",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:  "cache",
				Usage: "The cache driver your application will use",
			},
			&command.StringFlag{
				Name:  "database",
				Usage: "The database driver your application will use",
			},
			&command.BoolFlag{
				Name:               "dev",
				Usage:              `Install the latest "development" release`,
				DisableDefaultText: true,
			},
			&command.BoolFlag{
				Name:               "force",
				Aliases:            []string{"f"},
				Usage:              "Forces install even if the directory already exists",
				DisableDefaultText: true,
			},
			&command.StringFlag{
				Name:  "http",
				Usage: "The HTTP driver your application will use",
			},
			&command.StringFlag{
				Name:    "module",
				Aliases: []string{"m"},
				Usage:   "Specify the custom module name to replace the default 'goravel' module",
			},
			&command.StringFlag{
				Name:  "queue",
				Usage: "The queue driver your application will use",
			},
			&command.StringFlag{
				Name:  "session",
				Usage: "The session driver your application will use",
			},
			&command.StringSliceFlag{
				Name:  "storage",
				Usage: "The filesystem modules your application will use",
			},
		},
	}
}

// Handle Execute the console command.
func (r *NewCommand) Handle(ctx console.Context) (err error) {
	color.Printfln("<fg=52,124,153>%s</>", support.WelcomeHeading) // color hex code: #8ED3F9
	ctx.NewLine()
	name := ctx.Argument(0)
	if name == "" {
		name, err = ctx.Ask("What is the name of your project?", console.AskOption{
			Placeholder: "E.g example-app",
			Prompt:      "> ",
			Validate: func(value string) error {
				if value == "" {
					return errors.New("the project name is required")
				}

				return nil
			},
		})
		if err != nil {
			color.Errorln(err.Error())
			return nil
		}
	}

	if !regexp.MustCompile(`^[\w.-]+$`).MatchString(name) {
		color.Errorln("the name only supports letters, numbers, dashes, underscores, and periods")
		return nil
	}

	force := ctx.OptionBool("force")
	if !force && r.verifyIfDirectoryExists(r.getPath(name)) {
		color.Errorln("the directory already exists. use the --force flag to overwrite")
		return nil
	}

	module := ctx.Option("module")
	if module == "" {
		module, err = ctx.Ask("What is the module name?", console.AskOption{
			Placeholder: "E.g. github.com/yourusername/yourproject",
			Default:     support.DefaultModuleName,
			Prompt:      "> ",
			Validate: func(value string) error {
				if value == "" {
					return errors.New("module name is required")
				}

				return nil
			},
		})
		if err != nil {
			color.Errorln(err.Error())
			return nil
		}
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9./_-~]+$`).MatchString(module) {
		color.Errorln("invalid module name format. Use only letters, numbers, dots (.), slashes (/), underscores (_), hyphens (-), and tildes (~). Example: [github.com/yourusername/yourproject] or [yourproject]")
		return nil
	}

	if err = r.generateProject(ctx, name, module); err != nil {
		color.Errorln(err.Error())
		return nil
	}

	color.Successln("Application ready in [<op=bold>" + name + "</>]. Build something amazing!")
	color.Successln("Are you new to Goravel? Please visit https://goravel.dev to get started.")

	return
}

// verifyIfDirectoryExists Verify if the directory already exists.
func (r *NewCommand) verifyIfDirectoryExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// getPath Get the full path to the command.
func (r *NewCommand) getPath(name string) string {
	pwd, _ := os.Getwd()

	return filepath.Clean(filepath.Join(pwd, name))
}

func (r *NewCommand) generateProject(ctx console.Context, name string, module string) error {
	path := r.getPath(name)
	name = filepath.Clean(name)

	// remove the directory if it already exists
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove the directory: %s", err)
	}

	// clone the repository
	args := []string{"clone", "--depth=1", "https://github.com/goravel/goravel-lite.git", path}
	if ctx.OptionBool("dev") {
		args = slices.Insert(args, 2, "--branch=master")
	}
	clone := exec.Command("git", args...)
	if err := supportconsole.ExecuteCommand(ctx, clone, fmt.Sprintf(`Creating a "goravel/goravel" project at "%s"`, name)); err != nil {
		return fmt.Errorf("failed to clone goravel, please check your internet connection: %s", err)
	}
	color.Successln("created project in " + path)

	// git cleanup
	if err := ctx.Spinner("> @rm -rf "+name+"/.git "+name+"/.github", console.SpinnerOption{
		Action: func() error {
			return r.removeFiles(path)
		},
	}); err != nil {
		return fmt.Errorf("failed to remove .git and .github folders: %s", err)
	}
	color.Successln("git cleanup done")

	// Replace the module name if it's different from the default
	if module != support.DefaultModuleName {
		if err := ctx.Spinner("Updating module name to \""+module+"\"", console.SpinnerOption{
			Action: func() error {
				return r.replaceModule(path, module)
			},
		}); err != nil {
			return fmt.Errorf("failed to update module name: %s\n", err)
		}
		color.Successln("Module name updated successfully!")
	}

	// install dependencies
	install := exec.Command("go", "mod", "tidy")
	install.Dir = path
	if err := supportconsole.ExecuteCommand(ctx, install); err != nil {
		return fmt.Errorf("failed to install dependecies: %s", err)
	}
	color.Successln("Goravel installed successfully!")

	// generate .env file
	if err := ctx.Spinner("> @cp .env.example .env", console.SpinnerOption{
		Action: func() error {
			inputFilePath := filepath.Join(path, ".env.example")
			outputFilePath := filepath.Join(path, ".env")
			return r.copyFile(inputFilePath, outputFilePath)
		},
	}); err != nil {
		return fmt.Errorf("failed to generate .env file: %s", err)
	}
	color.Successln(".env file generated successfully!")

	// set execute permission
	if artisan := filepath.Join(path, "artisan"); file.Exists(artisan) {
		if err := os.Chmod(artisan, 0755); err != nil {
			color.Errorf("failed to set artisan execute permission: %s\n", err)
		} else {
			color.Successln("artisan execute permission set successfully!")
		}
	}

	// generate app key
	initAppKey := exec.Command("go", "run", ".", "artisan", "key:generate")
	initAppKey.Dir = path
	if err := supportconsole.ExecuteCommand(ctx, initAppKey); err != nil {

		return fmt.Errorf("failed to generate app key: %s", err)
	}

	color.Successln("App key generated successfully!")

	// install dependencies
	packageInstall := exec.Command("go", "run", ".", "artisan", "package:install")
	packageInstall.Dir = path
	if err := supportconsole.ExecuteCommand(ctx, packageInstall); err != nil {
		return fmt.Errorf("failed to install facades: %s", err)
	}
	color.Successln("Goravel installed successfully!")

	return nil
}

func (r *NewCommand) replaceModule(path, module string) error {
	module = strings.Trim(module, "/")
	reModule := regexp.MustCompile(`^module\s+goravel\b`)
	reImport := regexp.MustCompile(`"goravel/([^"]+)"`)

	return filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || (!strings.HasSuffix(filePath, ".go") && !strings.HasSuffix(filePath, ".mod")) {
			return err
		}

		fileContent, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("error opening %s: %w", filePath, err)
		}
		defer fileContent.Close()

		var newContent strings.Builder
		var modified bool
		scanner := bufio.NewScanner(fileContent)
		for scanner.Scan() {
			line := scanner.Text()
			var newLine string

			if strings.HasSuffix(filePath, ".mod") {
				newLine = reModule.ReplaceAllString(line, "module "+module)
			} else {
				newLine = reImport.ReplaceAllString(line, `"`+module+`/$1"`)
			}

			if newLine != line {
				modified = true
			}
			newContent.WriteString(newLine + "\n")
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading %s: %w", filePath, err)
		}

		if modified {
			return os.WriteFile(filePath, []byte(newContent.String()), 0644)
		}
		return nil
	})
}

func (r *NewCommand) removeFiles(path string) error {
	// Remove the .git directory
	if err := file.Remove(filepath.Join(path, ".git")); err != nil {
		return err
	}

	// Remove the .GitHub directory
	return file.Remove(filepath.Join(path, ".github"))
}

func (r *NewCommand) copyFile(inputFilePath, outputFilePath string) (err error) {
	// Open .env.example file
	in, err := os.Open(inputFilePath)
	if err != nil {
		return err
	}
	defer in.Close()

	// Create .env file
	out, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy .env.example to .env file
	_, err = io.Copy(out, in)
	return err
}
