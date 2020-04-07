package modules

import (
	"fmt"
	"os"

	"github.com/skx/marionette/file"
)

// LinkModule stores our state
type LinkModule struct {
}

// Check is part of the module-api, and checks arguments.
func (f *LinkModule) Check(args map[string]interface{}) error {

	// Ensure we have a target (i.e. name to operate upon).
	_, ok := args["target"]
	if !ok {
		return fmt.Errorf("missing 'target' parameter")
	}

	// Ensure we have a source (i.e. name to operate upon).
	_, ok = args["source"]
	if !ok {
		return fmt.Errorf("missing 'source' parameter")
	}

	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (f *LinkModule) Execute(args map[string]interface{}) (bool, error) {

	// Get the target
	t := args["target"]
	target, ok := t.(string)
	if !ok {
		return false, fmt.Errorf("failed to convert target to string")
	}

	// Get the source
	s := args["source"]
	source, ok := s.(string)
	if !ok {
		return false, fmt.Errorf("failed to convert source to string")
	}

	// If the target doesn't exist we create the link.
	if !file.Exists(target) {
		err := os.Symlink(source, target)
		return true, err
	}

	// If the target does exist see if it is correct.
	fileInfo, err := os.Lstat(target)

	if err != nil {
		return false, err
	}

	// Is it a symlink?
	if fileInfo.Mode()&os.ModeSymlink != 0 {

		// If so get the target file to which it points.
		originFile, err := os.Readlink(target)
		if err != nil {
			return false, err
		}

		// OK we have a target - is it correct?
		if originFile == source {
			return false, nil
		}

		// OK there is a symlink, but it points to the
		// wrong target-file.  Remove it.
		err = os.Remove(target)
		if err != nil {
			return false, err
		}
	} else {

		// We found something that wasn't a symlink.
		//
		// Remove it.
		err = os.Remove(target)
		if err != nil {
			return false, err
		}

	}

	// Fix the symlink.
	err = os.Symlink(source, target)
	return true, err
}

// init is used to dynamically register our module.
func init() {
	Register("link", func() ModuleAPI {
		return &LinkModule{}
	})
}
