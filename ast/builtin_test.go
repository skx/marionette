package ast

import (
	"testing"

	"github.com/skx/marionette/file"
)

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
	m["empty"] = 1
	m["equal"] = 2
	m["equals"] = 2
	m["exists"] = 1
	m["failure"] = 1
	m["len"] = 1
	m["lower"] = 1
	m["md5sum"] = 1
	m["md5"] = 1
	m["nonempty"] = 1
	m["on_path"] = 1
	m["set"] = 1
	m["sha1"] = 1
	m["sha1sum"] = 1
	m["success"] = 1
	m["unset"] = 1
	m["upper"] = 1

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

func TestFunctions(t *testing.T) {

	type TestCase struct {
		Name   string
		Input  []string
		Output Object
	}

	tests := []TestCase{

		TestCase{Name: "equal",
			Input: []string{
				"one",
				"two",
			},
			Output: &Boolean{Value: false},
		},
		TestCase{Name: "equal",
			Input: []string{
				"one",
				"one",
			},
			Output: &Boolean{Value: true},
		},

		TestCase{Name: "contains",
			Input: []string{
				"cake ",
				"pie",
			},
			Output: &Boolean{Value: false},
		},
		TestCase{Name: "contains",
			Input: []string{
				"cake",
				"ake",
			},
			Output: &Boolean{Value: true},
		},
		TestCase{Name: "empty",
			Input: []string{
				"",
			},
			Output: &Boolean{Value: true},
		},
		TestCase{Name: "empty",
			Input: []string{
				"one",
			},
			Output: &Boolean{Value: false},
		},
		TestCase{Name: "len",
			Input: []string{
				"one",
			},
			Output: &Number{Value: 3},
		},
		TestCase{Name: "len",
			Input: []string{
				"steve",
			},
			Output: &Number{Value: 5},
		},
		TestCase{Name: "md5",
			Input: []string{
				"password",
			},
			Output: &String{Value: "5f4dcc3b5aa765d61d8327deb882cf99"},
		},
		TestCase{Name: "sha1sum",
			Input: []string{
				"secret",
			},
			Output: &String{Value: "e5e9fa1ba31ecd1ae84f75caaa474f3a663f05f4"},
		},
		TestCase{Name: "upper",
			Input: []string{
				"one",
			},
			Output: &String{Value: "ONE"},
		},
		TestCase{Name: "lower",
			Input: []string{
				"OnE",
			},
			Output: &String{Value: "one"},
		},
		TestCase{Name: "set",
			Input: []string{
				"OnE",
			},
			Output: &Boolean{Value: true},
		},
		TestCase{Name: "set",
			Input: []string{
				"",
			},
			Output: &Boolean{Value: false},
		},
	}

	if file.Exists("/etc/passwd") {
		tests = append(tests,
			TestCase{Name: "exists",
				Input: []string{
					"/etc/passwd",
				},
				Output: &Boolean{Value: true},
			})
		tests = append(tests,
			TestCase{Name: "exists",
				Input: []string{
					"/etc/passwd.passwd/blah",
				},
				Output: &Boolean{Value: false},
			},
		)
	}
	for _, test := range tests {

		// Find the function
		fun, ok := FUNCTIONS[test.Name]
		if !ok {
			t.Fatalf("failed to find test")
		}

		// Call the function
		ret, err := fun(nil, test.Input)
		if err != nil {
			t.Fatalf("error calling %s(%v) %s", test.Name, test.Input, err)
		}

		// Compare the results
		a := test.Output.String()
		b := ret.String()

		if a != b {
			t.Fatalf("error running test %s(%v) - %s != %s", test.Name, test.Input, a, b)
		}
	}
}
