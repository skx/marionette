// equals() method tests

package conditionals

import (
	"testing"
)

// TestEqualsArgs ensures we have the correct argument count.
func TestEqualsArgs(t *testing.T) {

	_, err := Equals([]string{})
	if err == nil {
		t.Fatalf("expect an error with zero args, got none")
	}

	_, err = Equals([]string{"foo"})
	if err == nil {
		t.Fatalf("expected error with one args, got none")
	}

	_, err = Equals([]string{"foo", "bar"})
	if err != nil {
		t.Fatalf("expected no error with two args, got %s", err.Error())
	}

	_, err = Equals([]string{"foo", "bar", "baz"})
	if err == nil {
		t.Fatalf("expected error with three args, got none")
	}
}

// TestEquals ensures we report on equality properly.
func TestEquals(t *testing.T) {

	// Equal
	res, err := Equals([]string{"foo", "foo"})
	if err != nil {
		t.Fatalf("failed to test for equality")
	}
	if !res {
		t.Fatalf("unexpected equality result")
	}

	// Unequal
	res, err = Equals([]string{"foo", "foot"})
	if err != nil {
		t.Fatalf("failed to test for equality")
	}
	if res {
		t.Fatalf("unexpected equality result")
	}
}
