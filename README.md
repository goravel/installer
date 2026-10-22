# Goravel Installer

[![Doc](https://pkg.go.dev/badge/github.com/goravel/installer)](https://pkg.go.dev/github.com/goravel/installer)
[![Go](https://img.shields.io/github/go-mod/go-version/goravel/installer)](https://go.dev/)
[![Release](https://img.shields.io/github/release/goravel/installer.svg)](https://github.com/goravel/installer/releases)
[![Test](https://github.com/goravel/installer/actions/workflows/test.yml/badge.svg)](https://github.com/goravel/installer/actions)
[![Report Card](https://goreportcard.com/badge/github.com/goravel/installer)](https://goreportcard.com/report/github.com/goravel/installer)
[![Codecov](https://codecov.io/gh/goravel/installer/branch/master/graph/badge.svg)](https://codecov.io/gh/goravel/installer)
![License](https://img.shields.io/github/license/goravel/installer)

Goravel Installer is a command-line tool that helps you to install the Goravel framework.

## Installation

```bash
go install github.com/goravel/installer/goravel@latest
```

## Usage

```bash
# Create a new project, the installer asks for anything it needs
goravel new blog

# Set your own module name instead of the default "goravel"
goravel new blog --module github.com/yourname/blog

# Install Goravel Lite, which only includes essential facades,
# then pick the facades you want
goravel new blog --lite

# Don't ask any questions, useful for scripts and CI
goravel new blog -n --module github.com/yourname/blog

# Lite without questions installs no extra facades, add them later
# with ./artisan package:install
goravel new blog --lite -n

# Replace the folder if it already exists
goravel new blog --force

# Install the latest development version
goravel new blog --dev
```

## Skills

```bash
# List available Goravel agent skills
goravel skill:list

# List available Goravel agent skills with descriptions
goravel skill:list --detail

# Install all Goravel agent skills for OpenCode
# Destination (first match): $OPENCODE_CONFIG_DIR/skills, $XDG_CONFIG_HOME/opencode/skills,
#                            ~/.config/opencode/skills
goravel skill:install

# Install all Goravel agent skills to a custom folder
goravel skill:install --path ~/goravel-skills

# Install specific skills
goravel skill:install goravel-testing goravel-planning

# Overwrite existing skills
goravel skill:install --force goravel-testing
```

## Agents

```bash
# Install all Goravel agents (including subagents) for OpenCode
# Destination (first match): $OPENCODE_CONFIG_DIR/agents, $XDG_CONFIG_HOME/opencode/agents,
#                            ~/.config/opencode/agents
goravel agent:install

# Install to a custom folder
goravel agent:install --path ~/goravel-agents

# Overwrite existing agents
goravel agent:install --force
```

## Upgrade

```bash
goravel upgrade

# Upgrade to a specific version
goravel upgrade v1.1.1
```

## License

Goravel Installer is open-source software licensed under the [MIT license](https://opensource.org/licenses/MIT).
