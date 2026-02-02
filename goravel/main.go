package main

import (
	"github.com/goravel/installer/bootstrap"
)

func main() {
	app := bootstrap.Boot()

	app.Start()
}
