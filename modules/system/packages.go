// Package system contains some helpers for working with operating-system
// package management.
//
// Currently only Debian GNU/Linux systems are supported, but that might
// change.
package system

import (
	"fmt"
	"os"
)

// Known-system types
const (
	CENTOS = "CENTOS"
	DEBIAN = "DEBIAN"
)

// Mapping between CLI packages and systems
var (

	// These are used to identify systems.
	mappings = map[string]string{
		"/usr/bin/dpkg": DEBIAN,
		"/usr/bin/yum":  CENTOS,
	}

	// Install command for different systems.
	installCmd = map[string]string{
		DEBIAN: "apt-get install --yes %s",
		CENTOS: "yum install --assume-yes %s",
	}

	// Uninstallation command for different systems
	uninstallCmd = map[string]string{
		DEBIAN: "dpkg --purge %s",
		CENTOS: "rpm -e %s",
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
