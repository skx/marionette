package modules

import (
	"strings"
	"testing"
)

func TestPackageCheck(t *testing.T) {

	p := &PackageModule{}

	args := make(map[string]interface{})

	// Missing 'package'
	err := p.Check(args)
	if err == nil {
		t.Fatalf("expected error due to missing package")
	}
	if !strings.Contains(err.Error(), "missing 'package'") {
		t.Fatalf("got error - but wrong one : %s", err)
	}

	args["package"] = []string{"bash", "curl"}

	// state can be either "installed" or "absent"
	valid := []string{"installed", "absent"}
	for _, state := range valid {
		args["state"] = state

		err = p.Check(args)

		if err != nil {
			t.Fatalf("unexpected error: %s", err.Error())
		}
	}

	// Confirm a different one breaks
	args["state"] = "removed"
	err = p.Check(args)

	if err == nil {
		t.Fatalf("expected error, got none")
	}
	if !strings.Contains(err.Error(), "package state must be either") {
		t.Fatalf("got error, but not the correct one")
	}
}
