package modules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/skx/marionette/file"
)

func TestDirectoryCheck(t *testing.T) {

	d := &DirectoryModule{}

	args := make(map[string]interface{})

	// Missing 'target'
	err := d.Check(args)
	if err == nil {
		t.Fatalf("expected error due to missing target")
	}
	if !strings.Contains(err.Error(), "missing 'target'") {
		t.Fatalf("got error - but wrong one : %s", err)
	}

	// Valid target
	args["target"] = "/foo/bar"
	err = d.Check(args)
	if err != nil {
		t.Fatalf("unexpected error")
	}
}

func TestDirectoryMultiple(t *testing.T) {

	// Create a temporary directory
	dir, err := os.MkdirTemp("", "m_d_t")
	if err != nil {
		t.Fatalf("failed to make temporary directory")
	}

	// pair of directories we'll create
	a := filepath.Join(dir, "one")
	b := filepath.Join(dir, "two")

	// Create a bunch of directories
	args := make(map[string]interface{})
	args["target"] = []string{
		a,
		b,
	}

	d := &DirectoryModule{}
	changed, err := d.Execute(args)

	if err != nil {
		t.Fatalf("error making multiple directories:%s", err)
	}
	if !changed {
		t.Fatalf("expected to see a change")
	}

	// Second time around the directories should exist,
	// so we see no change
	changed, err = d.Execute(args)

	if err != nil {
		t.Fatalf("error making multiple directories:%s", err)
	}
	if changed {
		t.Fatalf("expected to see no change when directories exist")
	}

	// Ensure the directories exist
	if !file.Exists(a) {
		t.Fatalf("expected to see directory present!")
	}
	if !file.Exists(b) {
		t.Fatalf("expected to see directory present!")
	}

	// Now remove them
	args["state"] = "absent"
	changed, err = d.Execute(args)

	if err != nil {
		t.Fatalf("error removing multiple directories:%s", err)
	}
	if !changed {
		t.Fatalf("expected to see a change when removing directories")
	}

	if file.Exists(a) {
		t.Fatalf("expected to see no directory present after removal!")
	}
	if file.Exists(b) {
		t.Fatalf("expected to see no directory present after removal!")
	}

	// remove them again - should be no change
	changed, err = d.Execute(args)
	if err != nil {
		t.Fatalf("error removing multiple directories:%s", err)
	}
	if changed {
		t.Fatalf("expected to see no change when removing absent directories")
	}

	// cleanup
	os.RemoveAll(dir)
}

// Issue 104
func TestDirectoryMkdirP(t *testing.T) {

	// Create a temporary directory
	dir, err := os.MkdirTemp("", "t_d_m_p")
	if err != nil {
		t.Fatalf("failed to make temporary directory")
	}

	// the nested directory we'll create
	a := filepath.Join(dir, "one", "two", "three", "four")

	// Create a bunch of directories
	args := make(map[string]interface{})
	args["target"] = []string{
		a,
	}

	d := &DirectoryModule{}
	changed, err := d.Execute(args)

	if err != nil {
		t.Fatalf("error making nested-directories:%s", err)
	}
	if !changed {
		t.Fatalf("expected to see a change")
	}

	// Ensure the directories exist
	if !file.Exists(a) {
		t.Fatalf("expected to see directory present!")
	}

	// Now remove them
	args["state"] = "absent"
	changed, err = d.Execute(args)

	if err != nil {
		t.Fatalf("error removing nested-directories:%s", err)
	}
	if !changed {
		t.Fatalf("expected to see a change when removing directories")
	}

	if file.Exists(a) {
		t.Fatalf("expected to see no directory present after removal!")
	}

	// remove them again - should be no change
	changed, err = d.Execute(args)
	if err != nil {
		t.Fatalf("error removing multiple directories:%s", err)
	}
	if changed {
		t.Fatalf("expected to see no change when removing absent directories")
	}

	// cleanup
	os.RemoveAll(dir)
}
