package modules

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/console"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
)

type ModulesTestSuite struct {
	suite.Suite
	mockContext *mocksconsole.Context
}

func (s *ModulesTestSuite) SetupTest() {
	s.mockContext = mocksconsole.NewContext(s.T())
}

func (s *ModulesTestSuite) TearDownTest() {}

func TestModulesTestSuite(t *testing.T) {
	suite.Run(t, new(ModulesTestSuite))
}

func (s *ModulesTestSuite) TestChoiceDriver() {
	tests := []struct {
		name    string
		modules Modules
		setup   func()
		assert  func(output string, err error)
	}{
		{
			name:    "single driver invalid",
			modules: Modules{&Cache},
			setup: func() {
				s.mockContext.EXPECT().Option(Cache.Name).Return("invalid").Once()

			},
			assert: func(output string, err error) {
				s.Require().ErrorContains(err, "invalid cache driver")
				s.Empty(output)
			},
		},
		{
			name:    "single driver valid",
			modules: Modules{&Cache},
			setup: func() {
				s.mockContext.EXPECT().Option(Cache.Name).Return("redis").Once()
			},
			assert: func(output string, err error) {
				s.Require().NoError(err)
				s.Contains(Cache.chosenDrivers, "redis")
				s.Empty(output)
			},
		},
		{
			name:    "single driver choice failed",
			modules: Modules{&Cache},
			setup: func() {
				s.mockContext.EXPECT().Option(Cache.Name).Return("").Once()
				s.mockContext.EXPECT().Choice("Which cache driver will your application use?",
					Cache.getChoiceOption(),
					console.ChoiceOption{Default: Cache.DefaultDriver},
				).Return("", assert.AnError).Once()
			},
			assert: func(output string, err error) {
				s.Require().ErrorIs(err, assert.AnError)
				s.Empty(output)
			},
		},
		{
			name:    "single driver choice success",
			modules: Modules{&Cache},
			setup: func() {
				s.mockContext.EXPECT().Option(Cache.Name).Return("").Once()
				s.mockContext.EXPECT().Choice("Which cache driver will your application use?",
					Cache.getChoiceOption(),
					console.ChoiceOption{Default: Cache.DefaultDriver},
				).Return("redis", nil).Once()
			},
			assert: func(output string, err error) {
				s.Require().NoError(err)
				s.Contains(Cache.chosenDrivers, "redis")
				s.Empty(output)
			},
		},
		{
			name:    "multiple driver invalid",
			modules: Modules{&Storage},
			setup: func() {
				s.mockContext.EXPECT().OptionSlice(Storage.Name).Return([]string{"invalid"}).Once()
			},
			assert: func(output string, err error) {
				s.Require().ErrorContains(err, "invalid storage driver")
				s.Empty(output)
			},
		},
		{
			name:    "multiple driver valid",
			modules: Modules{&Storage},
			setup: func() {
				s.mockContext.EXPECT().OptionSlice(Storage.Name).Return([]string{"local", "s3"}).Once()
			},
			assert: func(output string, err error) {
				s.Require().NoError(err)
				s.Contains(Storage.chosenDrivers, "local")
				s.Contains(Storage.chosenDrivers, "s3")
				s.Empty(output)
			},
		},
		{
			name:    "multiple driver choice failed",
			modules: Modules{&Storage},
			setup: func() {
				s.mockContext.EXPECT().OptionSlice(Storage.Name).Return([]string{}).Once()
				s.mockContext.EXPECT().MultiSelect("Which storage drivers will your application use?",
					Storage.getChoiceOption(),
					console.MultiSelectOption{Default: []string{Storage.DefaultDriver}},
				).Return([]string{}, assert.AnError).Once()
			},
			assert: func(output string, err error) {
				s.Require().ErrorIs(err, assert.AnError)
				s.Empty(output)
			},
		},
		{
			name:    "multiple driver choice success",
			modules: Modules{&Storage},
			setup: func() {
				s.mockContext.EXPECT().OptionSlice(Storage.Name).Return([]string{}).Once()
				s.mockContext.EXPECT().MultiSelect("Which storage drivers will your application use?",
					Storage.getChoiceOption(),
					console.MultiSelectOption{Default: []string{Storage.DefaultDriver}},
				).Return([]string{"local", "s3"}, nil).Once()
			},
			assert: func(output string, err error) {
				s.Require().NoError(err)
				s.Contains(Storage.chosenDrivers, "local")
				s.Contains(Storage.chosenDrivers, "s3")
				s.Empty(output)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setup()
			var err error
			output := color.CaptureOutput(func(w io.Writer) {
				err = tt.modules.ChoiceDriver(s.mockContext)
			})
			tt.assert(output, err)
		})

	}
}

