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

	// env holds our environment
	env *environment.Environment
}

// Check is part of the module-api, and checks arguments.
func (f *DirectoryModule) Check(args map[string]interface{}) error {

	// Ensure we have a target (i.e. name to operate upon).
	_, ok := args["target"]
	if !ok {
		return fmt.Errorf("missing 'target' parameter")
	}

	// Target may be either a string or an array, so we don't test
	// the type here.
	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (f *DirectoryModule) Execute(args map[string]interface{}) (bool, error) {

	// Ensure we have one or more targets to process
	_, ok := args["target"]
	if !ok {
		return false, fmt.Errorf("missing 'target' parameter")
	}

	// Get the argument
	arg := args["target"]

	// if it is a string process it
	str, ok := arg.(string)
	if ok {
		return f.executeSingle(str, args)
	}

	// default to not being changed
	changed := false

	// otherwise we assume it is an array
	dirs := arg.([]string)

	// process each argument
	for _, arg := range dirs {

		// we'll see if it changed
		//
		// if any single directory resulted in a change then
		// our return value will reflect that
		change, err := f.executeSingle(arg, args)

		// but first process any error
		if err != nil {
			return false, err
		}

		// record the change
		if change {
			changed = true
		}
	}

	return changed, nil
}

// executeSingle executes a single directory action.
//
// All parameters are available, as is the single target of this function.
func (f *DirectoryModule) executeSingle(target string, args map[string]interface{}) (bool, error) {

	// Default to not having changed
	changed := false

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
	Register("directory", func(cfg *config.Config, env *environment.Environment) ModuleAPI {
		return &DirectoryModule{cfg: cfg,
			env: env}
	})
}
