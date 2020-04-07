// Package modules contain the implementation of our modules.  Each
// module has a name/type such as "git", "file", etc.  The modules
// each accept an arbitrary set of parameters which are module-specific.
package modules

// ModuleAPI is the interface which all our modules must confirm.
//
// There are only two methods, one to check if the supplied parameters
// make sense, the other to actually execute the rule.
type ModuleAPI interface {

	// Check allows a module to ensures that any mandatory parameters
	// are present, or perform similar setup-work.
	//
	// If no error is returned then the module will be executed later.
	Check(map[string]interface{}) error

	// Execute runs the module with the given arguments.
	//
	// The return value is true if the module made a change
	// and false otherwise.
	//
	Execute(map[string]interface{}) (bool, error)
}
