package modules

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"text/template"

	"github.com/skx/marionette/config"
	"github.com/skx/marionette/environment"
	"github.com/skx/marionette/file"
)

// FileModule stores our state
type FileModule struct {

	// cfg contains our configuration object.
	cfg *config.Config

	// env contains the environment.
	env *environment.Environment
}

// Check is part of the module-api, and checks arguments.
func (f *FileModule) Check(args map[string]interface{}) error {

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
func (f *FileModule) Execute(env *environment.Environment, args map[string]interface{}) (bool, error) {

	f.env = env

	var ret bool
	var err error

	// Get the target (i.e. file/directory we're operating upon.)
	target := StringParam(args, "target")

	// We assume we're creating the file, but we might be removing it.
	state := StringParam(args, "state")
	if state == "" {
		state = "present"
	}

	//
	// Now we start to handle the request.
	//
	// Remove the file/directory, if we should.
	if state == "absent" {
		return f.removeFile(target)
	}

	//
	// At this point we're going to create/update the file
	// via one of our support options.
	//
	// Go do that, then once that is complete we can update
	// the owner/group/mode, etc.
	//
	ret, err = f.populateFile(target, args)
	if err != nil {
		return ret, err
	}

	// File permission changes
	mode := StringParam(args, "mode")
	if mode != "" {
		var changed bool
		changed, err = file.ChangeMode(target, mode)
		if err != nil {
			return false, err
		}
		if changed {
			ret = true
		}
	}

	// User and group changes
	owner := StringParam(args, "owner")
	if owner != "" {
		var changed bool
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
		var changed bool
		changed, err = file.ChangeGroup(target, group)
		if err != nil {
			return false, err
		}
		if changed {
			ret = true
		}
	}

	return ret, err
}

// removeFile removes the named file, returning whether a change
// was made or not
func (f *FileModule) removeFile(target string) (bool, error) {

	// Does it exist?
	if file.Exists(target) {
		err := os.Remove(target)
		return true, err
	}

	// Didn't exist, nothing to change.
	return false, nil
}

// populateFile is designed to create/update the file contents via one
// of our supported methods.
func (f *FileModule) populateFile(target string, args map[string]interface{}) (bool, error) {

	var ret bool
	var err error

	// If we have a source file, copy that into place
	source := StringParam(args, "source")
	if source != "" {

		ret, err = f.CopyFile(source, target)
		return ret, err
	}

	// If we have a template file, render it.
	template := StringParam(args, "template")
	if template != "" {
		ret, err = f.CopyTemplateFile(template, target)
		return ret, err
	}

	// If we have a content to set, then use it.
	content := StringParam(args, "content")
	if content != "" {
		ret, err = f.CreateFile(target, content)
		return ret, err
	}

	// If we have a source URL, fetch.
	srcURL := StringParam(args, "source_url")
	if srcURL != "" {
		ret, err = f.FetchURL(srcURL, target)
		return ret, err
	}

	return ret, fmt.Errorf("neither 'content', 'source', 'source_url', or 'template' were specified")
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

// CopyTemplateFile copies the template file to the destination, rendering the
// template and returning if we changed the contents.
func (f *FileModule) CopyTemplateFile(src string, dst string) (bool, error) {

	// Create a temporary file to write the rendered template to
	tmpfile, err := ioutil.TempFile("", "marionette-")
	if err != nil {
		return false, nil
	}
	defer os.Remove(tmpfile.Name())

	// Parse the template file
	tpl, err := template.ParseFiles(src)
	if err != nil {
		return false, err
	}

	// Render the template, writing to the temp file
	err = tpl.Execute(tmpfile, f.env.Variables())
	if err != nil {
		return false, err
	}

	return f.CopyFile(tmpfile.Name(), dst)
}

// FetchURL retrieves the contents of the remote URL and saves them to
// the given file.  If the contents are identical no change is reported.
func (f *FileModule) FetchURL(url string, dst string) (bool, error) {

	// Download to temporary file
	tmpfile, err := ioutil.TempFile("", "marionette-")
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

	return f.CopyFile(tmpfile.Name(), dst)
}

// CreateFile writes the given content to the named file.
// If the contents are identical no change is reported.
func (f *FileModule) CreateFile(dst string, content string) (bool, error) {

	// Create a temporary file
	tmpfile, err := ioutil.TempFile("", "marionette-")
	if err != nil {
		return false, nil
	}
	defer os.Remove(tmpfile.Name())

	// Write to it.
	err = ioutil.WriteFile(tmpfile.Name(), []byte(content), 0644)
	if err != nil {
		return false, err
	}

	return f.CopyFile(tmpfile.Name(), dst)
}

// init is used to dynamically register our module.
func init() {
	Register("file", func(cfg *config.Config) ModuleAPI {
		return &FileModule{cfg: cfg}
	})
}
