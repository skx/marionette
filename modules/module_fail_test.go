package modules

import (
	"strings"
	"testing"

	"github.com/skx/marionette/environment"
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

	env := environment.New()

	// Setup params
	args := make(map[string]interface{})

	changed, err := f.Execute(env, args)
	if err == nil {
		t.Fatalf("expected error, got none")
	}
	if !strings.Contains(err.Error(), "missing 'message'") {
		t.Fatalf("got error - but wrong one : %s", err)
	}
	if changed {
		t.Fatalf("unexpected change")
	}

	// Setup a message
	args["message"] = "I have no cake"

	changed, err = f.Execute(env, args)
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
