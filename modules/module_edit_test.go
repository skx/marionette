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
	changed, err = e.Execute(args)
	if err == nil {
		t.Fatalf("expected error, got none")
	}

	// Remove the temporary file, and confirm we get somethign similar
	os.Remove(tmpfile.Name())
	changed, err = e.Execute(args)
	if err != nil {
		t.Fatalf("didn't expect error, got one")
	}

}
