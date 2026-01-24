package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/process"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksprocess "github.com/goravel/framework/mocks/process"
	"github.com/goravel/framework/support/env"
	frameworkmock "github.com/goravel/framework/testing/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckModuleName(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		validModules := []string{
			"goravel",
			"github.com/goravel/framework",
			"github.com/user/project",
			"example.com/my-project",
			"example.com/my_project",
			"gitlab.com/user/project/submodule",
			"project-name",
			"project_name",
			"project.name",
			"example.com/~user/project",
			"a",
			"123",
			"example.com/project-123_test.v2",
		}

		for _, module := range validModules {
			t.Run("valid_"+module, func(t *testing.T) {
				assert.True(t, checkModuleName(module), "Expected %s to be valid", module)
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		invalidModules := []string{
			"invalid:module",
			"module with spaces",
			"github.com/user/project!",
			"example.com/project@version",
			"project#name",
			"project$name",
			"project%name",
			"project&name",
			"project*name",
			"project(name)",
			"project[name]",
			"project{name}",
			"project|name",
			"project\\name",
			"project;name",
			"project'name",
			"project\"name",
			"project<name>",
			"project?name",
		}

		for _, module := range invalidModules {
			t.Run("invalid_"+module, func(t *testing.T) {
				assert.False(t, checkModuleName(module), "Expected %s to be invalid", module)
			})
		}
	})
}

func TestGetProjectName(t *testing.T) {
	newCommand := &NewCommand{}

	t.Run("valid project name provided", func(t *testing.T) {
		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Argument(0).Return("my-project").Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()

		name, err := newCommand.getProjectName(mockContext)
		assert.Nil(t, err)
		assert.Equal(t, "my-project", name)
	})

	t.Run("valid project name with underscores", func(t *testing.T) {
		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Argument(0).Return("my_project").Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()

		name, err := newCommand.getProjectName(mockContext)
		assert.Nil(t, err)
		assert.Equal(t, "my_project", name)
	})

	t.Run("valid project name with periods", func(t *testing.T) {
		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Argument(0).Return("my.project").Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()

		name, err := newCommand.getProjectName(mockContext)
		assert.Nil(t, err)
		assert.Equal(t, "my.project", name)
	})

	t.Run("valid project name with mixed characters", func(t *testing.T) {
		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Argument(0).Return("MyProject_123-v2.0").Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()

		name, err := newCommand.getProjectName(mockContext)
		assert.Nil(t, err)
		assert.Equal(t, "MyProject_123-v2.0", name)
	})

	t.Run("invalid project name with special characters", func(t *testing.T) {
		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Argument(0).Return("my@project").Once()

		name, err := newCommand.getProjectName(mockContext)
		assert.NotNil(t, err)
		assert.Equal(t, "the name only supports letters, numbers, dashes, underscores, and periods", err.Error())
		assert.Equal(t, "", name)
	})

	t.Run("invalid project name with spaces", func(t *testing.T) {
		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Argument(0).Return("my project").Once()

		name, err := newCommand.getProjectName(mockContext)
		assert.NotNil(t, err)
		assert.Equal(t, "the name only supports letters, numbers, dashes, underscores, and periods", err.Error())
		assert.Equal(t, "", name)
	})

	t.Run("invalid project name with slash", func(t *testing.T) {
		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Argument(0).Return("my/project").Once()

		name, err := newCommand.getProjectName(mockContext)
		assert.NotNil(t, err)
		assert.Equal(t, "the name only supports letters, numbers, dashes, underscores, and periods", err.Error())
		assert.Equal(t, "", name)
	})

	t.Run("directory already exists without force flag", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-project-exists")
		assert.Nil(t, err)
		defer func() {
			_ = os.RemoveAll(tmpDir)
		}()

		projectName := filepath.Base(tmpDir)
		currentDir, err := os.Getwd()
		assert.Nil(t, err)

		// Change to the parent directory temporarily
		parentDir := filepath.Dir(tmpDir)
		err = os.Chdir(parentDir)
		assert.Nil(t, err)
		defer func() {
			_ = os.Chdir(currentDir)
		}()

		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Argument(0).Return(projectName).Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()

		name, err := newCommand.getProjectName(mockContext)
		assert.NotNil(t, err)
		assert.Equal(t, "the directory already exists. use the --force flag to overwrite", err.Error())
		assert.Equal(t, "", name)
	})

	t.Run("directory already exists with force flag", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-project-force")
		assert.Nil(t, err)
		defer func() {
			_ = os.RemoveAll(tmpDir)
		}()

		projectName := filepath.Base(tmpDir)
		currentDir, err := os.Getwd()
		assert.Nil(t, err)

		// Change to the parent directory temporarily
		parentDir := filepath.Dir(tmpDir)
		err = os.Chdir(parentDir)
		assert.Nil(t, err)
		defer func() {
			_ = os.Chdir(currentDir)
		}()

		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Argument(0).Return(projectName).Once()
		mockContext.EXPECT().OptionBool("force").Return(true).Once()

		name, err := newCommand.getProjectName(mockContext)
		assert.Nil(t, err)
		assert.Equal(t, projectName, name)
	})
}

