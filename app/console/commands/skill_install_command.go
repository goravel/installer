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
	"github.com/goravel/framework/support/file"

	"github.com/goravel/installer/app/facades"
)

const agentsRepo = "https://github.com/goravel/agents.git"

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
	return "Install Goravel agent skills"
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
	destination, err := r.getDestination(ctx)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	installed, skipped, err := r.installSkills(destination, ctx.ArgumentStringSlice("skills"), ctx.OptionBool("force"))
	if err != nil {
		color.Errorln(err)
		return nil
	}

	if installed > 0 {
		color.Successf("Installed %d Goravel skill(s) to %s\n", installed, destination)
	}
	if skipped > 0 {
		color.Warnf("Skipped %d existing Goravel skill(s). Use --force to overwrite.\n", skipped)
	}

	return nil
}

func (r *SkillInstallCommand) getDestination(ctx console.Context) (string, error) {
	destination := ctx.Option("path")
	if destination == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}

		destination = filepath.Join(home, ".agents", "skills")
	} else {
		expanded, err := expandHomePath(destination)
		if err != nil {
			return "", err
		}
		destination = expanded
	}

	destination, err := filepath.Abs(destination)
	if err != nil {
		return "", fmt.Errorf("failed to resolve skills path: %w", err)
	}

	return destination, nil
}

func (r *SkillInstallCommand) installSkills(destination string, skillNames []string, force bool) (int, int, error) {
	tmpDir, err := os.MkdirTemp("", "goravel-agents-*")
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	repoPath := filepath.Join(tmpDir, "agents")
	if err := r.cloneAgents(repoPath); err != nil {
		return 0, 0, err
	}

	skillsPath := filepath.Join(repoPath, "skills")
	skills, err := r.resolveSkills(skillsPath, skillNames)
	if err != nil {
		return 0, 0, err
	}
	if len(skills) == 0 {
		return 0, 0, errors.New("no skills found in goravel/agents")
	}

	if err := os.MkdirAll(destination, 0755); err != nil {
		return 0, 0, fmt.Errorf("failed to create skills directory: %w", err)
	}

	var installed, skipped int
	for _, skill := range skills {
		wasInstalled, err := r.installSkill(skillsPath, destination, skill, force)
		if err != nil {
			return installed, skipped, err
		}
		if wasInstalled {
			installed++
		} else {
			skipped++
		}
	}

	return installed, skipped, nil
}

func (r *SkillInstallCommand) cloneAgents(path string) error {
	res := facades.Process().WithSpinner("Downloading Goravel agents").Run("git", "clone", "--depth=1", agentsRepo, path)
	if res.Failed() {
		return fmt.Errorf("failed to clone goravel agents: %v", res.Error())
	}

	return nil
}

func (r *SkillInstallCommand) resolveSkills(skillsPath string, skillNames []string) ([]string, error) {
	skillNames, err := normalizeSkillNames(skillNames)
	if err != nil {
		return nil, err
	}

	if len(skillNames) == 0 {
		return listSkills(skillsPath)
	}

	for _, skill := range skillNames {
		info, err := os.Stat(filepath.Join(skillsPath, skill))
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("skill %q does not exist", skill)
			}

			return nil, fmt.Errorf("failed to inspect skill %q: %w", skill, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("skill %q is not a directory", skill)
		}
	}

	return skillNames, nil
}

func (r *SkillInstallCommand) installSkill(skillsPath, destination, skill string, force bool) (bool, error) {
	source := filepath.Join(skillsPath, skill)
	target := filepath.Join(destination, skill)

	if file.Exists(target) {
		if !force {
			return false, nil
		}

		if err := os.RemoveAll(target); err != nil {
			return false, fmt.Errorf("failed to remove existing skill %q: %w", skill, err)
		}
	}

	if err := copyDirectory(source, target); err != nil {
		return false, fmt.Errorf("failed to install skill %q: %w", skill, err)
	}

	return true, nil
}

func normalizeSkillNames(skillNames []string) ([]string, error) {
	seen := make(map[string]bool, len(skillNames))
	unique := make([]string, 0, len(skillNames))

	for _, skill := range skillNames {
		skill = strings.TrimSpace(skill)
		if skill == "" {
			continue
		}
		if skill == "." || skill == ".." || strings.ContainsAny(skill, `/\\`) {
			return nil, fmt.Errorf("invalid skill name %q", skill)
		}
		if seen[skill] {
			continue
		}

		seen[skill] = true
		unique = append(unique, skill)
	}

	return unique, nil
}

func listSkills(skillsPath string) ([]string, error) {
	entries, err := os.ReadDir(skillsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read skills: %w", err)
	}

	skills := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			skills = append(skills, entry.Name())
		}
	}

	return skills, nil
}

func copyDirectory(source, target string) error {
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(target, relativePath)

		if entry.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported file type %q", path)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(targetPath, data, info.Mode())
	})
}

func expandHomePath(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") && !strings.HasPrefix(path, `~\`) {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	if path == "~" {
		return home, nil
	}

	return filepath.Join(home, path[2:]), nil
}
