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
	required := []string{"login", "state"}

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

	// Ensure state is one of "present"/"absent"
	state := StringParam(args, "state")
	if state == "absent" {
		return nil
	}
	if state == "present" {
		return nil
	}

	return fmt.Errorf("state must be one of 'absent' or 'present'")
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
	login := StringParam(args, "login")
	state := StringParam(args, "state")

	// TODO: Does the username have sane characters?

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
	// TODO: Does the shell have sane characters?
	shell := StringParam(args, "shell")

	// Setup default shell, if nothing was specified.
	if shell == "" {
		shell = "/bin/bash"
	}

	// Create the user
	cmdArgs := []string{"useradd", "-s", shell, login}
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
