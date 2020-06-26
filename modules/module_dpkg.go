package modules

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/skx/marionette/config"
)

// DPKGModule stores our state
type DPKGModule struct {

	// cfg contains our configuration object.
	cfg *config.Config
}

// Check is part of the module-api, and checks arguments.
func (f *DPKGModule) Check(args map[string]interface{}) error {

	// Ensure we have a command to run.
	_, ok := args["package"]
	if !ok {
		return fmt.Errorf("missing 'package' parameter")
	}

	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (f *DPKGModule) Execute(args map[string]interface{}) (bool, error) {

	// We might have multiple packages
	var packages []string
	packages = append(packages, "--purge")

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

	// Now run
	cmd := exec.Command("dpkg", packages...)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("error running command %s", err.Error())
	}

	return false, nil
}

// init is used to dynamically register our module.
func init() {
	Register("dpkg", func(cfg *config.Config) ModuleAPI {
		return &DPKGModule{cfg: cfg}
	})
}
