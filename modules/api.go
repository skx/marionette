// Package modules contain the implementation of our modules.  Each
// module has a name/type such as "git", "file", etc.  The modules
// each accept an arbitrary set of parameters which are module-specific.
package modules

import (
	"github.com/skx/marionette/environment"
)

// ModuleAPI is the interface to which all of our modules must conform.
//
// There are only two methods, one to check if the supplied parameters
// make sense, the other to actually execute the rule.
type ModuleAPI interface {

	// Check allows a module to ensures that any mandatory parameters
	// are present, or perform similar setup-work.
	//
	// If no error is returned then the module will be executed later
	// via a call to Execute.
	Check(map[string]interface{}) error

	// Execute runs the module with the given arguments.
	//
	// The return value is true if the module made a change
	// and false otherwise.
	Execute(*environment.Environment, map[string]interface{}) (bool, error)
}

// StringParam returns the named parameter, as a string, from the map.
//
// If the parameter was not present an empty array is returned.
func StringParam(vars map[string]interface{}, param string) string {

	// Get the value
	val, ok := vars[param]
	if !ok {
		return ""
	}

	// Can it be cast into a string?
	str, valid := val.(string)
	if valid {
		return str
	}

	// OK not a string parameter
	return ""
}

// ArrayParam returns the named parameter, as an array, from the map.
//
// If the parameter was not present an empty array is returned.
func ArrayParam(vars map[string]interface{}, param string) []string {

	var empty []string

	// Get the value
	val, ok := vars[param]
	if !ok {
		return empty
	}

	// Can it be cast into a string array?
	strs, valid := val.([]string)
	if valid {
		return strs
	}

	// OK not a string parameter
	return empty
}
