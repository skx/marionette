package modules

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// AptModule stores our state
type AptModule struct {
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

	// Run the command
	cmd := exec.Command("dpkg", "-s", pkg)

	// Get the output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("error running command 'dpkg -s %s' %s", pkg, err.Error())
	}

	// Look for "Status:"
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "Status:") {
			if strings.Contains(line, "installed") {
				return true, nil
			}
		}
	}

	return false, nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (am *AptModule) Execute(args map[string]interface{}) (bool, error) {

	// Are we updating first?
	p := StringParam(args, "update")
	if p == "yes" {
		cmd := exec.Command("apt-get", "update")
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return false, fmt.Errorf("error running 'apt-get update' %s", err.Error())
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
			fmt.Printf("Package missing: %s\n", pkg)
		}
	}

	// Package(s) are installed already.
	if installed {
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
	Register("apt", func() ModuleAPI {
		return &AptModule{}
	})
}
