// Package environment is used to store and retrieve variables
// by our run-time Executor.
package environment

import (
	"log"
	"os"
	"os/user"
	"runtime"
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
