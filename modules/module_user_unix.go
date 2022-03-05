//go:build !darwin && !windows

package modules

import (
	"fmt"
	"log"
	"os/exec"
	"os/user"
	"syscall"
)

// Execute is part of the module-api, and is invoked to run a rule.
func (g *UserModule) Execute(args map[string]interface{}) (bool, error) {

	// User/State - we've already confirmed these are valid
	// in our check function.
	login := StringParam(args, "login")
	state := StringParam(args, "state")

	// Does the user exist?
	if g.userExists(login) {

		if state == "present" {

			// We're supposed to create the user, but it
			// already exists.  Do nothing.
			return false, nil
		}
		if state == "absent" {

			// remove the user
			err := g.removeUser(login)
			return true, err
		}
	}

	if state == "absent" {

		// The user is not present, and we're supposed to remove
		// it.  Do nothing.
		return false, nil
	}

	// Create the user
	ret := g.createUser(args)

	// error?
	if ret != nil {
		return false, ret
	}

	return true, nil
}

// userExists tests if the given user exists.
func (g *UserModule) userExists(login string) bool {

	_, err := user.Lookup(login)

	return err == nil
}

// createUser creates a local user.
func (g *UserModule) createUser(args map[string]interface{}) error {

	login := StringParam(args, "login")

	// Optional arguments
	shell := StringParam(args, "shell")

	// Setup default shell, if nothing was specified.
	if shell == "" {
		shell = "/bin/bash"
	}

	// The user-creation command
	cmdArgs := []string{"useradd", "--shell", shell, login}

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

// removeUser removes the local user
func (g *UserModule) removeUser(login string) error {

	// The user-removal command
	cmdArgs := []string{"userdel", login}

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
