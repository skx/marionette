package modules

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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

	// Ensure we have a command to run.
	_, ok := args["package"]
	if !ok {
		return fmt.Errorf("missing 'package' parameter")
	}

	return nil
}

// isInstalled tests if the package is installed
func (am *AptModule) isInstalled(pkg string) (bool, error) {

	x := system.New()
	res, err := x.IsInstalled(pkg)
	return res, err
}

// Execute is part of the module-api, and is invoked to run a rule.
func (am *AptModule) Execute(args map[string]interface{}) (bool, error) {

	// Package abstraction
	x := system.New()

	// Are we updating first?
	p := StringParam(args, "update")
	if p == "yes" {

		err := x.Update()
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

	// Assume all packages are installed, but if not we'll add them.
	installed := true
	for _, pkg := range packages {
		present, err := am.isInstalled(pkg)
		if err != nil {
			return false, err
		}

		// One package was missing; so we'll install.
		if !present {
			installed = false
			if am.cfg.Verbose {
				fmt.Printf("\tPackages not installed: %s\n", pkg)
			}

		}
	}

	// Package(s) are installed already.
	if installed {
		if am.cfg.Verbose {
			fmt.Printf("\tPackages installed already: %s\n", strings.Join(packages, ","))
		}

		return false, nil
	}

	// Now run
	packages = append([]string{"install", "--yes"}, packages...)
	cmd := exec.Command("apt-get", packages...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("error running command %s", err.Error())
	}

	return true, nil
}

// init is used to dynamically register our module.
func init() {
	Register("apt", func(cfg *config.Config) ModuleAPI {
		return &AptModule{cfg: cfg}
	})
}
