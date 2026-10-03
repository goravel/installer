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
)

const agentFileExt = ".md"

type AgentInstallCommand struct{}

func NewAgentInstallCommand() *AgentInstallCommand {
	return &AgentInstallCommand{}
}

// Signature The name and signature of the console command.
func (r *AgentInstallCommand) Signature() string {
	return "agent:install"
}

// Description The console command description.
func (r *AgentInstallCommand) Description() string {
	return "Install Goravel agents for OpenCode"
}

// Extend The console command extend.
func (r *AgentInstallCommand) Extend() command.Extend {
	return command.Extend{
		Flags: []command.Flag{
			&command.StringFlag{
				Name:    "path",
				Aliases: []string{"p"},
				Usage:   "The destination agents folder",
			},
			&command.BoolFlag{
				Name:               "force",
				Aliases:            []string{"f"},
				Usage:              "Overwrite existing agents",
				DisableDefaultText: true,
			},
		},
	}
}

// Handle Execute the console command.
func (r *AgentInstallCommand) Handle(ctx console.Context) error {
	destination, err := r.getDestination(ctx)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	installed, skipped, err := r.installAgents(destination, ctx.OptionBool("force"))
	if err != nil {
		color.Errorln(err)
		return nil
	}

	if installed > 0 {
		color.Successf("Installed %d Goravel agent(s) to %s\n", installed, destination)
	}
	if skipped > 0 {
		color.Warnf("Skipped %d existing Goravel agent(s). Use --force to overwrite.\n", skipped)
	}

	return nil
}

func (r *AgentInstallCommand) getDestination(ctx console.Context) (string, error) {
	destination := ctx.Option("path")
	if destination == "" {
		var err error
		destination, err = opencodeAgentsPath()
		if err != nil {
			return "", err
		}
	}

	// Reuse the shared helper so a Windows "~\..." home path is handled.
	expanded, err := expandHomePath(destination)
	if err != nil {
		return "", err
	}

	destination, err = filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("failed to resolve agents path: %w", err)
	}

	return destination, nil
}

func opencodeAgentsPath() (string, error) {
	configDir, err := opencodeConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "agents"), nil
}

func (r *AgentInstallCommand) installAgents(destination string, force bool) (int, int, error) {
	tmpDir, err := os.MkdirTemp("", "goravel-agents-*")
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	repoPath := filepath.Join(tmpDir, "agents")
	if err := cloneAgents(repoPath); err != nil {
		return 0, 0, err
	}

	agentsPath := filepath.Join(repoPath, "agents")
	files, err := collectAgentFiles(agentsPath)
	if err != nil {
		return 0, 0, err
	}
	if len(files) == 0 {
		return 0, 0, errors.New("no agents found in goravel/agents")
	}

	if err := os.MkdirAll(destination, 0755); err != nil {
		return 0, 0, fmt.Errorf("failed to create agents directory: %w", err)
	}

	var installed, skipped int
	for _, rel := range files {
		wasInstalled, err := r.installAgent(filepath.Join(agentsPath, rel), filepath.Join(destination, rel), force)
		if err != nil {
			return installed, skipped, err
		}

		// Only top-level agents are reported; subagents install silently.
		if filepath.Dir(rel) != "." {
			continue
		}
		if wasInstalled {
			installed++
		} else {
			skipped++
		}
	}

	return installed, skipped, nil
}

// collectAgentFiles returns the paths of every regular ".md" file under
// agentsPath, relative to agentsPath. Subdirectories (such as subagents/) are
// traversed so their agents are always installed; non-".md" entries are ignored
// and non-regular entries are rejected.
func collectAgentFiles(agentsPath string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(agentsPath, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == agentsPath || entry.IsDir() {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported file type %q", path)
		}
		if !strings.HasSuffix(entry.Name(), agentFileExt) {
			return nil
		}

		rel, err := filepath.Rel(agentsPath, path)
		if err != nil {
			return err
		}
		files = append(files, rel)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (r *AgentInstallCommand) installAgent(source, target string, force bool) (bool, error) {
	if file.Exists(target) {
		if !force {
			return false, nil
		}

		if err := os.RemoveAll(target); err != nil {
			return false, fmt.Errorf("failed to remove existing agent %q: %w", filepath.Base(target), err)
		}
	}

	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return false, fmt.Errorf("failed to create agent directory: %w", err)
	}

	info, err := os.Stat(source)
	if err != nil {
		return false, fmt.Errorf("failed to inspect agent %q: %w", filepath.Base(source), err)
	}
	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("unsupported file type %q", source)
	}

	data, err := os.ReadFile(source)
	if err != nil {
		return false, fmt.Errorf("failed to read agent %q: %w", filepath.Base(source), err)
	}
	if err := os.WriteFile(target, data, info.Mode()); err != nil {
		return false, fmt.Errorf("failed to install agent %q: %w", filepath.Base(target), err)
	}

	return true, nil
}
