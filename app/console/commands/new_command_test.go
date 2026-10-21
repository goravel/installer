package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/process"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksprocess "github.com/goravel/framework/mocks/process"
	"github.com/goravel/framework/support/env"
	frameworkmock "github.com/goravel/framework/testing/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	invalidNameError = "the name only supports letters, numbers, dashes, underscores, and periods"
	appKeyPattern    = "[0-9a-zA-Z]{32}"
)

type NewCommandTestSuite struct {
	suite.Suite
	newCommand  *NewCommand
	mockContext *mocksconsole.Context
	mockProcess *mocksprocess.Process
	tmpDir      string
}

func TestNewCommandTestSuite(t *testing.T) {
	suite.Run(t, &NewCommandTestSuite{})
}

func (s *NewCommandTestSuite) SetupTest() {
	t := s.T()
	s.newCommand = NewNewCommand()
	s.mockContext = mocksconsole.NewContext(t)
	s.tmpDir = t.TempDir()
	t.Chdir(s.tmpDir)

	// The process mock has no *testing.T, so its expectations are checked here.
	mockProcess := frameworkmock.Factory().Process()
	s.mockProcess = mockProcess
	t.Cleanup(func() {
		mockProcess.AssertExpectations(t)
	})
}

func (s *NewCommandTestSuite) SetupSubTest() {
	s.SetupTest()
}

type newFlags struct {
	name, module                    string
	lite, dev, force, noInteraction bool
}

func (s *NewCommandTestSuite) expectFlags(flags newFlags) {
	s.mockContext.EXPECT().Argument(0).Return(flags.name).Once()
	s.mockContext.EXPECT().Option("module").Return(flags.module).Once()
	s.mockContext.EXPECT().OptionBool("lite").Return(flags.lite).Once()
	s.mockContext.EXPECT().OptionBool("dev").Return(flags.dev).Once()
	s.mockContext.EXPECT().OptionBool("force").Return(flags.force).Once()
	s.mockContext.EXPECT().OptionBool("no-interaction").Return(flags.noInteraction).Once()
}

func (s *NewCommandTestSuite) newResult(err error) *mocksprocess.Result {
	result := mocksprocess.NewResult(s.T())
	result.EXPECT().Failed().Return(err != nil).Once()
	if err != nil {
		result.EXPECT().Error().Return(err).Once()
	}

	return result
}

func (s *NewCommandTestSuite) expectClone(args []string, files map[string]string, err error) {
	result := s.newResult(err)
	cloneArgs := []any{"clone"}
	for _, arg := range args {
		cloneArgs = append(cloneArgs, arg)
	}
	cloneArgs = append(cloneArgs, mock.Anything)

	s.mockProcess.EXPECT().Run("git", cloneArgs...).RunAndReturn(func(_ string, args ...string) process.Result {
		for name, content := range files {
			s.Require().NoError(os.WriteFile(filepath.Join(args[len(args)-1], name), []byte(content), 0644))
		}

		return result
	}).Once()
}

func (s *NewCommandTestSuite) expectModTidy(err error) {
	s.mockProcess.EXPECT().WithSpinner("Installing dependencies").Return(s.mockProcess).Once()
	s.mockProcess.EXPECT().Path(mock.Anything).Return(s.mockProcess).Once()
	s.mockProcess.EXPECT().Run("go", "mod", "tidy").Return(s.newResult(err)).Once()
}

func (s *NewCommandTestSuite) expectSpinner() {
	s.mockContext.EXPECT().Spinner(mock.Anything, mock.Anything).RunAndReturn(func(_ string, opt console.SpinnerOption) error {
		return opt.Action()
	}).Once()
}

func (s *NewCommandTestSuite) readFile(path string) string {
	content, err := os.ReadFile(path)
	s.Require().NoError(err)

	return string(content)
}

