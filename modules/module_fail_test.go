package modules

import (
	"strings"
	"testing"
)

func TestFailCheck(t *testing.T) {

	f := &FailModule{}

	args := make(map[string]interface{})

	// Missing 'message'
	err := f.Check(args)
	if err == nil {
		t.Fatalf("expected error due to missing message")
	}
	if !strings.Contains(err.Error(), "missing 'message'") {
		t.Fatalf("got error - but wrong one : %s", err)
	}

	// Wrong kind of target
	args["message"] = 3
	err = f.Check(args)
	if err == nil {
		t.Fatalf("expected error due to missing message")
	}
	if !strings.Contains(err.Error(), "failed to convert") {
		t.Fatalf("got error - but wrong one : %s", err)
	}

	// Valid target
	args["message"] = "OK"
	err = f.Check(args)
	if err != nil {
		t.Fatalf("unexpected error")
	}
}

func TestFail(t *testing.T) {

	f := &FailModule{}

	// Append my name
	args := make(map[string]interface{})
	args["message"] = "I have no cake"

	changed, err := f.Execute(args)
	if err == nil {
		t.Fatalf("expected error, got none")
	}
	if changed {
		t.Fatalf("unexpected change")
	}
	if !strings.Contains(err.Error(), "I have no cake") {
		t.Fatalf("failure message was unexpected")
	}
}
