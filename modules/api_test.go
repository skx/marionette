package modules

import "testing"

func TestArrayParam(t *testing.T) {

	// Setup arguments
	args := make(map[string]interface{})

	// Known-Array
	input := []string{
		"Homer",
		"Marge",
		"Bart",
		"Lisa",
		"Maggie",
	}

	// String + Array values
	args["foo"] = "bar"
	args["family"] = input

	// Confirm string was OK
	if StringParam(args, "foo") != "bar" {
		t.Fatalf("failed to get string value")
	}

	// Get the array
	array := ArrayParam(args, "family")

	// confirm length matches expectation
	if len(array) != len(input) {
		t.Fatalf("Unexpected length")
	}

	// And values
	for i, v := range input {
		if array[i] != v {
			t.Fatalf("array mismatch for value %d", i)
		}
	}

	// Treat the string as an array
	array = ArrayParam(args, "foo")
	if len(array) != 0 {
		t.Fatalf("Got result for bogus key")
	}

	// Unknown key
	array = ArrayParam(args, "testing")
	if len(array) != 0 {
		t.Fatalf("Got result for missing key")
	}

}
