// true() method tests

package conditionals

import (
	"testing"
)

// TestTrueArgs ensures we have the correct argument count.
func TestTrueArgs(t *testing.T) {

	_, err := True([]string{})
	if err == nil {
		t.Fatalf("expect an error with zero args, got none")
	}

	_, err = True([]string{"foo"})
	if err != nil {
		t.Fatalf("expected no error with one args, got none")
	}

	_, err = True([]string{"foo", "bar"})
	if err == nil {
		t.Fatalf("expected error with two args, got %s", err.Error())
	}

	_, err = True([]string{"foo", "bar", "baz"})
	if err == nil {
		t.Fatalf("expected error with three args, got none")
	}
}

// TestTrue ensures we report on length correctly.
func TestTrue(t *testing.T) {

	// Empty
	res, err := True([]string{""})
	if err != nil {
		t.Fatalf("failed to test conditional")
	}
	if res {
		t.Fatalf("unexpected empty result")
	}

	// NonEmpty
	res, err = True([]string{"Steve"})
	if err != nil {
		t.Fatalf("failed to test conditional")
	}
	if !res {
		t.Fatalf("unexpected non-empty result")
	}
}
