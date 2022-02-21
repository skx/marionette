package environment

import (
	"runtime"
	"testing"
)

// Test built-in values
func TestExpected(t *testing.T) {

	x := New()

	a, aOK := x.Get("ARCH")
	if !aOK {
		t.Fatalf("Failed to get ${ARCH}")
	}
	if a != runtime.GOARCH {
		t.Fatalf("${ARCH} had wrong value != %s", runtime.GOARCH)
	}

	o, oOK := x.Get("OS")
	if !oOK {
		t.Fatalf("Failed to get ${OS}")
	}
	if o != runtime.GOOS {
		t.Fatalf("${OS} had wrong value != %s", runtime.GOOS)
	}
}

// TestSet ensures a value will remain
func TestSet(t *testing.T) {

	e := New()

	// Count the variables
	vars := e.Variables()
	vlen := len(vars)

	// Confirm getting a missing value fails
	_, ok := e.Get("STEVE")
	if ok {
		t.Fatalf("Got value for STEVE, shouldn't have done")
	}

	// Set the value
	e.Set("STEVE", "KEMP")

	// Get it again
	val := ""
	val, ok = e.Get("STEVE")
	if !ok {
		t.Fatalf("After setting the value wasn't available")
	}
	if val != "KEMP" {
		t.Fatalf("Wrong value retrieved")
	}

	if len(e.Variables()) != vlen+1 {
		t.Errorf("After setting variable length didn't increase")
	}
	// Update the value
	e.Set("STEVE", "STEVE")

	// Get it again
	val, ok = e.Get("STEVE")
	if !ok {
		t.Fatalf("After setting the value wasn't available")
	}
	if val != "STEVE" {
		t.Fatalf("Wrong value retrieved, after update")
	}

}
