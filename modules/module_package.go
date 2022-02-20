// This module handles package installation/removal.

package modules

import (
	"fmt"
	"log"

	"github.com/skx/marionette/config"
	"github.com/skx/marionette/environment"
	"github.com/skx/marionette/modules/system"
)

// PackageModule stores our state
type PackageModule struct {

	// cfg contains our configuration object.
	cfg *config.Config

	// state when using a compatibility-module
	state string
}

// Check is part of the module-api, and checks arguments.
func (pm *PackageModule) Check(args map[string]interface{}) error {

	// Ensure we have a package to install/uninstall.
	_, ok := args["package"]
	if !ok {
		return fmt.Errorf("missing 'package' parameter")
	}

	// Ensure we have a state to move towards.
	state, ok2 := args["state"]
	if !ok2 {
		if pm.state != "" {
			state = pm.state
		} else {
			return fmt.Errorf("missing 'state' parameter")
		}
	}

	// The state should make sense.
	if state != "installed" && state != "absent" {
		return fmt.Errorf("package state must be either 'installed' or 'absent'")
	}

	return nil
}

// getPackages returns the packages we're operating upon.
func (pm *PackageModule) getPackages(args map[string]interface{}) []string {

	packages := []string{}

	// Single package?
	p := StringParam(args, "package")
	if p != "" {
		packages = append(packages, p)
	}

	// Array of packages?
	a := ArrayParam(args, "package")
	if len(a) > 0 {
		packages = append(packages, a...)
	}

	return packages
}

// Execute is part of the module-api, and is invoked to run a rule.
func (pm *PackageModule) Execute(env *environment.Environment, args map[string]interface{}) (bool, error) {

	// Did we make a change, by installing/removing a package?
	changed := false

	// Package abstraction
	pkg := system.New()

	// Get the packages we're working with
	packages := pm.getPackages(args)

	// Do we need to update?
	//
	// This only makes sense for a package-installation, but
	// we'll accept it for the module globally as there is no
	// harm in it.
	p := StringParam(args, "update")
	if p == "yes" || p == "true" {
		err := pkg.Update()
		if err != nil {
			return false, err
		}
	}

	// Are we installing, or uninstalling?
	state := StringParam(args, "state")
	if state == "" {
		if pm.state != "" {
			state = pm.state
		} else {
			return false, fmt.Errorf("state must be a string")
		}
	}

	// For each package install/uninstall
	for _, name := range packages {

		log.Printf("[DEBUG] Testing package %s", name)

		// Is it installed?
		inst, err := pkg.IsInstalled(name)
		if err != nil {
			return false, err
		}

		// Show the output
		if inst {
			log.Printf("[DEBUG] Package installed: %s", name)
		} else {
			log.Printf("[DEBUG] Package not installed: %s", name)
		}

		if state == "installed" {

			// Already installed, do nothing.
			if inst {
				continue
			}

			// install the missing package.
			err = pkg.Install(name)
			if err != nil {
				return false, err
			}

			changed = true
		}

		if state == "absent" {

			// If it is not installed we have nothing to do.
			if !inst {
				continue
			}

			// remove the package.
			err = pkg.Uninstall(name)
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
	Register("package", func(cfg *config.Config) ModuleAPI {
		return &PackageModule{cfg: cfg}
	})

	// compat
	Register("apt", func(cfg *config.Config) ModuleAPI {
		return &PackageModule{cfg: cfg, state: "installed"}
	})
	Register("dpkg", func(cfg *config.Config) ModuleAPI {
		return &PackageModule{cfg: cfg, state: "absent"}
	})
}
