// on_path() method tests

package conditionals

import (
	"testing"
)

// TestOnPathArgs ensures we have the correct argument count.
func TestOnPathArgs(t *testing.T) {

	found := false

	_, err := OnPath([]string{})
	if err == nil {
		t.Fatalf("expect an error with zero args, got none")
	}

	_, err = OnPath([]string{"foo", "two"})
	if err == nil {
		t.Fatalf("expected error with two args, got none")
	}

	_, err = OnPath([]string{"echo"})
	if err != nil {
		t.Fatalf("expected no error with one arg, got %s", err.Error())
	}

	found, err = OnPath([]string{"cannot`-exist'"})
	if err != nil {
		t.Fatalf("expected no error with one arg, got %s", err.Error())
	}
	if found {
		t.Fatalf("unexpectedly found bogus file!")
	}
}
