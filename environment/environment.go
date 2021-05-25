// Package environment is used to store and retrieve variables.
//
// There is one environment which is shared by the driver and the parser.
package environment

import (
	"runtime"
)

// Environment stores our state
type Environment struct {

	// The variables we're holding.
	vars map[string]string
}

// New returns a new Environment object.
//
// The new environment receives some default variable/values, which currently
// include the architecture of the host system and the operating-system upon
// which we're running.
func New() *Environment {
	// Create a new environment
	tmp := &Environment{vars: make(map[string]string)}

	// Set some default values
	tmp.vars["ARCH"] = runtime.GOARCH
	tmp.vars["OS"] = runtime.GOOS

	return tmp
}

// Set updates the environment to store the given value against the
// specified key.
//
// Any previously-existing value will be overwritten.
func (e *Environment) Set(key string, val string) {
	e.vars[key] = val
}

// Get retrieves the named value from the environment, along with a boolean
// value to indicate whether the retrieval was successful.
func (e *Environment) Get(key string) (string, bool) {
	val, ok := e.vars[key]
	return val, ok
}
