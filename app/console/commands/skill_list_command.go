package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
)

type SkillListCommand struct{}

type skillDetail struct {
	Name        string
	Description string
}

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
	skills, err := r.fetchSkills(detail)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	color.Green().Printfln("Available Goravel skills:")
	for index, skill := range skills {
		if detail {
			color.Printfln("")
		}

		color.Printfln("%d. %s", index+1, skill.Name)
		if detail && skill.Description != "" {
			color.Printfln("   Description: %s", skill.Description)
		}
	}

	return nil
}

func (r *SkillListCommand) fetchSkills(detail bool) ([]skillDetail, error) {
	tmpDir, err := os.MkdirTemp("", "goravel-agents-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	repoPath := filepath.Join(tmpDir, "agents")
	if err := cloneAgents(repoPath); err != nil {
		return nil, err
	}

	skillsPath := filepath.Join(repoPath, "skills")
	skills, err := listSkillDetails(skillsPath, detail)
	if err != nil {
		return nil, err
	}
	if len(skills) == 0 {
		return nil, errors.New("no skills found in goravel/agents")
	}

	return skills, nil
}

func listSkillDetails(skillsPath string, detail bool) ([]skillDetail, error) {
	skillNames, err := listSkills(skillsPath)
	if err != nil {
		return nil, err
	}

	skills := make([]skillDetail, 0, len(skillNames))
	for _, skillName := range skillNames {
		skill := skillDetail{Name: skillName}
		if detail {
			description, err := readSkillDescription(filepath.Join(skillsPath, skillName, "SKILL.md"))
			if err != nil {
				return nil, fmt.Errorf("failed to read skill %q detail: %w", skillName, err)
			}

			skill.Description = description
		}

		skills = append(skills, skill)
	}

	return skills, nil
}

func readSkillDescription(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return parseSkillDescription(string(content)), nil
}

func parseSkillDescription(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return ""
	}

	var description []string
	collectDescription := false
	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			break
		}
		if collectDescription {
			if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") || trimmed == "" {
				description = append(description, trimmed)
				continue
			}

			break
		}
		if !strings.HasPrefix(trimmed, "description:") {
			continue
		}

		value := strings.TrimSpace(strings.TrimPrefix(trimmed, "description:"))
		if value == "" || strings.HasPrefix(value, ">") || strings.HasPrefix(value, "|") {
			collectDescription = true
			continue
		}

		return strings.Trim(value, `"'`)
	}

	return strings.Join(strings.Fields(strings.Join(description, " ")), " ")
}
