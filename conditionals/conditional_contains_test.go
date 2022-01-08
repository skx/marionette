// contains() method tests

package conditionals

import (
	"testing"
)

// TestContainsArgs ensures we have the correct argument count.
func TestContainsArgs(t *testing.T) {

	_, err := Contains([]string{})
	if err == nil {
		t.Fatalf("expect an error with zero args, got none")
	}

	_, err = Contains([]string{"foo"})
	if err == nil {
		t.Fatalf("expected error with one args, got none")
	}

	_, err = Contains([]string{"foo", "bar"})
	if err != nil {
		t.Fatalf("expected no error with two args, got %s", err.Error())
	}

	_, err = Contains([]string{"foo", "bar", "baz"})
	if err == nil {
		t.Fatalf("expected error with three args, got none")
	}
}

// TestContains ensures we report on equality properly.
func TestContains(t *testing.T) {

	// Equal
	res, err := Contains([]string{"foo", "foo"})
	if err != nil {
		t.Fatalf("failed to test for equality")
	}
	if !res {
		t.Fatalf("unexpected equality result")
	}

	// Unequal
	res, err = Contains([]string{"foo", "foot"})
	if err != nil {
		t.Fatalf("failed to test for substring inclusion")
	}
	if res {
		t.Fatalf("unexpected substring inclusion result")
	}
}
