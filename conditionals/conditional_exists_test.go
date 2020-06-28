// exists() method tests

package conditionals

import (
	"io/ioutil"
	"os"
	"testing"
)

// TestExistsArgs ensures we have the correct argument count.
func TestExistsArgs(t *testing.T) {

	_, err := Exists([]string{})
	if err == nil {
		t.Fatalf("expect an error with zero args, got none")
	}

	_, err = Exists([]string{"foo"})
	if err != nil {
		t.Fatalf("expected no error with one args, got %s", err.Error())
	}

	_, err = Exists([]string{"foo", "bar"})
	if err == nil {
		t.Fatalf("expected error with two args, got none")
	}

	_, err = Exists([]string{"foo", "bar", "baz"})
	if err == nil {
		t.Fatalf("expected error with three args, got none")
	}
}

// TestExists ensures we report on file existence.
func TestExists(t *testing.T) {

	// Create a file, and ensure it exists
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		t.Fatalf("create a temporary file failed")
	}

	// Does it exist
	res, err := Exists([]string{tmpfile.Name()})
	if err != nil {
		t.Fatalf("failed to test for file existence")
	}
	if !res {
		t.Fatalf("after creating a temporary file it doesnt exist")
	}

	// Remove the file
	os.Remove(tmpfile.Name()) // clean up

	// Does it exist, still?
	res, err = Exists([]string{tmpfile.Name()})
	if err != nil {
		t.Fatalf("failed to test for file existence")
	}
	if res {
		t.Fatalf("after removing a temporary file it still exists")
	}
}
