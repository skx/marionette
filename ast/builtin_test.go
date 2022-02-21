package ast

import "testing"

func TestFunctionArgs(t *testing.T) {

	// Ensure that all functions error without an argument
	for name, fun := range FUNCTIONS {

		_, err := fun(nil, []string{})

		if err == nil {
			t.Fatalf("expected error invoking %s with no arguments", name)
		}
	}

	// Ensure all functions abort with too many arguments
	for name, fun := range FUNCTIONS {
		_, err := fun(nil, []string{"one", "two", "three", "four"})

		if err == nil {
			t.Fatalf("expected error invoking %s with four arguments", name)
		}
	}

	// one arg functions
	m := make(map[string]int)
	m["contains"] = 2
	m["equal"] = 2
	m["empty"] = 1
	m["nonempty"] = 1
	m["set"] = 1
	m["unset"] = 1
	m["exists"] = 1
	m["equals"] = 2
	m["len"] = 1
	m["lower"] = 1
	m["upper"] = 1
	m["success"] = 1
	m["failure"] = 1
	m["on_path"] = 1

	one := []string{"one"}
	two := []string{"one", "two"}

	// Ensure that we can call functions with the right number
	// of arguments.
	for name, fun := range FUNCTIONS {

		var err error

		valid := m[name]

		if valid == 1 {
			_, err = fun(nil, one)
		} else if valid == 2 {
			_, err = fun(nil, two)
		} else {
			t.Fatalf("unhandled test-case for function '%s'", name)
		}

		if err != nil {
			t.Fatalf("unexpected error invoking %s with %d args", name, valid)
		}
	}

}
