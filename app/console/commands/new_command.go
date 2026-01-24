package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"

	"github.com/goravel/installer/app/facades"
	"github.com/goravel/installer/support"
)

var moduleNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9./_~-]+$`)

type NewCommand struct {
}

func NewNewCommand() *NewCommand {
	return &NewCommand{}
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
				Name:    "module",
				Aliases: []string{"m"},
				Usage:   "Specify the custom module name to replace the default 'goravel' module",
			},
		},
	}
}

// Handle Execute the console command.
func (r *NewCommand) Handle(ctx console.Context) (err error) {
	r.printWelcome(ctx)

	name, err := r.getProjectName(ctx)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	projectType, err := r.getProjectType(ctx)
	if err != nil {
		return fmt.Errorf("failed to get project type: %s", err)
	}

	module, err := r.getModuleName(ctx)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	installLite := projectType == "lite"
	if err = r.generateProject(ctx, name, module, installLite); err != nil {
		color.Errorln(err)
		return nil
	}

	color.Successln("Application ready in [<op=bold>" + name + "</>]. Build something amazing! 🚀🚀")
	color.Successln("Are you new to Goravel? Please visit https://goravel.dev to get started.")

	return
}

func (r *NewCommand) cloneGoravel(repo, path string, dev bool) error {
	args := []string{"clone", "--depth=1", repo, path}
	if dev {
		args = slices.Insert(args, 2, "--branch=master")
	}

	res := facades.Process().Run("git", args...)
	if res.Failed() {
		return fmt.Errorf("failed to clone goravel: %s", res.Error())
	}

	color.Successln("Cloned goravel in " + path)

	return nil
}

func (r *NewCommand) generateProject(ctx console.Context, name, module string, installLite bool) error {
	path := getAbsolutePath(name)

	// remove the directory if it already exists
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove the directory: %s", err)
	}

	repo := "https://github.com/goravel/goravel.git"
	if installLite {
		repo = "https://github.com/goravel/goravel-lite.git"
	}

	if err := r.cloneGoravel(repo, path, ctx.OptionBool("dev")); err != nil {
		return err
	}

	if err := r.replaceModule(ctx, path, module); err != nil {
		return err
	}

	if err := r.initProject(path); err != nil {
		return err
	}

	if installLite {
		if err := r.installFacades(path); err != nil {
			return err
		}
	}

	return nil
}

func (r *NewCommand) getModuleName(ctx console.Context) (string, error) {
	var err error
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
			return "", err
		}
	}
	if !checkModuleName(module) {
		return "", errors.New("invalid module name format. Use only letters, numbers, dots (.), slashes (/), underscores (_), hyphens (-), and tildes (~). Example: [github.com/yourusername/yourproject] or [yourproject]")
	}

	return module, nil
}

func (r *NewCommand) getProjectName(ctx console.Context) (string, error) {
	var err error
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
			return "", err
		}
	}

	if !regexp.MustCompile(`^[\w.-]+$`).MatchString(name) {
		return "", errors.New("the name only supports letters, numbers, dashes, underscores, and periods")
	}

	force := ctx.OptionBool("force")
	if !force && verifyIfDirectoryExists(getAbsolutePath(name)) {
		return "", errors.New("the directory already exists. use the --force flag to overwrite")
	}

	return name, nil
}

func (r *NewCommand) getProjectType(ctx console.Context) (string, error) {
	options := []console.Choice{
		{Key: "Goravel      - Includes all facades", Value: "goravel"},
		{Key: "Goravel Lite - Only includes essential facades", Value: "lite"},
	}

	return ctx.Choice("Which do you want to install?", options)
}

func (r *NewCommand) initProject(path string) error {
	if err := file.Remove(filepath.Join(path, ".git")); err != nil {
		return fmt.Errorf("failed to remove .git: %s", err)
	}

	if err := file.Remove(filepath.Join(path, ".github")); err != nil {
		return fmt.Errorf("failed to remove .github: %s", err)
	}

	if artisan := filepath.Join(path, "artisan"); file.Exists(artisan) {
		if err := os.Chmod(artisan, 0755); err != nil {
			return fmt.Errorf("failed to set artisan execute permission: %s", err)
		}
	}

	if res := facades.Process().WithSpinner("Installing dependencies").Run("go", "mod", "tidy"); res.Failed() {
		return fmt.Errorf("failed to install dependencies: %s", res.Error())
	}

	color.Successln("Installed dependencies")

	if err := file.Copy(filepath.Join(path, ".env.example"), filepath.Join(path, ".env")); err != nil {
		return fmt.Errorf("failed to generate .env file: %s", err)
	}

	color.Successln("Generated .env file")

	if res := facades.Process().WithSpinner("Generating application key").Path(path).Run("go", "run", ".", "artisan", "key:generate"); res.Failed() {
		return fmt.Errorf("failed to generate app key: %s", res.Error())
	}

	return nil
}

func (r *NewCommand) installFacades(path string) error {
	// install facades
	// process := facades.Process()
	if res := facades.Process().TTY().Path(path).Run("go", "run", ".", "artisan", "package:install"); res.Failed() {
		return fmt.Errorf("failed to install facades: %s", res.Error())
	}

	return nil
}

func (r *NewCommand) printWelcome(ctx console.Context) {
	color.Printfln("<fg=52,124,153>%s</>", support.WelcomeHeading) // color hex code: #8ED3F9
	ctx.NewLine()
}

func (r *NewCommand) replaceModule(ctx console.Context, path, module string) error {
	if module == support.DefaultModuleName {
		return nil
	}

	if err := ctx.Spinner("Updating module name to \""+module+"\"", console.SpinnerOption{
		Action: func() error {
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
				defer errors.Ignore(fileContent.Close)

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
		},
	}); err != nil {
		return fmt.Errorf("failed to update module name: %s", err)
	}

	color.Successln("Updated Module name")

	return nil
}

func checkModuleName(module string) bool {
	return moduleNameRegexp.MatchString(module)
}

// getPath Get the full path to the command.
func getAbsolutePath(name string) string {
	pwd, _ := os.Getwd()

	return filepath.Clean(filepath.Join(pwd, name))
}

// verifyIfDirectoryExists Verify if the directory already exists.
func verifyIfDirectoryExists(path string) bool {
	_, err := os.Stat(path)

	return !os.IsNotExist(err)
}
