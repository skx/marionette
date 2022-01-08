//go:build !windows

package modules

import (
	"fmt"
	"log"
	"os/exec"
	"os/user"
	"syscall"

	"github.com/skx/marionette/environment"
)

// Execute is part of the module-api, and is invoked to run a rule.
func (g *GroupModule) Execute(env *environment.Environment, args map[string]interface{}) (bool, error) {

	// Group/State - we've already confirmed these are valid
	// in our check function.
	group := StringParam(args, "group")
	state := StringParam(args, "state")

	// Does the group already exist?
	if g.groupExists(group) {

		if state == "present" {

			// We're supposed to create the group, but it
			// already exists.  Do nothing.
			return false, nil
		}
		if state == "absent" {

			// remove the group
			err := g.removeGroup(group)
			return true, err
		}
	}

	if state == "absent" {

		// The group is not present, and we're supposed to remove
		// it.  Do nothing.
		return false, nil
	}

	// Create the group
	ret := g.createUser(args)

	// error?
	if ret != nil {
		return false, ret
	}

	return true, nil
}

// groupExists tests if the given group exists.
func (g *GroupModule) groupExists(group string) bool {

	_, err := user.LookupGroup(group)

	return err == nil
}

// createGroup creates a local group.
func (g *GroupModule) createUser(args map[string]interface{}) error {

	group := StringParam(args, "group")

	// The creation command
	cmdArgs := []string{"groupadd", group}

	// Show what we're doing
	log.Printf("[DEBUG] Running %s", cmdArgs)

	// Run it
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	if err := cmd.Start(); err != nil {
		return err
	}

	// Wait for completion
	if err := cmd.Wait(); err != nil {

		if exiterr, ok := err.(*exec.ExitError); ok {

			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return fmt.Errorf("exit code was %d", status.ExitStatus())
			}
		}
	}

	return nil
}

// removeGroup removes the local group
func (g *GroupModule) removeGroup(group string) error {

	// The removal command
	cmdArgs := []string{"groupdel", group}

	// Show what we're doing
	log.Printf("[DEBUG] Running %s", cmdArgs)

	// Run it
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	if err := cmd.Start(); err != nil {
		return err
	}

	// Wait for completion
	if err := cmd.Wait(); err != nil {

		if exiterr, ok := err.(*exec.ExitError); ok {

			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return fmt.Errorf("exit code was %d", status.ExitStatus())
			}
		}
	}

	return nil
}
