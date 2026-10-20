package commands

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"

	"github.com/goravel/installer/app/facades"
	"github.com/goravel/installer/support"
)

var (
	moduleNameRegexp  = regexp.MustCompile(`^[a-zA-Z0-9./_~-]+$`)
	projectNameRegexp = regexp.MustCompile(`^[\w.-]+$`)
	goModModuleRegexp = regexp.MustCompile(`^module\s+goravel\b`)
	goImportRegexp    = regexp.MustCompile(`"goravel/([^"]+)"`)
	appKeyRegexp      = regexp.MustCompile(`(?m)^APP_KEY=[^\r\n]*`)

	// lookPath is replaced in tests to simulate a missing tool.
	lookPath = exec.LookPath
)

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
			&command.BoolFlag{
				Name:               "lite",
				Usage:              "Install Goravel Lite, which only includes essential facades",
				DisableDefaultText: true,
			},
			&command.BoolFlag{
				Name:               "no-interaction",
				Aliases:            []string{"n"},
				Usage:              "Don't ask any questions, use the flags and defaults instead",
				DisableDefaultText: true,
			},
		},
	}
}

// Handle Execute the console command.
func (r *NewCommand) Handle(ctx console.Context) error {
	r.printWelcome(ctx)

	if err := checkRequiredTools(); err != nil {
		return err
	}

	noInteraction := ctx.OptionBool("no-interaction")

	name, err := r.getProjectName(ctx, noInteraction)
	if err != nil {
		return err
	}

	projectType, err := r.getProjectType(ctx, noInteraction)
	if err != nil {
		return fmt.Errorf("failed to get project type: %w", err)
	}

	module, err := r.getModuleName(ctx, noInteraction)
	if err != nil {
		return err
	}

	installLite := projectType == "lite"
	if err := r.generateProject(ctx, name, module, installLite, noInteraction); err != nil {
		return err
	}

	color.Successln("Application ready in [<op=bold>" + name + "</>]. Build something amazing! 🚀🚀")
	color.Successln("Are you new to Goravel? Please visit https://goravel.dev to get started.")

	return nil
}

func (r *NewCommand) cloneGoravel(repo, path string, dev bool) error {
	args := []string{"clone", "--depth=1", repo, path}
	if dev {
		args = slices.Insert(args, 2, "--branch=master")
	}

	res := facades.Process().Run("git", args...)
	if res.Failed() {
		return fmt.Errorf("failed to clone goravel: %w", res.Error())
	}

	color.Successln("Cloned goravel")

	return nil
}