func TestHandle(t *testing.T) {
	newCommand := &NewCommand{}

	mockFactory := frameworkmock.Factory()
	mockProcess := mockFactory.Process()

	tmpDir, err := os.MkdirTemp("", "test-handle-happy")
	assert.Nil(t, err)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	currentDir, err := os.Getwd()
	assert.Nil(t, err)
	err = os.Chdir(tmpDir)
	assert.Nil(t, err)
	defer func() {
		_ = os.Chdir(currentDir)
	}()

	projectName := "test-project"
	moduleName := "github.com/test/project"
	projectPath := filepath.Join(tmpDir, projectName)

	// Create mock context
	mockContext := mocksconsole.NewContext(t)

	// Mock printWelcome (NewLine call)
	mockContext.EXPECT().NewLine().Once()

	// Mock getProjectName
	mockContext.EXPECT().Argument(0).Return(projectName).Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()

	// Mock getProjectType
	mockContext.EXPECT().Choice("Which do you want to install?", mock.MatchedBy(func(choices []console.Choice) bool {
		return len(choices) == 2
	})).Return("goravel", nil).Once()

	// Mock getModuleName
	mockContext.EXPECT().Option("module").Return(moduleName).Once()

	// Mock generateProject - cloneGoravel
	mockContext.EXPECT().OptionBool("dev").Return(false).Once()
	mockCloneResult := mocksprocess.NewResult(t)
	mockCloneResult.EXPECT().Failed().Return(false).Once()
	mockProcess.EXPECT().Run("git", "clone", "--depth=1", "https://github.com/goravel/goravel.git", mock.Anything).RunAndReturn(func(command string, args ...string) process.Result {
		// Simulate git clone by creating the project directory structure
		err := os.MkdirAll(projectPath, 0755)
		assert.Nil(t, err)

		// Create go.mod file
		modFile := filepath.Join(projectPath, "go.mod")
		err = os.WriteFile(modFile, []byte("module goravel\n"), 0644)
		assert.Nil(t, err)

		// Create .env.example file
		envExample := filepath.Join(projectPath, ".env.example")
		err = os.WriteFile(envExample, []byte("APP_NAME=TestApp\n"), 0644)
		assert.Nil(t, err)

		// Create a simple go file to test module replacement
		mainFile := filepath.Join(projectPath, "main.go")
		err = os.WriteFile(mainFile, []byte(`package main
import "goravel/app"
func main() {}`), 0644)
		assert.Nil(t, err)

		return mockCloneResult
	}).Once()

	// Mock replaceModule
	mockContext.EXPECT().Spinner("Updating module name to \""+moduleName+"\"", mock.MatchedBy(func(opt console.SpinnerOption) bool {
		return opt.Action != nil
	})).RunAndReturn(func(_ string, opt console.SpinnerOption) error {
		// Execute the actual action
		return opt.Action()
	}).Once()

	// Mock initProject - mod tidy
	mockProcess.EXPECT().WithSpinner("Installing dependencies").Return(mockProcess).Once()
	mockProcess.EXPECT().Path(mock.Anything).Return(mockProcess).Once()
	mockModTidyResult := mocksprocess.NewResult(t)
	mockModTidyResult.EXPECT().Failed().Return(false).Once()
	mockProcess.EXPECT().Run("go", "mod", "tidy").Return(mockModTidyResult).Once()

	// Mock initProject - key:generate
	mockProcess.EXPECT().WithSpinner("Generating application key").Return(mockProcess).Once()
	mockProcess.EXPECT().Path(mock.Anything).Return(mockProcess).Once()
	mockKeyGenResult := mocksprocess.NewResult(t)
	mockKeyGenResult.EXPECT().Failed().Return(false).Once()
	mockProcess.EXPECT().Run("go", "run", ".", "artisan", "key:generate").Return(mockKeyGenResult).Once()

	// Execute the Handle function
	err = newCommand.Handle(mockContext)
	assert.Nil(t, err)

	// Verify the project was created
	assert.DirExists(t, projectPath)

	// Verify .env was created
	envFile := filepath.Join(projectPath, ".env")
	assert.FileExists(t, envFile)

	// Verify module was replaced in go.mod
	modFile := filepath.Join(projectPath, "go.mod")
	modContent, err := os.ReadFile(modFile)
	assert.Nil(t, err)
	assert.Contains(t, string(modContent), "module "+moduleName)

	// Verify module was replaced in go files
	mainFile := filepath.Join(projectPath, "main.go")
	mainContent, err := os.ReadFile(mainFile)
	assert.Nil(t, err)
	assert.Contains(t, string(mainContent), `"`+moduleName+`/app"`)
}