func (s *NewCommandTestSuite) TestHandle() {
	skeleton := map[string]string{
		"go.mod":       "module goravel\n",
		"main.go":      "package main\nimport \"goravel/app\"\n",
		".env.example": "APP_NAME=Goravel\nAPP_KEY=\n",
	}

	tests := []struct {
		name   string
		setup  func()
		assert func(err error)
	}{
		{
			name: "installs goravel from the dev branch with a custom module",
			setup: func() {
				s.expectFlags(newFlags{name: "app", module: "github.com/me/app", dev: true})
				s.mockContext.EXPECT().Choice("Which do you want to install?", mock.Anything).Return("goravel", nil).Once()
				s.expectClone([]string{"--depth=1", "--branch=master", "https://github.com/goravel/goravel.git"}, skeleton, nil)
				s.expectSpinner()
				s.expectModTidy(nil)
			},
			assert: func(err error) {
				s.NoError(err)
				projectPath := filepath.Join(s.tmpDir, "app")
				if !env.IsWindows() {
					info, err := os.Stat(projectPath)
					s.Require().NoError(err)
					s.Equal(os.FileMode(0755), info.Mode().Perm(), "same permissions as git clone")
				}
				s.Contains(s.readFile(filepath.Join(projectPath, "go.mod")), "module github.com/me/app")
				s.Contains(s.readFile(filepath.Join(projectPath, "main.go")), `"github.com/me/app/app"`)
				s.Regexp(`(?m)^APP_KEY=`+appKeyPattern+`$`, s.readFile(filepath.Join(projectPath, ".env")))
			},
		},
		{
			name: "lite runs package:install when interactive",
			setup: func() {
				s.expectFlags(newFlags{name: "app", module: "goravel", lite: true})
				s.expectClone([]string{"--depth=1", "https://github.com/goravel/goravel-lite.git"}, skeleton, nil)
				s.expectModTidy(nil)
				s.mockProcess.EXPECT().TTY().Return(s.mockProcess).Once()
				s.mockProcess.EXPECT().Path(mock.Anything).Return(s.mockProcess).Once()
				s.mockProcess.EXPECT().Run("go", "run", ".", "artisan", "package:install").Return(s.newResult(nil)).Once()
			},
			assert: func(err error) {
				s.NoError(err)
				s.DirExists(filepath.Join(s.tmpDir, "app"))
			},
		},
		{
			// package:install is not expected: it asks questions, so it must be skipped
			name: "lite skips package:install without interaction",
			setup: func() {
				s.expectFlags(newFlags{name: "app", lite: true, noInteraction: true})
				s.expectClone([]string{"--depth=1", "https://github.com/goravel/goravel-lite.git"}, skeleton, nil)
				s.expectModTidy(nil)
			},
			assert: func(err error) {
				s.NoError(err)
				s.FileExists(filepath.Join(s.tmpDir, "app", ".env"))
			},
		},
		{
			name: "failed clone keeps the existing project and leaves nothing behind",
			setup: func() {
				s.Require().NoError(os.Mkdir("app", 0755))
				s.Require().NoError(os.WriteFile(filepath.Join("app", "keep.txt"), []byte("keep"), 0644))
				s.expectFlags(newFlags{name: "app", force: true, noInteraction: true})
				s.expectClone([]string{"--depth=1", "https://github.com/goravel/goravel.git"}, nil, assert.AnError)
			},
			assert: func(err error) {
				s.ErrorIs(err, assert.AnError)
				s.Equal("keep", s.readFile(filepath.Join(s.tmpDir, "app", "keep.txt")))
				entries, err := os.ReadDir(s.tmpDir)
				s.Require().NoError(err)
				s.Len(entries, 1)
			},
		},
		{
			name: "stops before asking anything when a tool is missing",
			setup: func() {
				original := lookPath
				s.T().Cleanup(func() { lookPath = original })
				lookPath = func(file string) (string, error) { return "", assert.AnError }
			},
			assert: func(err error) {
				s.EqualError(err, "git is required but was not found in PATH")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.mockContext.EXPECT().NewLine().Once()
			tt.setup()

			tt.assert(s.newCommand.Handle(s.mockContext))
		})
	}
}

