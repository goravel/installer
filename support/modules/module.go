package modules

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/support/color"
	supportconsole "github.com/goravel/framework/support/console"
)

var installed = make(map[string]bool)

type Driver struct {
	Name        string
	Signature   string
	Package     string
	ModifyFiles func(path string) error
}

type Module struct {
	Name            string
	DefaultDriver   string
	SupportMultiple bool
	Drivers         []Driver
	chosenDrivers   []string
}

type Modules []*Module

func (r *Modules) ChoiceDriver(ctx console.Context) error {
	for _, module := range *r {
		if err := module.ChoiceDriver(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (r *Modules) Install(ctx console.Context, version, path string) error {
	for _, module := range *r {
		if err := module.Install(ctx, version, path); err != nil {
			return err
		}
	}

	return nil
}

func (r *Module) ChoiceDriver(ctx console.Context) error {
	if r.SupportMultiple {
		return r.choiceMultiDriver(ctx)
	}

	return r.choiceSingleDriver(ctx)
}

func (r *Module) Install(ctx console.Context, version, path string) error {
	var defaultDriver Driver
	for _, driver := range r.Drivers {
		// default driver not need to install
		if driver.Signature == r.DefaultDriver {
			defaultDriver = driver
			continue
		}

		if slices.Contains(r.chosenDrivers, driver.Signature) {
			if err := driver.Install(ctx, version, path); err != nil {
				return err
			}
			color.Successln(fmt.Sprintf("installed %s driver for %s.", driver.Name, r.Name))
		}
	}

	// uninstall default driver
	if !slices.Contains(r.chosenDrivers, defaultDriver.Signature) && len(defaultDriver.Package) != 0 {
		if err := defaultDriver.Uninstall(ctx, path); err != nil {
			return err
		}
		color.Successln(fmt.Sprintf("uninstalled %s driver for %s.", defaultDriver.Name, r.Name))
	}

	return nil
}

func (r *Module) checkDriver(drivers ...string) error {
	var available []string
	for _, driver := range r.Drivers {
		available = append(available, driver.Signature)
	}

	for _, driver := range drivers {
		if !slices.Contains(available, driver) {
			return fmt.Errorf("invalid %s driver [%s]. Valid options are: %s", r.Name, driver, strings.Join(available, ", "))
		}
	}

	return nil
}

func (r *Module) choiceMultiDriver(ctx console.Context) (err error) {
	if drivers := ctx.OptionSlice(r.Name); len(drivers) > 0 {
		r.chosenDrivers = drivers
		return r.checkDriver(drivers...)
	}

	if r.chosenDrivers, err = ctx.MultiSelect(fmt.Sprintf("Which %s drivers will your application use?", r.Name), r.getChoiceOption(), console.MultiSelectOption{
		Default: strings.Split(r.DefaultDriver, ","),
	}); err != nil {
		return err
	}

	return nil
}

func (r *Module) choiceSingleDriver(ctx console.Context) error {
	if driver := ctx.Option(r.Name); driver != "" {
		r.chosenDrivers = []string{driver}
		return r.checkDriver(driver)
	}

	driver, err := ctx.Choice(fmt.Sprintf("Which %s driver will your application use?", r.Name), r.getChoiceOption(), console.ChoiceOption{
		Default: r.DefaultDriver,
	})
	if err != nil {
		return err
	}

	r.chosenDrivers = []string{driver}

	return nil
}

func (r *Module) getChoiceOption() []console.Choice {
	var options []console.Choice
	for _, driver := range r.Drivers {
		options = append(options, console.Choice{
			Key:   driver.Name,
			Value: driver.Signature,
		})
	}

	return options
}

func (r *Driver) Install(ctx console.Context, version, path string) error {
	if len(r.Package) > 0 && !installed[r.Package] {
		install := exec.Command("go", "run", ".", "artisan", "package:install", r.Package+"@"+version)
		install.Dir = path

		if err := supportconsole.ExecuteCommand(ctx, install); err != nil {
			return err
		}
		installed[r.Package] = true
	}

	if r.ModifyFiles != nil {
		if err := r.ModifyFiles(path); err != nil {
			return err
		}
	}

	return nil
}

func (r *Driver) Uninstall(ctx console.Context, path string) error {
	uninstall := exec.Command("go", "run", ".", "artisan", "package:uninstall", r.Package)
	uninstall.Dir = path

	return supportconsole.ExecuteCommand(ctx, uninstall)
}
