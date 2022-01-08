//go:build linux
// +build linux

package modules

import (
	"strings"
	"testing"

	"github.com/skx/marionette/config"
	"github.com/skx/marionette/environment"
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

	// Valid target
	args["command"] = "/usr/bin/uptime"
	err = s.Check(args)
	if err != nil {
		t.Fatalf("unexpected error")
	}
}

func TestShell(t *testing.T) {

	// Quiet and Verbose
	sQuiet := &ShellModule{cfg: &config.Config{Verbose: false}}
	sVerbose := &ShellModule{cfg: &config.Config{Verbose: true}}

	env := environment.New()

	// Arguments
	args := make(map[string]interface{})

	// Run with no arguments to see an error
	changed, err := sQuiet.Execute(env, args)
	if changed {
		t.Fatalf("unexpected change")
	}
	if err == nil {
		t.Fatalf("Expected error with no command")
	}
	if !strings.Contains(err.Error(), "missing 'command'") {
		t.Fatalf("Got error, but wrong one")
	}

	// Now setup a command to run - a harmless one!
	args["command"] = "true"

	changed, err = sQuiet.Execute(env, args)

	if !changed {
		t.Fatalf("Expected to see changed result")
	}
	if err != nil {
		t.Fatalf("unexpected error:%s", err.Error())
	}

	changed, err = sVerbose.Execute(env, args)

	if !changed {
		t.Fatalf("Expected to see changed result")
	}
	if err != nil {
		t.Fatalf("unexpected error:%s", err.Error())
	}

	// Try a command with redirection
	args["command"] = "true >/dev/null"
	changed, err = sQuiet.Execute(env, args)

	if !changed {
		t.Fatalf("Expected to see changed result")
	}
	if err != nil {
		t.Fatalf("unexpected error:%s", err.Error())
	}

	// Now finally try a command that doesn't exist.
	args["command"] = "/this/does/not/exist"
	changed, err = sQuiet.Execute(env, args)

	if err == nil {
		t.Fatalf("wanted error running missing command, got none")
	}
	if changed {
		t.Fatalf("Didn't expect to see changed result")
	}
}
