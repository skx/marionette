// This module handles package installation/removal.

package modules

import (
	"fmt"
	"log"
	"strings"

	"github.com/skx/marionette/config"
	"github.com/skx/marionette/environment"
	"github.com/skx/marionette/modules/system"
)

// PackageModule stores our state
type PackageModule struct {

	// cfg contains our configuration object.
	cfg *config.Config

	// env holds our environment
	env *environment.Environment

	// state when using a compatibility-module
	state string
}

// Check is part of the module-api, and checks arguments.
func (pm *PackageModule) Check(args map[string]interface{}) error {

	// Ensure we have a package, or set of packages, to install/uninstall.
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

// Execute is part of the module-api, and is invoked to run a rule.
func (pm *PackageModule) Execute(args map[string]interface{}) (bool, error) {

	// Did we make a change, by installing/removing a package?
	changed := false

	// Package abstraction
	pkg := system.New()

	// Do we need to use doas/sudo?
	privs := StringParam(args, "elevate")
	if privs != "" {
		pkg.UsePrivilegeHelper(privs)
	}

	// Get the packages we're working with
	packages := ArrayCastParam(args, "package")

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

	// We might have 10+ packages, but we want to ensure that we
	// install/remove all the packages at once.
	//
	// So while we can test the packages that are already present
	// to work out our actions we do need to add/remove things
	// en mass
	toInstall := []string{}
	toRemove := []string{}

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

			// Save this package as something to install
			// once we've tested the rest.
			toInstall = append(toInstall, name)
		}

		if state == "absent" {

			// If it is not installed we have nothing to do.
			if !inst {
				continue
			}

			// Save this package as something to remove
			// once we've tested the rest
			toRemove = append(toRemove, name)
		}
	}

	// Something to install?
	if len(toInstall) > 0 {

		// Log it
		log.Printf("[DEBUG] Package(s) which need to be installed: %s", strings.Join(toInstall, ","))

		// Do it
		err := pkg.Install(toInstall)
		if err != nil {
			return false, err
		}

		// We resulted in a change, because we had things to install
		// and presumably they're now installed.
		changed = true
	}

	// Something to uninstall?
	if len(toRemove) > 0 {

		// Log it
		log.Printf("[DEBUG] Package(s) which need to be removed: %s", strings.Join(toRemove, ","))

		// Do it
		err := pkg.Uninstall(toRemove)
		if err != nil {
			return false, err
		}

		// We resulted in a change, because we had things to remove
		// and presumably they're now purged.
		changed = true
	}

	return changed, nil
}

// init is used to dynamically register our module.
func init() {
	Register("package", func(cfg *config.Config, env *environment.Environment) ModuleAPI {
		return &PackageModule{
			cfg: cfg,
			env: env,
		}
	})

	// compatibility with previous releases.
	Register("apt", func(cfg *config.Config, env *environment.Environment) ModuleAPI {
		return &PackageModule{
			cfg:   cfg,
			env:   env,
			state: "installed",
		}
	})
	Register("dpkg", func(cfg *config.Config, env *environment.Environment) ModuleAPI {
		return &PackageModule{
			cfg:   cfg,
			env:   env,
			state: "absent",
		}
	})
}
