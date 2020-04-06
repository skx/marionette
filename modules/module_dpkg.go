package modules

import (
	"fmt"
	"os"
	"os/exec"
)

// DPKGModule stores our state
type DPKGModule struct {
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

	// Get the packages we're going to add/remove
	c, ok := args["package"]
	if !ok {
		return false, fmt.Errorf("missing 'package' parameter")
	}

	// We might have multiple packages
	var packages []string
	packages = append(packages, "--purge")

	// cast to string
	str, ok := c.(string)
	if ok {
		packages = append(packages, str)
	}
	strs, ok := c.([]string)
	if ok {
		packages = append(packages, strs...)
	}

	// Now run
	cmd := exec.Command("dpkg", packages...)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("error running command '%s' %s", str, err.Error())
	}

	return false, nil
}

// init is used to dynamically register our module.
func init() {
	Register("dpkg", func() ModuleAPI {
		return &DPKGModule{}
	})
}
