package environment

import (
	"os"
	"runtime"
	"testing"
)

// Test built-in values
func TestExpected(t *testing.T) {

	x := New()

	// Get the arch
	a, aOK := x.Get("ARCH")
	if !aOK {
		t.Fatalf("Failed to get ${ARCH}")
	}
	if a != runtime.GOARCH {
		t.Fatalf("${ARCH} had wrong value != %s", runtime.GOARCH)
	}

	// Get the OS
	o, oOK := x.Get("OS")
	if !oOK {
		t.Fatalf("Failed to get ${OS}")
	}
	if o != runtime.GOOS {
		t.Fatalf("${OS} had wrong value != %s", runtime.GOOS)
	}

	// Test getting environmental variables
	testVar := "steve"
	os.Setenv("TEST_ME", testVar)
	out := x.ExpandVariables("${TEST_ME}")
	if out != "steve" {
		t.Fatalf("${TEST_ME} had wrong value != %s", testVar)
	}

	// Chagne the variable in the map, which will
	// take precedence to the env
	updated := "OK, Computer"
	x.vars["TEST_ME"] = updated
	out = x.ExpandVariables("${TEST_ME}")
	if out != updated {
		t.Fatalf("${TEST_ME} had wrong value %s != %s", out, updated)
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
