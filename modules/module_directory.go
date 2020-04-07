package modules

import (
	"fmt"
	"os"
	"strconv"

	"github.com/skx/marionette/file"
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
	target := StringParam(args, "target")
	if target == "" {
		return false, fmt.Errorf("failed to convert target to string")
	}

	// We assume we're creating the directory, but we might be removing it.
	state := StringParam(args, "state")
	if state == "" {
		state = "present"
	}

	// Remove the directory, if we should.
	if state == "absent" {

		// Does it exist?
		if _, err := os.Stat(target); err != nil {
			if os.IsNotExist(err) {

				// Does not exist
				return false, nil
			}
		}

		// OK remove
		os.RemoveAll(target)
		return true, nil
	}

	// Get the mode, if any.  We'll have a default here.
	mode := StringParam(args, "mode")
	if mode == "" {
		mode = "0755"
	}

	// Convert mode to int
	modeI, _ := strconv.ParseInt(mode, 8, 64)

	// Create the directory, if it is missing, with the correct mode.
	if _, err := os.Stat(target); err != nil {
		if os.IsNotExist(err) {
			// make it
			os.MkdirAll(target, os.FileMode(modeI))
			changed = true
		} else {
			// Error running the stat
			return false, err
		}
	}

	// Get the details of the directory, so we can see if we need
	// to change owner, group, and mode.
	info, err := os.Stat(target)
	if err != nil {
		return false, err
	}

	// User and group changes
	owner := StringParam(args, "owner")
	group := StringParam(args, "group")

	// User and group changes
	if owner != "" {
		change, err := file.ChangeOwner(target, owner)
		if err != nil {
			return false, err
		}
		if change {
			changed = true
		}
	}
	if group != "" {
		change, err := file.ChangeGroup(target, group)
		if err != nil {
			return false, err
		}
		if change {
			changed = true
		}
	}

	// The current mode.
	if mode != "" && (info.Mode().Perm() != os.FileMode(modeI)) {
		err := os.Chmod(target, os.FileMode(modeI))
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
