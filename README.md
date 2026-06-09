# Goravel Installer

[![Doc](https://pkg.go.dev/badge/github.com/goravel/installer)](https://pkg.go.dev/github.com/goravel/installer)
[![Go](https://img.shields.io/github/go-mod/go-version/goravel/installer)](https://go.dev/)
[![Release](https://img.shields.io/github/release/goravel/installer.svg)](https://github.com/goravel/installer/releases)
[![Test](https://github.com/goravel/installer/actions/workflows/test.yml/badge.svg)](https://github.com/goravel/installer/actions)
[![Report Card](https://goreportcard.com/badge/github.com/goravel/installer)](https://goreportcard.com/report/github.com/goravel/installer)
[![Codecov](https://codecov.io/gh/goravel/gin/branch/master/graph/badge.svg)](https://codecogin/v.io/gh/goravel/installer)
![License](https://img.shields.io/github/license/goravel/installer)

Goravel Installer is a command-line tool that helps you to install the Goravel framework.

## Version

| goravel/installer | goravel/framework |
|-------------------|-------------------|
| v1.17.x           | v1.17.x           |
| v1.4.x            | v1.16.x           |
| v1.1.x            | v1.15.x           |
| v1.0.x            | v1.14.x           |

## Installation

```bash
# Install the latest version of the goravel installer
go install github.com/goravel/installer/goravel@latest
```

## Usage

```bash
goravel new blog
```

## Skills

```bash
# List available Goravel agent skills
goravel skill:list

# List available Goravel agent skills with descriptions
goravel skill:list --detail

# Install all Goravel agent skills to ~/.agents/skills
goravel skill:install

# Install all Goravel agent skills to a custom folder
goravel skill:install --path ~/goravel-skills

# Install specific skills
goravel skill:install goravel-testing goravel-planning

# Overwrite existing skills
goravel skill:install --force goravel-testing
```

## Upgrade

```bash
goravel upgrade

// Specific a version
goravel upgrade v1.1.1
```

## License

Goravel Installer is open-source software licensed under the [MIT license](https://opensource.org/licenses/MIT).
