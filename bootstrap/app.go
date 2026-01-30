package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"

	"github.com/goravel/installer/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithConfig(config.Boot).
		WithProviders(Providers).
		Create()
}