func (s *NewCommandTestSuite) TestGetOptions() {
	tests := []struct {
		name        string
		flags       newFlags
		existing    bool
		choice      string
		expectError string
		expected    newOptions
	}{
		{
			name:        "existing directory without force",
			flags:       newFlags{name: "app"},
			existing:    true,
			expectError: "the directory already exists. use the --force flag to overwrite",
		},
		{
			name:     "existing directory with force",
			flags:    newFlags{name: "app", force: true, noInteraction: true},
			existing: true,
			expected: newOptions{name: "app", module: "goravel", force: true, noInteraction: true},
		},
		{
			name:     "asks for the project type",
			flags:    newFlags{name: "app", module: "github.com/me/app"},
			choice:   "lite",
			expected: newOptions{name: "app", module: "github.com/me/app", lite: true},
		},
		{
			name:     "lite flag skips the project type question",
			flags:    newFlags{name: "app", module: "github.com/me/app", lite: true},
			expected: newOptions{name: "app", module: "github.com/me/app", lite: true},
		},
		{
			name:     "no interaction uses defaults",
			flags:    newFlags{name: "app", noInteraction: true},
			expected: newOptions{name: "app", module: "goravel", noInteraction: true},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.existing {
				s.Require().NoError(os.Mkdir(tt.flags.name, 0755))
			}
			if tt.choice != "" {
				s.mockContext.EXPECT().Choice("Which do you want to install?", mock.Anything).Return(tt.choice, nil).Once()
			}
			s.expectFlags(tt.flags)

			opts, err := s.newCommand.getOptions(s.mockContext)
			if tt.expectError != "" {
				s.EqualError(err, tt.expectError)
				return
			}

			s.NoError(err)
			tt.expected.path = filepath.Join(s.tmpDir, tt.flags.name)
			s.Equal(tt.expected, opts)
		})
	}
}

func (s *NewCommandTestSuite) TestGetProjectName() {
	tests := []struct {
		name          string
		input         string
		noInteraction bool
		expectError   string
	}{
		{name: "dashes", input: "my-project"},
		{name: "underscores", input: "my_project"},
		{name: "periods", input: "my.project"},
		{name: "mixed characters", input: "MyProject_123-v2.0"},
		{name: "special characters", input: "my@project", expectError: invalidNameError},
		{name: "spaces", input: "my project", expectError: invalidNameError},
		{name: "slash", input: "my/project", expectError: invalidNameError},
		{name: "current directory", input: ".", expectError: `the name can't be "." or ".."`},
		{name: "parent directory", input: "..", expectError: `the name can't be "." or ".."`},
		{name: "required without interaction", noInteraction: true, expectError: "the project name is required, e.g. goravel new example-app"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			name, err := s.newCommand.getProjectName(s.mockContext, tt.input, tt.noInteraction)
			if tt.expectError != "" {
				s.EqualError(err, tt.expectError)
				return
			}

			s.NoError(err)
			s.Equal(tt.input, name)
		})
	}
}

func (s *NewCommandTestSuite) TestGetModuleName() {
	tests := []struct {
		name        string
		module      string
		expected    string
		expectError bool
	}{
		{name: "defaults to goravel without interaction", expected: "goravel"},
		{name: "keeps the given module", module: "github.com/me/app", expected: "github.com/me/app"},
		{name: "rejects an invalid module", module: "github.com/me/app!", expectError: true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			module, err := s.newCommand.getModuleName(s.mockContext, tt.module, true)
			if tt.expectError {
				s.ErrorContains(err, "invalid module name format")
				return
			}

			s.NoError(err)
			s.Equal(tt.expected, module)
		})
	}
}

func (s *NewCommandTestSuite) TestCheckModuleName() {
	valid := []string{
		"goravel", "github.com/goravel/framework", "github.com/user/project", "example.com/my-project",
		"example.com/my_project", "gitlab.com/user/project/submodule", "project-name", "project_name",
		"project.name", "example.com/~user/project", "a", "123", "example.com/project-123_test.v2",
	}
	invalid := []string{
		"invalid:module", "module with spaces", "github.com/user/project!", "example.com/project@version",
		"project#name", "project$name", "project%name", "project&name", "project*name", "project(name)",
		"project[name]", "project{name}", "project|name", "project\\name", "project;name", "project'name",
		"project\"name", "project<name>", "project?name",
	}

	for _, module := range valid {
		s.True(checkModuleName(module), "expected %s to be valid", module)
	}
	for _, module := range invalid {
		s.False(checkModuleName(module), "expected %s to be invalid", module)
	}
}

