package environment

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/skx/marionette/file"
	"github.com/skx/marionette/token"
)

// Test built-in values
func TestExpected(t *testing.T) {

	x := New()

	a, aOK := x.Get("ARCH")
	if !aOK {
		t.Fatalf("Failed to get ${ARCH}")
	}
	if a != runtime.GOARCH {
		t.Fatalf("${ARCH} had wrong value != %s", runtime.GOARCH)
	}

	o, oOK := x.Get("OS")
	if !oOK {
		t.Fatalf("Failed to get ${OS}")
	}
	if o != runtime.GOOS {
		t.Fatalf("${OS} had wrong value != %s", runtime.GOOS)
	}
}

// TestSet ensures a value will remain
func TestSet(t *testing.T) {

	e := New()

	// Count the variables
	vars := e.Variables()
	vlen := len(vars)

	// Confirm getting a missing value fails
	_, ok := e.Get("STEVE")
	if ok {
		t.Fatalf("Got value for STEVE, shouldn't have done")
	}

	// Set the value
	e.Set("STEVE", "KEMP")

	// Get it again
	val := ""
	val, ok = e.Get("STEVE")
	if !ok {
		t.Fatalf("After setting the value wasn't available")
	}
	if val != "KEMP" {
		t.Fatalf("Wrong value retrieved")
	}

	if len(e.Variables()) != vlen+1 {
		t.Errorf("After setting variable length didn't increase")
	}
	// Update the value
	e.Set("STEVE", "STEVE")

	// Get it again
	val, ok = e.Get("STEVE")
	if !ok {
		t.Fatalf("After setting the value wasn't available")
	}
	if val != "STEVE" {
		t.Fatalf("Wrong value retrieved, after update")
	}

}

// Test we can get/set environment variables properly
func TestEnvVariable(t *testing.T) {

	// Create a new object
	e := New()

	// Set a variable, via the environment
	os.Setenv("NAME", "world")

	out, err := e.ExpandTokenVariables(token.Token{
		Type:    token.STRING,
		Literal: "Hello, ${NAME}",
	})

	if err != nil {
		t.Fatalf("unexpected error running expansion: %s\n", out)
	}

	if out != "Hello, world" {
		t.Fatalf("Unexpected output: %s", out)
	}

	// Variables take precedence
	e.Set("NAME", "Steve")

	out, err = e.ExpandTokenVariables(token.Token{
		Type:    token.STRING,
		Literal: "Hello, ${NAME}",
	})

	if err != nil {
		t.Fatalf("unexpected error running expansion: %s\n", out)
	}

	if out != "Hello, Steve" {
		t.Fatalf("Unexpected output: %s", out)
	}
}

// Test we can execute commands to expand tokens

func TestCommandExpansion(t *testing.T) {
	// Create a new object
	e := New()

	// Do we have /bin/ls
	if file.Exists("/bin/ls") && file.Exists("/etc/passwd") {

		out, err := e.ExpandTokenVariables(token.Token{
			Type:    token.BACKTICK,
			Literal: "ls /etc/passwd",
		})

		if err != nil {
			t.Fatalf("error running ls: %s", err)
		}
		if !strings.Contains(out, "passwd") {
			t.Fatalf("/bin/ls /etc/passwd had weird output:%s", out)
		}
	}

	// Test running a command that doesn't exist
	cmd := "/no/such/file/directory"
	_, err := e.ExpandTokenVariables(token.Token{
		Type:    token.BACKTICK,
		Literal: cmd,
	})

	if err == nil {
		t.Fatalf("expected error running missing command, got none")
	}
	if !strings.Contains(err.Error(), cmd) {
		t.Fatalf("expected a decent error message, got %s", err)
	}
}
