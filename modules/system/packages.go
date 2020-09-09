// Package system contains some helpers for working with operating-system
// package management.
//
// Currently only Debian GNU/Linux systems are supported, but that might
// change.
package system

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// Known-system types
const (
	CENTOS_YUM = "CENTOS_YUM"
	DEBIAN     = "DEBIAN"
)

// Mapping between CLI packages and systems
var (

	// These are used to identify systems.
	mappings = map[string]string{
		"/usr/bin/dpkg": DEBIAN,
		"/usr/bin/yum":  CENTOS_YUM,
	}

	// Is installed?
	checkCmd = map[string]string{
		DEBIAN:     "dpkg -s %s",
		CENTOS_YUM: "yum list installed %s",
	}

	// Install command for different systems.
	installCmd = map[string]string{
		DEBIAN:     "apt-get install --yes %s",
		CENTOS_YUM: "yum install --assume-yes %s",
	}

	// Uninstallation command for different systems
	uninstallCmd = map[string]string{
		DEBIAN:     "dpkg --purge %s",
		CENTOS_YUM: "yum remove %s",
	}

	// Update command for each system
	updateCmd = map[string]string{
		DEBIAN:     "apt-get update --quiet --quiet",
		CENTOS_YUM: "??",
	}
)

// Package maintains our object state
type Package struct {

	// System contains our identified system.
	system string
}

// New creates a new instance of this object, attempting to identify the
// system during the initial phase.
func New() *Package {
	p := &Package{}
	p.identify()
	return p
}

// identify tries to identify this system, if a binary we know is found
// then it is assumed to be used - this might cause confusion if a Debian
// system has RPM installed, for example, but should otherwise perform
// reasonably well.
func (p *Package) identify() {

	// Look over our helpers
	for file, system := range mappings {

		_, err := os.Stat(file)
		if err == nil {
			p.system = system
			return
		}
	}
}

// System returns the O/S we've identified
func (p *Package) System() string {
	return p.system
}

// IsKnown reports whether this system is using a known packaging-system.
func (p *Package) IsKnown() bool {
	return (p.system != "")
}

// Update carries out the update command for a given system
func (p *Package) Update() error {

	if !p.IsKnown() {
		return fmt.Errorf("failed to recognize system-type")
	}

	// Get the command
	tmp := updateCmd[p.System()]

	// Split
	run := strings.Split(tmp, " ")

	// Run
	cmd := exec.Command(run[0], run[1:]...)
	if err := cmd.Start(); err != nil {
		return err
	}

	// Wait for completion
	if err := cmd.Wait(); err != nil {

		if exiterr, ok := err.(*exec.ExitError); ok {

			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return fmt.Errorf("exit code for '%s' was %d", tmp, status.ExitStatus())
			}
		}
	}

	return nil
}

// IsInstalled checks a package installed?
func (p *Package) IsInstalled(name string) (bool, error) {

	if !p.IsKnown() {
		return false, fmt.Errorf("failed to recognize system-type")
	}

	// Get the command
	tmp := checkCmd[p.System()]
	tmp = strings.ReplaceAll(tmp, "%s", name)

	// Split
	run := strings.Split(tmp, " ")

	// Run
	cmd := exec.Command(run[0], run[1:]...)
	if err := cmd.Start(); err != nil {
		return false, err
	}

	// Wait for completion
	if err := cmd.Wait(); err != nil {

		if exiterr, ok := err.(*exec.ExitError); ok {

			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				log.Printf("Exit Status: %d", status.ExitStatus())
				return false, nil
			}
		}
	}

	// Package is installed.
	return true, nil

}

// Install a single package to the system.
func (p *Package) Install(name string) error {
	return fmt.Errorf("todo - use our installCmd")
}

// InstallPackages allows Installing multiple packages to the system.
func (p *Package) InstallPackages(names []string) error {

	for _, ent := range names {
		err := p.Install(ent)
		if err != nil {
			return err
		}
	}

	return nil
}

// Uninstall a single package from the system.
func (p *Package) Uninstall(name string) error {
	return fmt.Errorf("todo - use our uninstallCmd")
}

// UninstallPackages allows uninstalling multiple packages from the system.
func (p *Package) UninstallPackages(names []string) error {

	for _, ent := range names {
		err := p.Uninstall(ent)
		if err != nil {
			return err
		}
	}

	return nil
}
