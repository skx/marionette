package modules

import (
	"fmt"

	"github.com/skx/marionette/config"
	"github.com/skx/marionette/modules/system"
)

// DPKGModule stores our state
type DPKGModule struct {

	// cfg contains our configuration object.
	cfg *config.Config
}

// Check is part of the module-api, and checks arguments.
func (dm *DPKGModule) Check(args map[string]interface{}) error {

	// Ensure we have a command to run.
	_, ok := args["package"]
	if !ok {
		return fmt.Errorf("missing 'package' parameter")
	}

	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (dm *DPKGModule) Execute(args map[string]interface{}) (bool, error) {

	// Did we make a change, by removing a package?
	changed := false

	// Package abstraction
	pkg := system.New()

	// We might have multiple packages
	var packages []string

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

	// For each package, remove if installed
	for _, name := range packages {

		// Is it instaled?
		inst, err := pkg.IsInstalled(name)
		if err != nil {
			return false, err
		}

		// If it is not installed we do nothing
		if !inst {
			continue
		}

		if dm.cfg.Verbose {
			fmt.Printf("\tPackage is not installed, removing: %s\n", name)
		}
		// Uninstall
		err = pkg.Uninstall(name)
		if err != nil {
			return false, err
		}

		changed = true
	}

	return changed, nil
}

// init is used to dynamically register our module.
func init() {
	Register("dpkg", func(cfg *config.Config) ModuleAPI {
		return &DPKGModule{cfg: cfg}
	})
}
