//go:build !darwin && !windows

package conditionals

import (
	"testing"

	"github.com/skx/marionette/file"
)

func TestSuccessArgs(t *testing.T) {

	_, err := Success([]string{})
	if err == nil {
		t.Fatalf("expect an error with zero args, got none")
	}

	_, err = Success([]string{"foo"})
	if err != nil {
		t.Fatalf("expected no error with one arg, got %s", err)
	}

	_, err = Success([]string{"foo", "bar"})
	if err == nil {
		t.Fatalf("expected an error with two args, got none")
	}

	_, err = Success([]string{"foo", "bar", "baz"})
	if err == nil {
		t.Fatalf("expected an error with three args, got none")
	}
}

func TestSuccessCmd(t *testing.T) {

	if !file.Exists("/bin/ls") {
		t.Skip("/bin/ls not present")
	}

	// No failure
	out, err := Success([]string{"/bin/ls"})
	if err != nil {
		t.Fatalf("failed to run ls:%s", err)
	}
	if !out {
		t.Fatalf("expected true, got false")
	}

	// failure
	out, err = Success([]string{"ls /no/such/file/or/directory"})
	if err != nil {
		t.Fatalf("failed to run ls:%s", err)
	}
	if out {
		t.Fatalf("expected false, got true")
	}
}
