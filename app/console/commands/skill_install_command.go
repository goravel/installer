package commands

import (
	"os"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
)

type SkillInstallCommand struct{}

func NewSkillInstallCommand() *SkillInstallCommand {
	return &SkillInstallCommand{}
}

// Signature The name and signature of the console command.
func (r *SkillInstallCommand) Signature() string {
	return "skill:install"
}

// Description The console command description.
func (r *SkillInstallCommand) Description() string {
	return "Install Goravel agent skills for OpenCode"
}

// Extend The console command extend.
func (r *SkillInstallCommand) Extend() command.Extend {
	return command.Extend{
		ArgsUsage: " [skills...]",
		Arguments: []command.Argument{
			&command.ArgumentStringSlice{
				Name:  "skills",
				Usage: "The skills to install. Installs all skills when omitted",
				Min:   0,
				Max:   -1,
			},
		},
		Flags: []command.Flag{
			&command.StringFlag{
				Name:    "path",
				Aliases: []string{"p"},
				Usage:   "The destination skills folder",
			},
			&command.BoolFlag{
				Name:               "force",
				Aliases:            []string{"f"},
				Usage:              "Overwrite existing skills",
				DisableDefaultText: true,
			},
		},
	}
}

// Handle Execute the console command.
func (r *SkillInstallCommand) Handle(ctx console.Context) error {
	openCodeConfig := os.Getenv(envOpenCodeConfigDir)
	xdgConfig := os.Getenv(envXDGConfigHome)

	destination, err := resolveDestination(ctx.Option("path"), openCodeConfig, xdgConfig)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	installed, skipped, err := installSkills(destination, ctx.ArgumentStringSlice("skills"), ctx.OptionBool("force"))
	if err != nil {
		color.Errorln(err)
		return nil
	}

	if installed > 0 {
		color.Successf("Installed %d Goravel skill(s) to %s\n", installed, sanitizeForTerminal(destination /* keepFormatting */, false))
	}
	if skipped > 0 {
		color.Warnf("Skipped %d existing Goravel skill(s). Use --force to overwrite.\n", skipped)
	}

	return nil
}
