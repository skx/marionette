package modules

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/skx/marionette/file"
)

func TestEditCheck(t *testing.T) {

	e := &EditModule{}

	args := make(map[string]interface{})

	// Missing 'target'
	err := e.Check(args)
	if err == nil {
		t.Fatalf("expected error due to missing target")
	}
	if !strings.Contains(err.Error(), "missing 'target'") {
		t.Fatalf("got error - but wrong one : %s", err)
	}

	// Wrong kind of target
	args["target"] = 3
	err = e.Check(args)
	if err == nil {
		t.Fatalf("expected error due to missing target")
	}
	if !strings.Contains(err.Error(), "failed to convert") {
		t.Fatalf("got error - but wrong one : %s", err)
	}

	// Valid target
	args["target"] = "/foo/bar"
	err = e.Check(args)
	if err != nil {
		t.Fatalf("unexpected error")
	}
}

func TestEditAppend(t *testing.T) {

	// create a temporary file
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		t.Fatalf("create a temporary file failed")
	}

	// delete it
	os.Remove(tmpfile.Name())

	e := &EditModule{}

	// Append my name
	args := make(map[string]interface{})
	args["target"] = tmpfile.Name()
	args["append_if_missing"] = "Steve Kemp"

	changed, err := e.Execute(args)
	if err != nil {
		t.Fatalf("error changing file")
	}
	if !changed {
		t.Fatalf("expected file change, got none")
	}

	// If the file doesn't exist now that's a bug
	if !file.Exists(tmpfile.Name()) {
		t.Fatalf("file doesn't exist")
	}

	// Get the file size
	var size int64

	size, err = file.Size(tmpfile.Name())
	if err != nil {
		t.Fatalf("error getting file size")
	}

	// Call again
	changed, err = e.Execute(args)
	if err != nil {
		t.Fatalf("error changing file")
	}
	if changed {
		t.Fatalf("didn't expect file change, got one")
	}

	// file size shouldn't have changed
	var newSize int64
	newSize, err = file.Size(tmpfile.Name())
	if err != nil {
		t.Fatalf("error getting file size")
	}

	if newSize != size {
		t.Fatalf("file size changed!")
	}

	// Finally append "Test"
	args["append_if_missing"] = "Test"
	changed, err = e.Execute(args)
	if err != nil {
		t.Fatalf("error changing file")
	}
	if !changed {
		t.Fatalf("expected file change, got none")
	}

	// And confirm new size is four (+newline) bytes longer
	newSize, err = file.Size(tmpfile.Name())
	if err != nil {
		t.Fatalf("error getting file size")
	}

	if newSize != (size + 5) {
		t.Fatalf("file size mismatch!")
	}
	os.Remove(tmpfile.Name())
}

func TestEditRemove(t *testing.T) {

	// create a temporary file
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		t.Fatalf("create a temporary file failed")
	}

	// Write the input
	_, err = tmpfile.Write([]byte("# This is a comment\n# So is this\n"))
	if err != nil {
		t.Fatalf("error writing temporary file")
	}

	e := &EditModule{}

	// Remove all lines matching "^#" in the temporary file
	args := make(map[string]interface{})
	args["target"] = tmpfile.Name()
	args["remove_lines"] = "^#"

	// Make the change
	changed, err := e.Execute(args)
	if err != nil {
		t.Fatalf("unexpected error")
	}
	if !changed {
		t.Fatalf("expected change, but got none")
	}

	// Second time nothing should happen
	changed, err = e.Execute(args)
	if err != nil {
		t.Fatalf("unexpected error")
	}
	if changed {
		t.Fatalf("unexpected change, nothing should happen")
	}

	// Confirm the file is zero-sized
	size, er := file.Size(tmpfile.Name())
	if er != nil {
		t.Fatalf("error getting file size")
	}
	if size != 0 {
		t.Fatalf("the edit didn't work")
	}

	// Now test that an invalid regexp is taken
	args["remove_lines"] = "*"
	_, err = e.Execute(args)
	if err == nil {
		t.Fatalf("expected error, got none")
	}

	// Remove the temporary file, and confirm we get somethign similar
	os.Remove(tmpfile.Name())
	_, err = e.Execute(args)
	if err != nil {
		t.Fatalf("didn't expect error, got one")
	}

}
