package modules

import (
	"fmt"
	"os"
	"strconv"

	"github.com/skx/marionette/config"
	"github.com/skx/marionette/environment"
	"github.com/skx/marionette/file"
)

// DirectoryModule stores our state
type DirectoryModule struct {
	// cfg contains our configuration object.
	cfg *config.Config
}

// Check is part of the module-api, and checks arguments.
func (f *DirectoryModule) Check(args map[string]interface{}) error {

	// Ensure we have a target (i.e. name to operate upon).
	_, ok := args["target"]
	if !ok {
		return fmt.Errorf("missing 'target' parameter")
	}

	target := StringParam(args, "target")
	if target == "" {
		return fmt.Errorf("failed to convert target to string")
	}

	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (f *DirectoryModule) Execute(env *environment.Environment, args map[string]interface{}) (bool, error) {

	// Default to not having changed
	changed := false

	// Get the target
	target := StringParam(args, "target")

	// We assume we're creating the directory, but we might be removing it.
	state := StringParam(args, "state")
	if state == "" {
		state = "present"
	}

	// Remove the directory, if we should.
	if state == "absent" {

		// Does it exist?
		if !file.Exists(target) {
			// Does not exist - nothing to do
			return false, nil
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
	if !file.Exists(target) {

		// make the directory hierarchy
		er := os.MkdirAll(target, os.FileMode(modeI))
		if er != nil {
			return false, er
		}

		changed = true
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

	// If we created the directory it will have the correct
	// mode, but if it was already present with the wrong value
	// we must fix it.
	change, err := file.ChangeMode(target, mode)
	if err != nil {
		return false, err
	}
	if change {
		changed = true
	}

	return changed, nil
}

// init is used to dynamically register our module.
func init() {
	Register("directory", func(cfg *config.Config) ModuleAPI {
		return &DirectoryModule{cfg: cfg}
	})
}
