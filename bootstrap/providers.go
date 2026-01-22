package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/process"

	"github.com/goravel/installer/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&process.ServiceProvider{},
		&providers.ArtisanServiceProvider{},
	}
}
