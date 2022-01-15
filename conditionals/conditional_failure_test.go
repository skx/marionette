//go:build !darwin && !windows

package conditionals

import (
	"testing"

	"github.com/skx/marionette/file"
)

func TestFailureArgs(t *testing.T) {

	_, err := Failure([]string{})
	if err == nil {
		t.Fatalf("expect an error with zero args, got none")
	}

	_, err = Failure([]string{"foo"})
	if err != nil {
		t.Fatalf("expected no error with one arg, got %s", err)
	}

	_, err = Failure([]string{"foo", "bar"})
	if err == nil {
		t.Fatalf("expected an error with two args, got none")
	}

	_, err = Failure([]string{"foo", "bar", "baz"})
	if err == nil {
		t.Fatalf("expected an error with three args, got none")
	}
}

func TestFailureCmd(t *testing.T) {

	if !file.Exists("/bin/ls") {
		t.Skip("/bin/ls not present")
	}

	// No failure
	out, err := Failure([]string{"/bin/ls"})
	if err != nil {
		t.Fatalf("failed to run /bin/ls: %s", err)
	}
	if out {
		t.Fatalf("expected false, got true")
	}

	// failure
	out, err = Failure([]string{"/bin/ls /no/such/file/or/directory"})
	if err != nil {
		t.Fatalf("failed to run ls:%s", err)
	}
	if !out {
		t.Fatalf("expected true, got false")
	}
}
