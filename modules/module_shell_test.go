package modules

import (
	"strings"
	"testing"
)

func TestShellCheck(t *testing.T) {

	s := &ShellModule{}

	args := make(map[string]interface{})

	// Missing 'command'
	err := s.Check(args)
	if err == nil {
		t.Fatalf("expected error due to missing command")
	}
	if !strings.Contains(err.Error(), "missing 'command'") {
		t.Fatalf("got error - but wrong one : %s", err)
	}

	// Wrong kind of target
	args["command"] = 3
	err = s.Check(args)
	if err == nil {
		t.Fatalf("expected error due to missing command")
	}
	if !strings.Contains(err.Error(), "failed to convert") {
		t.Fatalf("got error - but wrong one : %s", err)
	}

	// Valid target
	args["command"] = "/usr/bin/uptime"
	err = s.Check(args)
	if err != nil {
		t.Fatalf("unexpected error")
	}
}