func (s *NewCommandTestSuite) TestInitProject() {
	tests := []struct {
		name        string
		files       map[string]string
		tidyErr     error
		expectError string
	}{
		{
			name: "removes git files, sets artisan permissions and writes .env",
			files: map[string]string{
				".git/HEAD":      "ref",
				".github/ci.yml": "ci",
				"main_test.go":   "package main",
				"artisan":        "#!/bin/bash",
				".env.example":   "APP_NAME=Goravel\nAPP_KEY=\n",
			},
		},
		{
			name:        "fails when go mod tidy fails",
			files:       map[string]string{".env.example": "APP_KEY=\n"},
			tidyErr:     assert.AnError,
			expectError: "failed to install dependencies",
		},
		{
			name:        "fails when .env.example is missing",
			expectError: "failed to generate .env file",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			for name, content := range tt.files {
				s.Require().NoError(os.MkdirAll(filepath.Dir(name), 0755))
				s.Require().NoError(os.WriteFile(name, []byte(content), 0644))
			}
			s.expectModTidy(tt.tidyErr)

			err := s.newCommand.initProject(s.tmpDir)
			if tt.expectError != "" {
				s.ErrorContains(err, tt.expectError)
				return
			}

			s.NoError(err)
			s.NoDirExists(".git")
			s.NoDirExists(".github")
			s.NoFileExists("main_test.go")
			s.Regexp(`^APP_NAME=Goravel\nAPP_KEY=`+appKeyPattern+`\n$`, s.readFile(".env"))

			info, err := os.Stat("artisan")
			s.Require().NoError(err)
			// Windows has no execute bit, so the mode stays 0666 there
			if env.IsWindows() {
				s.Equal(os.FileMode(0666), info.Mode().Perm())
			} else {
				s.Equal(os.FileMode(0755), info.Mode().Perm())
			}
		})
	}
}

func (s *NewCommandTestSuite) TestReplaceModule() {
	// A line longer than bufio.Scanner's 64 KB limit
	longLine := "// " + strings.Repeat("x", 70*1024)

	tests := []struct {
		name     string
		module   string
		files    map[string]string
		expected map[string]string
	}{
		{
			name:   "replaces the module in go.mod and imports only",
			module: "github.com/me/app",
			files: map[string]string{
				"go.mod":     "module goravel\nrequire example.com v1.0.0",
				"main.go":    "package main\nimport \"goravel/utils\"",
				"ignore.txt": "module goravel",
			},
			expected: map[string]string{
				"go.mod":     "module github.com/me/app\nrequire example.com v1.0.0",
				"main.go":    "package main\nimport \"github.com/me/app/utils\"",
				"ignore.txt": "module goravel",
			},
		},
		{
			name:   "keeps line endings and long lines",
			module: "github.com/me/app",
			files: map[string]string{
				"go.mod":  "module goravel\r\n\r\ngo 1.25.0\r\n",
				"main.go": "package main\r\nimport \"goravel/app\"\r\n" + longLine + "\r\n",
			},
			expected: map[string]string{
				"go.mod":  "module github.com/me/app\r\n\r\ngo 1.25.0\r\n",
				"main.go": "package main\r\nimport \"github.com/me/app/app\"\r\n" + longLine + "\r\n",
			},
		},
		{
			name:     "does nothing for the default module",
			module:   "goravel",
			files:    map[string]string{"go.mod": "module goravel\n"},
			expected: map[string]string{"go.mod": "module goravel\n"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			for name, content := range tt.files {
				s.Require().NoError(os.WriteFile(name, []byte(content), 0644))
			}
			if tt.module != "goravel" {
				s.expectSpinner()
			}

			s.NoError(s.newCommand.replaceModule(s.mockContext, s.tmpDir, tt.module))
			for name, content := range tt.expected {
				s.Equal(content, s.readFile(name), name)
			}
		})
	}
}

