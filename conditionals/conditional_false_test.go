// false() method tests

package conditionals

import (
	"testing"
)

// TestFalseArgs ensures we have the correct argument count.
func TestFalseArgs(t *testing.T) {

	_, err := False([]string{})
	if err == nil {
		t.Fatalf("expect an error with zero args, got none")
	}

	_, err = False([]string{"foo"})
	if err != nil {
		t.Fatalf("expected no error with one args, got none")
	}

	_, err = False([]string{"foo", "bar"})
	if err == nil {
		t.Fatalf("expected error with two args, got %s", err.Error())
	}

	_, err = False([]string{"foo", "bar", "baz"})
	if err == nil {
		t.Fatalf("expected error with three args, got none")
	}
}

// TestFalse ensures we report on length correctly.
func TestFalse(t *testing.T) {

	// Empty
	res, err := False([]string{""})
	if err != nil {
		t.Fatalf("failed to test conditional")
	}
	if !res {
		t.Fatalf("unexpected empty result")
	}

	// NonEmpty
	res, err = False([]string{"Steve"})
	if err != nil {
		t.Fatalf("failed to test conditional")
	}
	if res {
		t.Fatalf("unexpected non-empty result")
	}
}
