package modules

import (
	"fmt"
	"os"
	"strconv"
)

// DirectoryModule stores our state
type DirectoryModule struct {
}

// Check is part of the module-api, and checks arguments.
func (f *DirectoryModule) Check(args map[string]interface{}) error {

	// Ensure we have a target (i.e. name to operate upon).
	_, ok := args["target"]
	if !ok {
		return fmt.Errorf("missing 'target' parameter")
	}

	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (f *DirectoryModule) Execute(args map[string]interface{}) (bool, error) {

	// Default to not having changed
	changed := false

	// Get the target
	target := args["target"]
	str, ok := target.(string)
	if !ok {
		return false, fmt.Errorf("failed to convert target to string")
	}

	// Get the mode, if any.  We'll have a default here.
	mode := "0755"
	mode_param, mode_param_present := args["mode"]
	if mode_param_present {
		m, ok := mode_param.(string)
		if ok {
			mode = m
		}
	}

	// Convert mode to int
	mode_i, _ := strconv.ParseInt(mode, 8, 64)

	// Create the directory, if it is missing.
	if _, err := os.Stat(str); err != nil {
		if os.IsNotExist(err) {
			// make it
			os.MkdirAll(str, os.FileMode(mode_i))
			changed = true
		} else {
			// Error running the stat
			return false, err
		}
	}

	// Get the details of the directory, so we can see if we need
	// to change owner (TODO) group (TODO) and mode.
	info, err := os.Stat(str)
	if err != nil {
		return false, err
	}

	// The current mode.
	if mode_param_present && (info.Mode().Perm() != os.FileMode(mode_i)) {
		err := os.Chmod(str, os.FileMode(mode_i))
		if err != nil {
			return false, err
		}
		changed = true
	}

	return changed, nil
}

// init is used to dynamically register our module.
func init() {
	Register("directory", func() ModuleAPI {
		return &DirectoryModule{}
	})
}
