// Package environment is used to store and retrieve variables
// by our run-time Executor.
package environment

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"

	"github.com/skx/marionette/token"
)

// Environment stores our state
type Environment struct {

	// The variables we're holding.
	vars map[string]string
}

// New returns a new Environment object.
//
// The new environment receives some default variable/values,
// which currently include the architecture of the host system,
// the operating-system upon which we're running, & etc.
func New() *Environment {
	// Create a new environment
	tmp := &Environment{vars: make(map[string]string)}

	// Set some default values
	tmp.vars["ARCH"] = runtime.GOARCH
	tmp.vars["OS"] = runtime.GOOS

	// Default hostname
	tmp.vars["HOSTNAME"] = "unknown"

	// Get the real one, and set it if no errors
	host, err := os.Hostname()
	if err == nil {
		tmp.vars["HOSTNAME"] = host
	}

	// Default username and homedir as empty
	tmp.vars["USERNAME"] = ""
	tmp.vars["HOMEDIR"] = ""

	// Get the real username and homedir, and set it if no errors
	user, err := user.Current()
	if err == nil {
		tmp.vars["USERNAME"] = user.Username
		tmp.vars["HOMEDIR"] = user.HomeDir
	}

	// Log our default variables
	for key, val := range tmp.vars {
		log.Printf("[DEBUG] Set default variable %s -> %s\n", key, val)
	}

	return tmp
}

// Set updates the environment to store the given value against the
// specified key.
//
// Any previously-existing value will be overwritten.
func (e *Environment) Set(key string, val string) {
	e.vars[key] = val
}

// Get retrieves the named value from the environment, along
// with a boolean value to indicate whether the retrieval was
// successful.
func (e *Environment) Get(key string) (string, bool) {
	val, ok := e.vars[key]
	return val, ok
}

// Variables returns all of variables which have been set, as
// well as their values.
//
// This is used such that include-files inherit the variables
// which were already in-scope at the point the inclusion happens.
func (e *Environment) Variables() map[string]string {
	return e.vars
}

// ExpandVariables takes a string which contains embedded
// variable references, such as ${USERNAME}, and expands the
// result.
func (e *Environment) ExpandVariables(input string) string {
	return os.Expand(input, e.expandVariablesMapper)
}

// ExpandTokenVariables is similar to the ExpandVariables, the
// difference is that it uses a token as an input, rather than a string.
//
// This is done so that if a token is used of type `BACKTICK` we can
// execute the appropriate shell command(s).
//
// TODO: Delete Me.
func (e *Environment) ExpandTokenVariables(tok token.Token) (string, error) {

	// Expand any variables
	value := tok.Literal
	value = e.ExpandVariables(value)

	// If we're not a backtick we're done here
	if tok.Type != token.BACKTICK {
		return value, nil
	}

	// Now we need to execute the command and return the value
	// Build up the thing to run, using a shell so that
	// we can handle pipes/redirection.
	toRun := []string{"/bin/bash", "-c", value}

	// Run the command
	cmd := exec.Command(toRun[0], toRun[1:]...)

	// Get the output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running command '%s' %s", value, err.Error())
	}

	// Strip trailing newline.
	ret := strings.TrimSuffix(string(output), "\n")
	return ret, nil
}

// ExpandBacktick is similar to the ExpandVariables, it expands any
// variables within the given string, then executes that as a command.
func (e *Environment) ExpandBacktick(value string) (string, error) {

	// Expand any variables within the command.
	value = e.ExpandVariables(value)

	// Now we need to execute the command and return the value
	// Build up the thing to run, using a shell so that
	// we can handle pipes/redirection.
	toRun := []string{"/bin/bash", "-c", value}

	// Run the command
	cmd := exec.Command(toRun[0], toRun[1:]...)

	// Get the output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running command '%s' %s", value, err.Error())
	}

	// Strip trailing newline.
	ret := strings.TrimSuffix(string(output), "\n")
	return ret, nil
}

// expandVariablesMapper is a helper to expand variables.
//
// ${foo} will be converted to the contents of the variable named foo
// which was created with `let foo = "bar"`, or failing that the contents
// of the environmental variable named `foo`.
//
func (e *Environment) expandVariablesMapper(val string) string {

	// Lookup a variable which exists?
	res, ok := e.Get(val)
	if ok {
		return res
	}

	// Lookup an environmental variable?
	return os.Getenv(val)
}