func (s *ModulesTestSuite) TestInstall() {
	tests := []struct {
		name    string
		modules Modules
		setup   func(path string)
		assert  func(output, path string, err error)
	}{
		{
			name:    "install skipped",
			modules: Modules{&Cache},
			setup: func(path string) {
				s.mockContext.EXPECT().Option(Cache.Name).Return("memory").Once()
			},
			assert: func(output, path string, err error) {
				s.Require().NoError(err)
				s.Empty(output)
			},
		},
		{
			name:    "install failed(package install failed)",
			modules: Modules{&Cache},
			setup: func(path string) {
				s.mockContext.EXPECT().Option(Cache.Name).Return("redis").Once()
				s.mockContext.EXPECT().Spinner("> @go run . artisan package:install github.com/goravel/redis@latest", mock.Anything).Return(assert.AnError).Once()
			},
			assert: func(output, path string, err error) {
				s.Require().ErrorIs(err, assert.AnError)
				s.Empty(output)
			},
		},
		{
			name:    "install failed(env file modify failed)",
			modules: Modules{&Database},
			setup: func(path string) {
				s.mockContext.EXPECT().Option(Database.Name).Return("mysql").Once()
				s.mockContext.EXPECT().Spinner("> @go run . artisan package:install github.com/goravel/mysql@latest", mock.Anything).Return(nil).Once()
			},
			assert: func(output, path string, err error) {
				s.Require().ErrorIs(err, os.ErrNotExist)
				s.Empty(output)
			},
		},
		{
			name:    "install failed(default driver uninstall failed)",
			modules: Modules{&Database},
			setup: func(path string) {
				s.Require().NoError(file.PutContent(filepath.Join(path, ".env"), "DB_CONNECTION=postgres"))

				s.mockContext.EXPECT().Option(Database.Name).Return("sqlite").Once()
				s.mockContext.EXPECT().Spinner("> @go run . artisan package:install github.com/goravel/sqlite@latest", mock.Anything).Return(nil).Once()
				s.mockContext.EXPECT().Spinner("> @go run . artisan package:uninstall github.com/goravel/postgres", mock.Anything).Return(assert.AnError).Once()
			},
			assert: func(output, path string, err error) {
				s.False(file.Contain(filepath.Join(path, ".env"), "DB_CONNECTION=postgres"), "db connection should not be postgres")
				s.True(file.Contain(filepath.Join(path, ".env"), "DB_CONNECTION=sqlite"), "db connection should be sqlite")

				s.Contains(output, "installed SQLite driver for database.")
				s.Require().ErrorIs(err, assert.AnError)
			},
		},
		{
			name:    "install success(skipped installed drivers)",
			modules: Modules{&Queue, &Session},
			setup: func(path string) {
				installed = make(map[string]bool)
				s.Require().NoError(file.PutContent(filepath.Join(path, ".env"), "SESSION_DRIVER=file"))

				s.mockContext.EXPECT().Option(Queue.Name).Return("redis").Once()
				s.mockContext.EXPECT().Option(Session.Name).Return("redis").Once()

				s.mockContext.EXPECT().Spinner("> @go run . artisan package:install github.com/goravel/redis@latest", mock.Anything).Return(nil).Once()
			},
			assert: func(output, path string, err error) {
				s.Require().NoError(err)
				s.False(file.Contain(filepath.Join(path, ".env"), "SESSION_DRIVER=file"), "session driver should not be file")
				s.True(file.Contain(filepath.Join(path, ".env"), "SESSION_DRIVER=redis"), "session driver should be redis")
				s.Contains(output, "installed Redis driver for queue.")
				s.Contains(output, "installed Redis driver for session.")
			},
		},
		{
			name:    "install success",
			modules: Modules{&Cache, &Database, &HTTP, &Queue},
			setup: func(path string) {
				installed = make(map[string]bool)
				s.Require().NoError(file.PutContent(filepath.Join(path, "config", "http.go"), `package config
				func init() {
					config := facades.Config()
					config.Add("http", map[string]any{
						// HTTP Driver
						"default": "gin",
					})
				}`))
				s.Require().NoError(file.PutContent(filepath.Join(path, ".env"),
					"CACHE_STORE=memory\n"+
						"DB_CONNECTION=postgres\n"+
						"DB_HOST=127.0.0.1\n"+
						"DB_PORT=5432\n"+
						"DB_DATABASE=goravel\n"+
						"DB_USERNAME=root\n"+
						"DB_PASSWORD="),
				)

				s.mockContext.EXPECT().Option(Cache.Name).Return("redis").Once()
				s.mockContext.EXPECT().Option(Database.Name).Return("sqlserver").Once()
				s.mockContext.EXPECT().Option(HTTP.Name).Return("fiber").Once()
				s.mockContext.EXPECT().Option(Queue.Name).Return("async").Once()

				s.mockContext.EXPECT().Spinner("> @go run . artisan package:install github.com/goravel/redis@latest", mock.Anything).Return(nil).Once()
				s.mockContext.EXPECT().Spinner("> @go run . artisan package:install github.com/goravel/sqlserver@latest", mock.Anything).Return(nil).Once()
				s.mockContext.EXPECT().Spinner("> @go run . artisan package:uninstall github.com/goravel/postgres", mock.Anything).Return(nil).Once()
				s.mockContext.EXPECT().Spinner("> @go run . artisan package:install github.com/goravel/fiber@latest", mock.Anything).Return(nil).Once()
				s.mockContext.EXPECT().Spinner("> @go run . artisan package:uninstall github.com/goravel/gin", mock.Anything).Return(nil).Once()

			},
			assert: func(output, path string, err error) {
				s.False(file.Contain(filepath.Join(path, ".env"), "DB_CONNECTION=postgres"), "db connection should not be postgres")
				s.True(file.Contain(filepath.Join(path, ".env"),
					"DB_CONNECTION=sqlserver\n"+
						"DB_HOST=\n"+
						"DB_PORT=1433\n"+
						"DB_DATABASE=forge\n"+
						"DB_USERNAME=\n"+
						"DB_PASSWORD=",
				), "db connection should be sqlserver")
				s.True(file.Contain(filepath.Join(path, ".env"), "QUEUE_CONNECTION=async"), "queue connection should be async")

				s.Require().NoError(err)
				s.Contains(output, "installed Redis driver for cache")
				s.Contains(output, "installed SQL Server driver for database.")
				s.Contains(output, "uninstalled PostgreSQL driver for database.")
				s.Contains(output, "installed Fiber driver for http.")
				s.Contains(output, "uninstalled Gin driver for http.")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			path := s.T().TempDir()
			tt.setup(path)
			var err error
			output := color.CaptureOutput(func(w io.Writer) {
				s.Require().NoError(tt.modules.ChoiceDriver(s.mockContext))
				err = tt.modules.Install(s.mockContext, "latest", path)
			})
			tt.assert(output, path, err)
		})

	}
}
