package conditionals

import "testing"

// TestLookup just ensures we can find expected methods
func TestLookup(t *testing.T) {

	expected := []string{"exists", "equals", "equal"}
	bogus := []string{"bar", "baz"}

	for _, name := range expected {

		fn := Lookup(name)
		if fn == nil {
			t.Errorf("expected to find method '%s' - didn't get it", name)
		}
	}

	for _, name := range bogus {

		fn := Lookup(name)
		if fn != nil {
			t.Errorf("didn't expect to find bogus method '%s', but did", name)
		}
	}
}
