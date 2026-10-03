package commands

import (
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
)

type SkillListCommand struct{}

func NewSkillListCommand() *SkillListCommand {
	return &SkillListCommand{}
}

// Signature The name and signature of the console command.
func (r *SkillListCommand) Signature() string {
	return "skill:list"
}

// Description The console command description.
func (r *SkillListCommand) Description() string {
	return "List available Goravel agent skills"
}

// Extend The console command extend.
func (r *SkillListCommand) Extend() command.Extend {
	return command.Extend{
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:               "detail",
				Aliases:            []string{"d"},
				Usage:              "Print skill details",
				DisableDefaultText: true,
			},
		},
	}
}

// Handle Execute the console command.
func (r *SkillListCommand) Handle(ctx console.Context) error {
	detail := ctx.OptionBool("detail")
	skills, err := fetchSkills(detail)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	color.Green().Printfln("Available Goravel skills:")
	for index, skill := range skills {
		if detail {
			color.Printfln("")
		}

		color.Printfln("%d. %s", index+1, sanitizeForTerminal(skill.Name /* keepFormatting */, false))
		if detail && skill.Description != "" {
			// Fold to one line so a description cannot forge extra list entries.
			description := strings.Join(strings.Fields(sanitizeForTerminal(skill.Description /* keepFormatting */, true)), " ")
			color.Printfln("   Description: %s", description)
		}
	}

	return nil
}
