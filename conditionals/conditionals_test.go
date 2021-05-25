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

// TestString confirms we can convert a call into a string
func TestString(t *testing.T) {

	tmp := &ConditionCall{Name: "equal",
		Args: []string{"one", "two"}}

	out := tmp.String()

	if out != "equal(one,two)" {
		t.Fatalf("wrong string result")
	}

}
