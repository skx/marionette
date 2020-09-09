package modules

import (
	"fmt"

	"github.com/skx/marionette/config"
	"github.com/skx/marionette/modules/system"
)

// AptModule stores our state
type AptModule struct {

	// cfg contains our configuration object.
	cfg *config.Config
}

// Check is part of the module-api, and checks arguments.
func (am *AptModule) Check(args map[string]interface{}) error {

	// Ensure we have a package to install.
	_, ok := args["package"]
	if !ok {
		return fmt.Errorf("missing 'package' parameter")
	}

	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (am *AptModule) Execute(args map[string]interface{}) (bool, error) {

	// Did we make a change, by installing a package?
	changed := false

	// Package abstraction
	pkg := system.New()

	// Are we updating first?
	p := StringParam(args, "update")
	if p == "yes" {
		err := pkg.Update()
		if err != nil {
			return false, err
		}
	}

	// We might have multiple packages
	var packages []string

	// Single package?
	p = StringParam(args, "package")
	if p != "" {
		packages = append(packages, p)
	}

	// Array of packages?
	a := ArrayParam(args, "package")
	if len(a) > 0 {
		packages = append(packages, a...)
	}

	// For each package, install if missing
	for _, name := range packages {

		// Is it instaled?
		inst, err := pkg.IsInstalled(name)
		if err != nil {
			return false, err
		}

		// If not then we add it
		if !inst {
			if am.cfg.Verbose {
				fmt.Printf("\tPackage is not installed, installing: %s\n", name)
			}

			// Install
			err = pkg.Install(name)
			if err != nil {
				return false, err
			}

			changed = true
		}
	}

	return changed, nil
}

// init is used to dynamically register our module.
func init() {
	Register("apt", func(cfg *config.Config) ModuleAPI {
		return &AptModule{cfg: cfg}
	})
}
