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

	"github.com/google/shlex"
)

// Known-system types
const (
	YUM    = "YUM"
	DEBIAN = "DEBIAN"
)

// Mapping between CLI packages and systems
var (

	// These are used to identify systems.
	mappings = map[string]string{
		"/usr/bin/dpkg": DEBIAN,
		"/usr/bin/yum":  YUM,
	}

	// Is installed?
	checkCmd = map[string]string{
		DEBIAN: "/usr/bin/dpkg -s %s",
		YUM:    "/usr/bin/yum list installed %s",
	}

	// Install command for different systems.
	installCmd = map[string]string{
		DEBIAN: "/usr/bin/apt-get install --yes %s",
		YUM:    "/usr/bin/yum install --assumeyes %s",
	}

	// Uninstallation command for different systems
	uninstallCmd = map[string]string{
		DEBIAN: "/usr/bin/dpkg --purge %s",
		YUM:    "/usr/bin/yum remove --assumeyes %s",
	}

	// Update command for each system
	updateCmd = map[string]string{
		DEBIAN: "/usr/bin/apt-get update --quiet --quiet",
		YUM:    "/usr/bin/yum clean expire-cache --quiet",
	}

	// Environment variables used for commands on each system
	envCmd = map[string]string{
		DEBIAN: "DEBIAN_FRONTEND=noninteractive NEEDRESTART_MODE=a",
		YUM:    "",
	}
)

// Package maintains our object state
type Package struct {

	// System contains our identified system.
	system string

	// privilegedhelper contains the name of a binary to prefix
	// our commands with, to elevate privileges
	privilegedhelper string
}

// New creates a new instance of this object, attempting to identify the
// system during the initial phase.
func New() *Package {
	p := &Package{}
	p.identify()
	return p
}

// UsePrivilegeHelper is used to ensure that all executed commands
// are prefixed with "sudo ..", "doas ..", or similar.
func (p *Package) UsePrivilegeHelper(cmd string) {
	p.privilegedhelper = cmd
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

	// Add privileges if we need to
	if p.privilegedhelper != "" {
		tmp = p.privilegedhelper + " " + tmp
	}

	// Split
	run, err := shlex.Split(tmp)
	if err != nil {
		return err
	}
	env, err := shlex.Split(envCmd[p.System()])
	if err != nil {
		return err
	}

	// Run the command
	return p.run(run, env)
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
	run, err := shlex.Split(tmp)
	if err != nil {
		return false, err
	}
	env, err := shlex.Split(envCmd[p.System()])
	if err != nil {
		return false, err
	}

	// Run the command
	err = p.run(run, env)

	// No error?  Then the package is installed
	if err == nil {
		return true, nil
	}

	// Error means it isn't.
	return false, nil
}

// Install a single package to the system.
func (p *Package) Install(name []string) error {

	if !p.IsKnown() {
		return fmt.Errorf("failed to recognize system-type")
	}

	// Get the command
	tmp := installCmd[p.System()]
	tmp = strings.ReplaceAll(tmp, "%s", strings.Join(name, " "))

	// Add privileges if we need to
	if p.privilegedhelper != "" {
		tmp = p.privilegedhelper + " " + tmp
	}

	// Show what we're going to run
	log.Printf("[DEBUG] packages:Install will run %s\n", tmp)

	// Split
	run, err := shlex.Split(tmp)
	if err != nil {
		return err
	}
	env, err := shlex.Split(envCmd[p.System()])
	if err != nil {
		return err
	}

	// Run the command
	return p.run(run, env)
}

// Uninstall a single package from the system.
func (p *Package) Uninstall(name []string) error {

	if !p.IsKnown() {
		return fmt.Errorf("failed to recognize system-type")
	}

	// Get the command
	tmp := uninstallCmd[p.System()]
	tmp = strings.ReplaceAll(tmp, "%s", strings.Join(name, " "))

	// Add privileges if we need to
	if p.privilegedhelper != "" {
		tmp = p.privilegedhelper + " " + tmp
	}

	// Show what we're going to run
	log.Printf("[DEBUG] packages:Uninstall will run %s\n", tmp)

	// Split
	run, err := shlex.Split(tmp)
	if err != nil {
		return err
	}
	env, err := shlex.Split(envCmd[p.System()])
	if err != nil {
		return err
	}

	// Run the command
	return p.run(run, env)
}

// run executes the named command and returns an error unless
// the execution launched and the return-code was zero.
func (p *Package) run(run []string, env []string) error {

	// Run
	cmd := exec.Command(run[0], run[1:]...)
	cmd.Env = append(cmd.Environ(), env...)
	if err := cmd.Start(); err != nil {
		return err
	}

	// Wait for completion
	if err := cmd.Wait(); err != nil {

		if exiterr, ok := err.(*exec.ExitError); ok {

			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return fmt.Errorf("exit code for '%s' was %d", strings.Join(run, " "), status.ExitStatus())
			}
		}
	}

	return nil

}