func (r *NewCommand) generateProject(ctx console.Context, name, module string, installLite, noInteraction bool) error {
	path, err := getAbsolutePath(name)
	if err != nil {
		return err
	}

	// Build the project in a temp directory next to the target, so a failed
	// install never leaves a half-built project or touches an existing one.
	tmpPath, err := os.MkdirTemp(filepath.Dir(path), "."+name+"-*")
	if err != nil {
		return fmt.Errorf("failed to create a temp directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpPath)
	}()

	repo := "https://github.com/goravel/goravel.git"
	if installLite {
		repo = "https://github.com/goravel/goravel-lite.git"
	}

	if err := r.cloneGoravel(repo, tmpPath, ctx.OptionBool("dev")); err != nil {
		return err
	}

	if err := r.replaceModule(ctx, tmpPath, module); err != nil {
		return err
	}

	if err := r.initProject(tmpPath); err != nil {
		return err
	}

	// package:install asks which facades to install, so it is skipped when
	// questions are turned off.
	if installLite && !noInteraction {
		if err := r.installFacades(tmpPath); err != nil {
			return err
		}
	}

	if err := moveIntoPlace(tmpPath, path); err != nil {
		return err
	}

	if installLite && noInteraction {
		color.Infoln("No facades were installed. Run [<op=bold>./artisan package:install</>] in the project to add them.")
	}

	return nil
}

func (r *NewCommand) getModuleName(ctx console.Context, noInteraction bool) (string, error) {
	var err error
	module := ctx.Option("module")

	if module == "" && noInteraction {
		module = support.DefaultModuleName
	}

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

func (r *NewCommand) getProjectName(ctx console.Context, noInteraction bool) (string, error) {
	var err error
	name := ctx.Argument(0)

	if name == "" && noInteraction {
		return "", errors.New("the project name is required, e.g. goravel new example-app")
	}

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

	if !projectNameRegexp.MatchString(name) {
		return "", errors.New("the name only supports letters, numbers, dashes, underscores, and periods")
	}

	// "." and ".." pass the check above but point at the current or parent
	// directory, which --force would then replace.
	if name == "." || name == ".." {
		return "", errors.New(`the name can't be "." or ".."`)
	}

	path, err := getAbsolutePath(name)
	if err != nil {
		return "", err
	}

	force := ctx.OptionBool("force")
	if !force && verifyIfDirectoryExists(path) {
		return "", errors.New("the directory already exists. use the --force flag to overwrite")
	}

	return name, nil
}

func (r *NewCommand) getProjectType(ctx console.Context, noInteraction bool) (string, error) {
	if ctx.OptionBool("lite") {
		return "lite", nil
	}

	if noInteraction {
		return "goravel", nil
	}

	options := []console.Choice{
		{Key: "Goravel      - Includes all facades", Value: "goravel"},
		{Key: "Goravel Lite - Only includes essential facades", Value: "lite"},
	}

	return ctx.Choice("Which do you want to install?", options)
}

func (r *NewCommand) initProject(path string) error {
	if err := file.Remove(filepath.Join(path, ".git")); err != nil {
		return fmt.Errorf("failed to remove .git: %w", err)
	}

	if err := file.Remove(filepath.Join(path, ".github")); err != nil {
		return fmt.Errorf("failed to remove .github: %w", err)
	}

	if err := file.Remove(filepath.Join(path, "main_test.go")); err != nil {
		return fmt.Errorf("failed to remove main_test.go: %w", err)
	}

	if artisan := filepath.Join(path, "artisan"); file.Exists(artisan) {
		if err := os.Chmod(artisan, 0755); err != nil {
			return fmt.Errorf("failed to set artisan execute permission: %w", err)
		}
	}

	if res := facades.Process().WithSpinner("Installing dependencies").Path(path).Run("go", "mod", "tidy"); res.Failed() {
		return fmt.Errorf("failed to install dependencies: %w", res.Error())
	}

	color.Successln("Installed dependencies")

	if err := writeEnvFile(path); err != nil {
		return fmt.Errorf("failed to generate .env file: %w", err)
	}

	color.Successln("Generated .env file with an application key")

	return nil
}

// writeEnvFile creates .env from .env.example with a new APP_KEY. It sets the
// key directly instead of running key:generate, which would build the whole
// project first and doesn't report its own errors.
func writeEnvFile(path string) error {
	content, err := os.ReadFile(filepath.Join(path, ".env.example"))
	if err != nil {
		return err
	}

	// str.Random is what key:generate uses, so the key looks the same.
	appKey := "APP_KEY=" + str.Random(32)
	env := string(content)
	if appKeyRegexp.MatchString(env) {
		env = appKeyRegexp.ReplaceAllLiteralString(env, appKey)
	} else {
		env = strings.TrimRight(env, "\n") + "\n" + appKey + "\n"
	}

	return os.WriteFile(filepath.Join(path, ".env"), []byte(env), 0644)
}

func (r *NewCommand) installFacades(path string) error {
	if res := facades.Process().TTY().Path(path).Run("go", "run", ".", "artisan", "package:install"); res.Failed() {
		return fmt.Errorf("failed to install facades: %w", res.Error())
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

			return filepath.WalkDir(path, func(filePath string, entry fs.DirEntry, err error) error {
				if err != nil || entry.IsDir() || (!strings.HasSuffix(filePath, ".go") && !strings.HasSuffix(filePath, ".mod")) {
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
						newLine = goModModuleRegexp.ReplaceAllString(line, "module "+module)
					} else {
						newLine = goImportRegexp.ReplaceAllString(line, `"`+module+`/$1"`)
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
		return fmt.Errorf("failed to update module name: %w", err)
	}

	color.Successln("Updated Module name")

	return nil
}

func checkModuleName(module string) bool {
	return moduleNameRegexp.MatchString(module)
}

// checkRequiredTools checks that git and go are installed before anything runs.
func checkRequiredTools() error {
	for _, tool := range []string{"git", "go"} {
		if _, err := lookPath(tool); err != nil {
			return fmt.Errorf("%s is required but was not found in PATH", tool)
		}
	}

	return nil
}

// getAbsolutePath returns the absolute path of the project directory.
func getAbsolutePath(name string) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get the current directory: %w", err)
	}

	return filepath.Clean(filepath.Join(pwd, name)), nil
}

// moveIntoPlace moves the finished project from src to dst. An existing dst
// (only possible with --force) is moved aside first and is put back if the
// move fails, so it is never lost.
func moveIntoPlace(src, dst string) error {
	if !verifyIfDirectoryExists(dst) {
		if err := os.Rename(src, dst); err != nil {
			return fmt.Errorf("failed to move the project into place: %w", err)
		}

		return nil
	}

	backupDir, err := os.MkdirTemp(filepath.Dir(dst), "."+filepath.Base(dst)+"-old-*")
	if err != nil {
		return fmt.Errorf("failed to create a backup directory: %w", err)
	}

	backup := filepath.Join(backupDir, filepath.Base(dst))
	if err := os.Rename(dst, backup); err != nil {
		_ = os.RemoveAll(backupDir)
		return fmt.Errorf("failed to move the existing directory aside: %w", err)
	}

	if err := os.Rename(src, dst); err != nil {
		if restoreErr := os.Rename(backup, dst); restoreErr != nil {
			return fmt.Errorf("failed to move the project into place: %w. The existing directory was kept at %s", err, backup)
		}
		_ = os.RemoveAll(backupDir)

		return fmt.Errorf("failed to move the project into place: %w", err)
	}

	if err := os.RemoveAll(backupDir); err != nil {
		color.Warningln("Failed to remove the old directory at " + backup + ": " + err.Error())
	}

	return nil
}

// verifyIfDirectoryExists Verify if the directory already exists.
func verifyIfDirectoryExists(path string) bool {
	_, err := os.Stat(path)

	return !errors.Is(err, fs.ErrNotExist)
}
