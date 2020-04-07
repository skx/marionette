package modules

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/skx/marionette/file"
)

// FileModule stores our state
type FileModule struct {
}

// Check is part of the module-api, and checks arguments.
func (f *FileModule) Check(args map[string]interface{}) error {

	// Ensure we have a target (i.e. name to operate upon).
	_, ok := args["target"]
	if !ok {
		return fmt.Errorf("missing 'target' parameter")
	}

	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (f *FileModule) Execute(args map[string]interface{}) (bool, error) {

	var ret bool
	var err error

	// Get the target
	target := StringParam(args, "target")
	if target == "" {
		return false, fmt.Errorf("failed to convert target to string")
	}

	// We assume we're creating the file, but we might be removing it.
	state := StringParam(args, "state")
	if state == "" {
		state = "present"
	}

	// Remove the directory, if we should.
	if state == "absent" {

		// Does it exist?
		if file.Exists(target) {
			err = os.Remove(target)
			return true, err
		}

		// Didn't exist, nothing to change.
		return false, nil
	}

	// If we have a source file, copy
	source := StringParam(args, "source")
	if source != "" {
		ret, err = f.CopyFile(source, target)
		if err != nil {
			return ret, err
		}
	}

	// If we have a content to set, then use it
	content := StringParam(args, "content")
	if content != "" {
		ret, err = f.CreateFile(target, content)
		if err != nil {
			return ret, err
		}
	}

	// If we have a source URL, fetch.
	srcURL := StringParam(args, "source_url")
	if srcURL != "" {
		ret, err = f.FetchURL(srcURL, target)
		if err != nil {
			return ret, err
		}
	}

	//
	// Now we can change the owner/group
	//

	// Get the mode, if any.  We'll have a default here.
	mode := StringParam(args, "mode")
	if mode == "" {
		mode = "0755"
	}

	// User and group changes
	owner := StringParam(args, "owner")
	if owner != "" {
		changed := false
		changed, err = file.ChangeOwner(target, owner)
		if err != nil {
			return false, err
		}
		if changed {
			ret = true
		}
	}
	group := StringParam(args, "group")
	if group != "" {
		changed := false
		changed, err = file.ChangeGroup(target, group)
		if err != nil {
			return false, err
		}
		if changed {
			ret = true
		}
	}

	// The current mode.
	modeI, _ := strconv.ParseInt(mode, 8, 64)

	// Get the details of the file, so we can see if we need
	// to change owner, group, and mode.
	info, err := os.Stat(target)
	if err != nil {
		return false, err
	}

	if mode != "" && (info.Mode().Perm() != os.FileMode(modeI)) {
		err = os.Chmod(target, os.FileMode(modeI))
		if err != nil {
			return false, err
		}
		ret = true
	}

	return ret, err
}

// CopyFile copies the source file to the destination, returning if we changed
// the contents.
func (f *FileModule) CopyFile(src string, dst string) (bool, error) {

	// File doesn't exist - copy it
	if !file.Exists(dst) {
		err := file.Copy(src, dst)
		return true, err
	}

	// Are the files identical?
	identical, err := file.Identical(src, dst)
	if err != nil {
		return false, err
	}

	// If identical no change
	if identical {
		return false, err
	}

	// Since they differ we refresh and that's a change
	err = file.Copy(src, dst)
	return true, err
}

// FetchURL retrieves the contents of the remote URL and saves them to
// the given file.  If the contents are identical no change is reported.
func (f *FileModule) FetchURL(url string, dst string) (bool, error) {

	// Download to temporary file
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		return false, nil
	}
	defer os.Remove(tmpfile.Name())

	// Get the remote URL
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(tmpfile, resp.Body)
	if err != nil {
		return false, err
	}

	// File doesn't exist - copy it
	if !file.Exists(dst) {
		err = file.Copy(tmpfile.Name(), dst)
		return true, err
	}

	// OK file does exist.  Compare contents
	identical, err := file.Identical(tmpfile.Name(), dst)
	if err != nil {
		return false, err
	}

	// hashes are identical?  No change
	if identical {
		return false, nil
	}

	// otherwise change
	err = file.Copy(tmpfile.Name(), dst)
	return true, err
}

// CreateFile writes the given content to the named file.
// If the contents are identical no change is reported.
func (f *FileModule) CreateFile(dst string, content string) (bool, error) {

	// Create a temporary file
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		return false, nil
	}
	defer os.Remove(tmpfile.Name())

	// Write to it.
	ioutil.WriteFile(tmpfile.Name(), []byte(content), 0644)

	// File doesn't exist - copy it
	if !file.Exists(dst) {
		err = file.Copy(tmpfile.Name(), dst)
		return true, err
	}

	// Are the two files identical?
	identical, err := file.Identical(tmpfile.Name(), dst)
	if err != nil {
		return false, err
	}

	// hashes are identical?  No change
	if identical {
		return false, nil
	}

	// otherwise change
	err = file.Copy(tmpfile.Name(), dst)
	return true, err
}

// init is used to dynamically register our module.
func init() {
	Register("file", func() ModuleAPI {
		return &FileModule{}
	})
}
