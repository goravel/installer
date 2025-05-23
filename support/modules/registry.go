package modules

import (
	"path/filepath"

	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"

	"github.com/goravel/installer/support/envfile"
)

var Cache = Module{
	Name:          "cache",
	DefaultDriver: "memory",
	Drivers: []Driver{
		{
			Name:      "Memory",
			Signature: "memory",
		},
		{
			Name:      "Redis",
			Signature: "redis",
			Package:   "github.com/goravel/redis",
			ModifyFiles: func(path string) error {
				return envfile.Modify(filepath.Join(path, ".env"), map[string]string{
					"CACHE_STORE": "redis",
				})
			},
		},
	},
}

var Database = Module{
	Name:          "database",
	DefaultDriver: "postgres",
	Drivers: []Driver{
		{
			Name:      "SQLite",
			Signature: "sqlite",
			Package:   "github.com/goravel/sqlite",
			ModifyFiles: func(path string) error {
				return envfile.Modify(filepath.Join(path, ".env"), map[string]string{
					"DB_CONNECTION": "sqlite",
					"DB_DATABASE":   "forge",
				})
			},
		},
		{
			Name:      "MySQL",
			Signature: "mysql",
			Package:   "github.com/goravel/mysql",
			ModifyFiles: func(path string) error {
				return envfile.Modify(filepath.Join(path, ".env"), map[string]string{
					"DB_CONNECTION": "mysql",
					"DB_HOST":       "",
					"DB_PORT":       "3306",
					"DB_DATABASE":   "forge",
					"DB_USERNAME":   "",
					"DB_PASSWORD":   "",
				})
			},
		},
		{
			Name:      "PostgreSQL",
			Signature: "postgres",
			Package:   "github.com/goravel/postgres",
		},
		{
			Name:      "SQL Server",
			Signature: "sqlserver",
			Package:   "github.com/goravel/sqlserver",
			ModifyFiles: func(path string) error {
				return envfile.Modify(filepath.Join(path, ".env"), map[string]string{
					"DB_CONNECTION": "sqlserver",
					"DB_HOST":       "",
					"DB_PORT":       "1433",
					"DB_DATABASE":   "forge",
					"DB_USERNAME":   "",
					"DB_PASSWORD":   "",
				})
			},
		},
	},
}

var HTTP = Module{
	Name:          "http",
	DefaultDriver: "gin",
	Drivers: []Driver{
		{
			Name:      "Gin",
			Signature: "gin",
			Package:   "github.com/goravel/gin",
		},
		{
			Name:      "Fiber",
			Signature: "fiber",
			Package:   "github.com/goravel/fiber",
			ModifyFiles: func(path string) error {
				return modify.GoFile(filepath.Join(path, "config", "http.go")).
					Find(match.Config("http")).
					Modify(modify.ReplaceConfig("default", `"fiber"`)).Apply()
			},
		},
	},
}

var Queue = Module{
	Name:          "queue",
	DefaultDriver: "sync",
	Drivers: []Driver{
		{
			Name:      "Sync",
			Signature: "sync",
		},
		{
			Name:      "Redis",
			Signature: "redis",
			Package:   "github.com/goravel/redis",
			ModifyFiles: func(path string) error {
				return envfile.Modify(filepath.Join(path, ".env"), map[string]string{
					"QUEUE_CONNECTION": "redis",
				})
			},
		},
	},
}

var Session = Module{
	Name:          "session",
	DefaultDriver: "file",
	Drivers: []Driver{
		{
			Name:      "File",
			Signature: "file",
		},
		{
			Name:      "Redis",
			Signature: "redis",
			Package:   "github.com/goravel/redis",
			ModifyFiles: func(path string) error {
				return envfile.Modify(filepath.Join(path, ".env"), map[string]string{
					"SESSION_DRIVER": "redis",
				})
			},
		},
	},
}

var Storage = Module{
	Name:            "storage",
	DefaultDriver:   "local",
	SupportMultiple: true,
	Drivers: []Driver{
		{
			Name:      "Local",
			Signature: "local",
		},
		{
			Name:      "S3",
			Signature: "s3",
			Package:   "github.com/goravel/s3",
		},
		{
			Name:      "OSS",
			Signature: "oss",
			Package:   "github.com/goravel/oss",
		},
		{
			Name:      "COS",
			Signature: "cos",
			Package:   "github.com/goravel/cos",
		},
		{
			Name:      "MinIO",
			Signature: "minio",
			Package:   "github.com/goravel/minio",
		},
		{
			Name:      "Cloudinary",
			Signature: "cloudinary",
			Package:   "github.com/goravel/cloudinary",
		},
	},
}
