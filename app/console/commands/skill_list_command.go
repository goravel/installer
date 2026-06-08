package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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
	return command.Extend{}
}

// Handle Execute the console command.
func (r *SkillListCommand) Handle(ctx console.Context) error {
	skills, err := r.fetchSkills()
	if err != nil {
		color.Errorln(err)
		return nil
	}

	color.Successln("Available Goravel skills:")
	for _, skill := range skills {
		color.Printfln("  - %s", skill)
	}

	return nil
}

func (r *SkillListCommand) fetchSkills() ([]string, error) {
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

	skills, err := listSkills(filepath.Join(repoPath, "skills"))
	if err != nil {
		return nil, err
	}
	if len(skills) == 0 {
		return nil, errors.New("no skills found in goravel/agents")
	}

	return skills, nil
}
