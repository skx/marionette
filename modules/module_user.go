package modules

import (
	"fmt"
	"os/exec"
	"os/user"
	"syscall"

	mcfg "github.com/skx/marionette/config"
)

// UserModule stores our state
type UserModule struct {

	// cfg contains our configuration object.
	cfg *mcfg.Config
}

// Check is part of the module-api, and checks arguments.
func (g *UserModule) Check(args map[string]interface{}) error {

	// Required keys for this module
	required := []string{"user", "state"}

	// Ensure they exist.
	for _, key := range required {
		_, ok := args[key]
		if !ok {
			return fmt.Errorf("missing '%s' parameter", key)
		}

		val := StringParam(args, key)
		if val == "" {
			return fmt.Errorf("'%s' wasn't a simple string", key)

		}

	}
	return nil
}

// verbose will show the message if the verbose flag is set
func (g *UserModule) verbose(msg string) {
	if g.cfg.Verbose {
		fmt.Printf("%s\n", msg)
	}
}

// Execute is part of the module-api, and is invoked to run a rule.
func (g *UserModule) Execute(args map[string]interface{}) (bool, error) {

	// User/State - we've already confirmed these are valid
	// in our check function.
	user := StringParam(args, "user")
	state := StringParam(args, "state")

	// Optional arguments
	shell := StringParam(args, "shell")

	// TODO: Does the username have sane characters?
	// TODO: Does the shell have sane characters?

	// Does the user exist?
	if g.userExists(user) {

		// We're supposed to create the user, but it
		// already exists.  Do nothing.
		if state == "present" {
			return false, nil
		}
		if state == "absent" {

			// remove the user
			err := g.removeUser(user)
			return true, err

		}

		return false, fmt.Errorf("Invalid state - only 'absent' or 'present' are supported")
	}

	// User is missing.
	if state == "absent" {
		return false, nil
	}
	if state != "present" {
		return false, fmt.Errorf("Invalid state - only 'absent' or 'present' are supported")
	}

	// Setup default shell, if nothing was specified.
	if shell == "" {
		shell = "/bin/bash"
	}

	// Create the user
	cmdArgs := []string{"useradd", "-s", shell, user}
	g.verbose(fmt.Sprintf("Running %v", cmdArgs))

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)

	if err := cmd.Start(); err != nil {
		return false, err
	}

	// Wait for completion
	if err := cmd.Wait(); err != nil {

		if exiterr, ok := err.(*exec.ExitError); ok {

			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return false, fmt.Errorf("exit code was %d", status.ExitStatus())
			}
		}
	}

	return true, nil
}

// userExists tests if the given user exists.
func (g *UserModule) userExists(login string) bool {

	_, err := user.Lookup(login)
	if err == nil {
		return true
	}
	return false
}

// removeUser removes the local user
func (g *UserModule) removeUser(login string) error {

	// Remove the user
	cmdArgs := []string{"userdel", login}

	g.verbose(fmt.Sprintf("Running %v", cmdArgs))

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

// init is used to dynamically register our module.
func init() {
	Register("user", func(cfg *mcfg.Config) ModuleAPI {
		return &UserModule{cfg: cfg}
	})
}
