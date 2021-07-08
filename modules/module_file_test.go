package modules

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/skx/marionette/file"
)

func TestCheck(t *testing.T) {

	f := &FileModule{}

	args := make(map[string]interface{})

	// Missing 'target'
	err := f.Check(args)
	if err == nil {
		t.Fatalf("expected error due to missing target")
	}
	if !strings.Contains(err.Error(), "missing 'target'") {
		t.Fatalf("got error - but wrong one : %s", err)
	}

	// Wrong kind of target
	args["target"] = 3
	err = f.Check(args)
	if err == nil {
		t.Fatalf("expected error due to missing target")
	}
	if !strings.Contains(err.Error(), "failed to convert") {
		t.Fatalf("got error - but wrong one : %s", err)
	}

	// Valid target
	args["target"] = "/foo/bar"
	err = f.Check(args)
	if err != nil {
		t.Fatalf("unexpected error")
	}
}

func TestAbsent(t *testing.T) {

	// Create a temporary file
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		t.Fatalf("create a temporary file failed")
	}

	defer os.Remove(tmpfile.Name())

	// Confirm it exists
	if !file.Exists(tmpfile.Name()) {
		t.Fatalf("file doesn't exist, after creation")
	}

	// Remove it
	args := make(map[string]interface{})
	args["target"] = tmpfile.Name()
	args["state"] = "absent"

	// Run the module
	f := &FileModule{}
	changed, err := f.Execute(args)

	if err != nil {
		t.Fatalf("unexpected error")
	}
	if !changed {
		t.Fatalf("expected a change")
	}

	// The file is gone?
	if file.Exists(tmpfile.Name()) {
		t.Fatalf("File still exists, but should have been removed!")
	}

	// Run the module again to confirm "no change" when asked to remove a file
	// that does not exist.
	changed, err = f.Execute(args)

	if err != nil {
		t.Fatalf("unexpected error")
	}
	if changed {
		t.Fatalf("didn't expect a change, but got one")
	}
}
