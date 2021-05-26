package modules

import (
	"strings"
	"testing"
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

	// Wrong kind of target
	args["target"] = 3
	err = d.Check(args)
	if err == nil {
		t.Fatalf("expected error due to missing target")
	}
	if !strings.Contains(err.Error(), "failed to convert") {
		t.Fatalf("got error - but wrong one : %s", err)
	}

	// Valid target
	args["target"] = "/foo/bar"
	err = d.Check(args)
	if err != nil {
		t.Fatalf("unexpected error")
	}
}