func (s *NewCommandTestSuite) TestWriteEnvFile() {
	tests := []struct {
		name     string
		example  string
		expected string
	}{
		{name: "replaces APP_KEY and keeps the rest", example: "APP_NAME=Goravel\r\nAPP_KEY=\r\nAPP_DEBUG=true\r\n", expected: `^APP_NAME=Goravel\r\nAPP_KEY=` + appKeyPattern + `\r\nAPP_DEBUG=true\r\n$`},
		{name: "adds APP_KEY with the file's own line ending", example: "APP_NAME=Goravel\r\n", expected: `^APP_NAME=Goravel\r\nAPP_KEY=` + appKeyPattern + `\r\n$`},
		{name: "adds APP_KEY when it is missing", example: "APP_NAME=Goravel", expected: `^APP_NAME=Goravel\nAPP_KEY=` + appKeyPattern + `\n$`},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Require().NoError(os.WriteFile(".env.example", []byte(tt.example), 0644))

			s.NoError(writeEnvFile(s.tmpDir))
			s.Regexp(tt.expected, s.readFile(".env"))
		})
	}
}

func (s *NewCommandTestSuite) TestCheckRequiredTools() {
	tests := []struct {
		name        string
		missing     string
		expectError string
	}{
		{name: "all tools found"},
		{name: "git missing", missing: "git", expectError: "git is required but was not found in PATH"},
		{name: "go missing", missing: "go", expectError: "go is required but was not found in PATH"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			original := lookPath
			s.T().Cleanup(func() { lookPath = original })
			lookPath = func(file string) (string, error) {
				if file == tt.missing {
					return "", assert.AnError
				}
				return "/usr/bin/" + file, nil
			}

			err := checkRequiredTools()
			if tt.expectError != "" {
				s.EqualError(err, tt.expectError)
				return
			}

			s.NoError(err)
		})
	}
}

func (s *NewCommandTestSuite) TestMoveIntoPlace() {
	tests := []struct {
		name            string
		srcExists       bool
		dstExists       bool
		force           bool
		expectError     string
		expectedFile    string
		expectedEntries []string
	}{
		{name: "moves to a new directory", srcExists: true, expectedFile: "new.txt", expectedEntries: []string{"app"}},
		{name: "replaces an existing directory with force", srcExists: true, dstExists: true, force: true, expectedFile: "new.txt", expectedEntries: []string{"app"}},
		{name: "refuses an existing directory without force", srcExists: true, dstExists: true, expectError: "the directory already exists. use the --force flag to overwrite", expectedFile: "old.txt", expectedEntries: []string{"app", "src"}},
		// src doesn't exist, so the move fails after dst was moved aside
		{name: "restores the existing directory when the move fails", dstExists: true, force: true, expectError: "failed to move the project into place", expectedFile: "old.txt", expectedEntries: []string{"app"}},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			for dir, exists := range map[string]bool{"src": tt.srcExists, "app": tt.dstExists} {
				if exists {
					file := map[string]string{"src": "new.txt", "app": "old.txt"}[dir]
					s.Require().NoError(os.Mkdir(dir, 0755))
					s.Require().NoError(os.WriteFile(filepath.Join(dir, file), nil, 0644))
				}
			}

			err := moveIntoPlace(filepath.Join(s.tmpDir, "src"), filepath.Join(s.tmpDir, "app"), tt.force)
			if tt.expectError != "" {
				s.ErrorContains(err, tt.expectError)
			} else {
				s.NoError(err)
			}

			s.FileExists(filepath.Join("app", tt.expectedFile))
			entries, err := os.ReadDir(s.tmpDir)
			s.Require().NoError(err)
			var names []string
			for _, entry := range entries {
				names = append(names, entry.Name())
			}
			s.Equal(tt.expectedEntries, names)
		})
	}
}
