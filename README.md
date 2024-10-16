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
| v1.0.x            | v1.14.x           |

## Installation

```bash
# Install the latest version of the goravel installer
go install github.com/goravel/installer@latest

# You can rename the executable file
# Linux / MacOS
mv "$GOBIN/installer" "$GOBIN/goravel"

# Windows
move "%GOBIN%\installer.exe" "%GOBIN%\goravel.exe"
# Windows Powershell
move "$Env:gopath\bin\installer.exe" "$Env:gopath\bin\goravel.exe"
```

## Usage

```bash
goravel new blog
```

## License

Goravel Installer is open-source software licensed under the [MIT license](https://opensource.org/licenses/MIT).