func TestInitProject(t *testing.T) {
	newCommand := &NewCommand{}

	t.Run("successfully initializes project", func(t *testing.T) {
		mockFactory := frameworkmock.Factory()
		mockProcess := mockFactory.Process()

		tmpDir, err := os.MkdirTemp("", "test-init-project")
		assert.Nil(t, err)
		defer func() {
			_ = os.RemoveAll(tmpDir)
		}()

		// Create .git and .github directories
		gitDir := filepath.Join(tmpDir, ".git")
		githubDir := filepath.Join(tmpDir, ".github")
		err = os.MkdirAll(gitDir, 0755)
		assert.Nil(t, err)
		err = os.MkdirAll(githubDir, 0755)
		assert.Nil(t, err)

		// Create a file inside .git
		testFile := filepath.Join(gitDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test"), 0644)
		assert.Nil(t, err)

		// Create .env.example
		envExample := filepath.Join(tmpDir, ".env.example")
		err = os.WriteFile(envExample, []byte("APP_NAME=TestApp\nAPP_KEY="), 0644)
		assert.Nil(t, err)

		// Mock go mod tidy
		mockProcess.EXPECT().WithSpinner("Installing dependencies").Return(mockProcess).Once()
		mockProcess.EXPECT().Path(tmpDir).Return(mockProcess).Once()
		mockModTidyResult := mocksprocess.NewResult(t)
		mockModTidyResult.EXPECT().Failed().Return(false).Once()
		mockProcess.EXPECT().Run("go", "mod", "tidy").Return(mockModTidyResult).Once()

		// Mock key:generate
		mockProcess.EXPECT().WithSpinner("Generating application key").Return(mockProcess).Once()
		mockProcess.EXPECT().Path(tmpDir).Return(mockProcess).Once()
		mockKeyGenResult := mocksprocess.NewResult(t)
		mockKeyGenResult.EXPECT().Failed().Return(false).Once()
		mockProcess.EXPECT().Run("go", "run", ".", "artisan", "key:generate").Return(mockKeyGenResult).Once()

		err = newCommand.initProject(tmpDir)
		assert.Nil(t, err)

		// Verify .git and .github were removed
		assert.NoDirExists(t, gitDir)
		assert.NoDirExists(t, githubDir)

		// Verify .env was created
		envFile := filepath.Join(tmpDir, ".env")
		assert.FileExists(t, envFile)
		envContent, err := os.ReadFile(envFile)
		assert.Nil(t, err)
		assert.Contains(t, string(envContent), "APP_NAME=TestApp")
	})

	t.Run("fails when go mod tidy fails", func(t *testing.T) {
		mockFactory := frameworkmock.Factory()
		mockProcess := mockFactory.Process()

		tmpDir, err := os.MkdirTemp("", "test-init-project-fail")
		assert.Nil(t, err)
		defer func() {
			_ = os.RemoveAll(tmpDir)
		}()

		// Create .env.example
		envExample := filepath.Join(tmpDir, ".env.example")
		err = os.WriteFile(envExample, []byte("APP_NAME=TestApp"), 0644)
		assert.Nil(t, err)

		// Mock go mod tidy failure
		mockProcess.EXPECT().WithSpinner("Installing dependencies").Return(mockProcess).Once()
		mockProcess.EXPECT().Path(tmpDir).Return(mockProcess).Once()
		mockModTidyResult := mocksprocess.NewResult(t)
		mockModTidyResult.EXPECT().Failed().Return(true).Once()
		mockModTidyResult.EXPECT().Error().Return(assert.AnError).Once()
		mockProcess.EXPECT().Run("go", "mod", "tidy").Return(mockModTidyResult).Once()

		err = newCommand.initProject(tmpDir)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "failed to install dependencies")
	})

	t.Run("fails when key:generate fails", func(t *testing.T) {
		mockFactory := frameworkmock.Factory()
		mockProcess := mockFactory.Process()

		tmpDir, err := os.MkdirTemp("", "test-init-project-keyfail")
		assert.Nil(t, err)
		defer func() {
			_ = os.RemoveAll(tmpDir)
		}()

		// Create .env.example
		envExample := filepath.Join(tmpDir, ".env.example")
		err = os.WriteFile(envExample, []byte("APP_NAME=TestApp"), 0644)
		assert.Nil(t, err)

		// Mock go mod tidy success
		mockProcess.EXPECT().WithSpinner("Installing dependencies").Return(mockProcess).Once()
		mockProcess.EXPECT().Path(tmpDir).Return(mockProcess).Once()
		mockModTidyResult := mocksprocess.NewResult(t)
		mockModTidyResult.EXPECT().Failed().Return(false).Once()
		mockProcess.EXPECT().Run("go", "mod", "tidy").Return(mockModTidyResult).Once()

		// Mock key:generate failure
		mockProcess.EXPECT().WithSpinner("Generating application key").Return(mockProcess).Once()
		mockProcess.EXPECT().Path(tmpDir).Return(mockProcess).Once()
		mockKeyGenResult := mocksprocess.NewResult(t)
		mockKeyGenResult.EXPECT().Failed().Return(true).Once()
		mockKeyGenResult.EXPECT().Error().Return(assert.AnError).Once()
		mockProcess.EXPECT().Run("go", "run", ".", "artisan", "key:generate").Return(mockKeyGenResult).Once()

		err = newCommand.initProject(tmpDir)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "failed to generate app key")
	})

	t.Run("sets artisan file permissions when artisan exists", func(t *testing.T) {
		mockFactory := frameworkmock.Factory()
		mockProcess := mockFactory.Process()

		tmpDir, err := os.MkdirTemp("", "test-init-artisan")
		assert.Nil(t, err)
		defer func() {
			_ = os.RemoveAll(tmpDir)
		}()

		// Create .env.example
		envExample := filepath.Join(tmpDir, ".env.example")
		err = os.WriteFile(envExample, []byte("APP_NAME=TestApp"), 0644)
		assert.Nil(t, err)

		// Create artisan file with different permissions
		artisanFile := filepath.Join(tmpDir, "artisan")
		err = os.WriteFile(artisanFile, []byte("#!/bin/bash\necho test"), 0644)
		assert.Nil(t, err)

		// Mock go mod tidy
		mockProcess.EXPECT().WithSpinner("Installing dependencies").Return(mockProcess).Once()
		mockProcess.EXPECT().Path(tmpDir).Return(mockProcess).Once()
		mockModTidyResult := mocksprocess.NewResult(t)
		mockModTidyResult.EXPECT().Failed().Return(false).Once()
		mockProcess.EXPECT().Run("go", "mod", "tidy").Return(mockModTidyResult).Once()

		// Mock key:generate
		mockProcess.EXPECT().WithSpinner("Generating application key").Return(mockProcess).Once()
		mockProcess.EXPECT().Path(tmpDir).Return(mockProcess).Once()
		mockKeyGenResult := mocksprocess.NewResult(t)
		mockKeyGenResult.EXPECT().Failed().Return(false).Once()
		mockProcess.EXPECT().Run("go", "run", ".", "artisan", "key:generate").Return(mockKeyGenResult).Once()

		err = newCommand.initProject(tmpDir)
		assert.Nil(t, err)

		// Verify artisan permissions were set correctly
		info, err := os.Stat(artisanFile)
		assert.Nil(t, err)
		// On Windows, file permissions are different (0666), on Unix-like systems it's 0755
		if env.IsWindows() {
			assert.Equal(t, os.FileMode(0666), info.Mode().Perm())
		} else {
			assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
		}
	})
}

