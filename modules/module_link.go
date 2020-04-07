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

	required := []string{"source", "target"}

	for _, key := range required {
		_, ok := args[key]
		if !ok {
			return fmt.Errorf("missing '%s' parameter", key)
		}
	}

	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (f *LinkModule) Execute(args map[string]interface{}) (bool, error) {

	// Get the target
	target := StringParam(args, "target")
	if target == "" {
		return false, fmt.Errorf("failed to convert target to string")
	}

	// Get the source
	source := StringParam(args, "source")
	if source == "" {
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
		var originFile string
		originFile, err = os.Readlink(target)
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