func TestReplaceModule(t *testing.T) {
	newCommand := &NewCommand{}

	tmpDir, err := os.MkdirTemp("", "test-replace-module")
	assert.Nil(t, err)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	goFile := filepath.Join(tmpDir, "main.go")
	modFile := filepath.Join(tmpDir, "go.mod")
	invalidFile := filepath.Join(tmpDir, "ignore.txt")

	newModule := "github.com/example/project"

	// Write content to files
	err = os.WriteFile(goFile, []byte(`package main
import "goravel/utils"
func main() {}`), os.ModePerm)
	assert.Nil(t, err)

	err = os.WriteFile(modFile, []byte("module goravel\nrequire example.com v1.0.0"), os.ModePerm)
	assert.Nil(t, err)

	err = os.WriteFile(invalidFile, []byte("This is a test file."), os.ModePerm)
	assert.Nil(t, err)

	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Spinner("Updating module name to \""+newModule+"\"", mock.MatchedBy(func(opt console.SpinnerOption) bool {
		return opt.Action != nil
	})).RunAndReturn(func(_ string, opt console.SpinnerOption) error {
		return opt.Action()
	}).Once()

	err = newCommand.replaceModule(mockContext, tmpDir, newModule)
	assert.Nil(t, err)

	modContent, err := os.ReadFile(modFile)
	assert.Nil(t, err)
	assert.Contains(t, string(modContent), "module github.com/example/project")
	assert.NotContains(t, string(modContent), "module goravel")

	goContent, err := os.ReadFile(goFile)
	assert.Nil(t, err)
	assert.Contains(t, string(goContent), "import \"github.com/example/project/utils\"")
	assert.NotContains(t, string(goContent), "import \"goravel/utils\"")

	invalidContent, err := os.ReadFile(invalidFile)
	assert.Nil(t, err)
	assert.Equal(t, "This is a test file.", string(invalidContent))
}
