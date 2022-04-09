package ast

import (
	"bufio"
	"fmt"
	"strings"
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
	m["field"] = 2
	m["gt"] = 2
	m["gte"] = 2
	m["len"] = 1
	m["lower"] = 1
	m["lt"] = 2
	m["lte"] = 2
	m["matches"] = 2
	m["md5"] = 1
	m["md5sum"] = 1
	m["nonempty"] = 1
	m["on_path"] = 1
	m["prompt"] = 1
	m["rand"] = 2
	m["set"] = 1
	m["sha1"] = 1
	m["sha1sum"] = 1
	m["success"] = 1
	m["unset"] = 1
	m["upper"] = 1
	one := []string{"1"}
	two := []string{"23", "34"}

	// Replace STDIN
	old := STDIN

	// Ensure that we can call functions with the right number
	// of arguments.
	for name, fun := range FUNCTIONS {

		// Replace STDIN
		STDIN = bufio.NewReader(strings.NewReader("STEVE\n"))

		var err error

		valid := m[name]

		if valid == 1 {
			t.Run(name, func(t *testing.T) {
				_, err = fun(nil, one)
			})
			if err != nil {
				t.Fatalf("unexpected error with 1 arg:%s", err)
			}
		} else if valid == 2 {
			t.Run(name, func(t *testing.T) {
				_, err = fun(nil, two)
			})
			if err != nil {
				t.Fatalf("unexpected error with 2 args")
			}
		} else {
			t.Fatalf("unhandled test-case for function '%s'", name)
		}

	}

	STDIN = old
}

func TestFunctions(t *testing.T) {

	// Replace STDIN
	old := STDIN

	type TestCase struct {
		// Name of function
		Name string

		// Arguments to pass to it
		Input []string

		// Expected output
		Output Object

		// Faked stdin?
		StdIn string

		// If non-empty we expect an error, and it should match
		// this text.
		Error string
	}

	tests := []TestCase{

		TestCase{Name: "lt",
			Input: []string{
				"1",
				"2",
			},
			Output: &Boolean{Value: true},
		},
		TestCase{Name: "lt",
			Input: []string{
				"1",
				"kemp",
			},
			Error: "strconv.ParseInt: parsing",
		},
		TestCase{Name: "lt",
			Input: []string{
				"steve",
				"2",
			},
			Error: "strconv.ParseInt: parsing",
		},
		TestCase{Name: "lt",
			Input: []string{
				"2",
				"2",
			},
			Output: &Boolean{Value: false},
		},
		TestCase{Name: "lt",
			Input: []string{
				"3",
				"2",
			},
			Output: &Boolean{Value: false},
		},
		TestCase{Name: "lte",
			Input: []string{
				"1",
				"kemp",
			},
			Error: "strconv.ParseInt: parsing",
		},
		TestCase{Name: "lte",
			Input: []string{
				"steve",
				"2",
			},
			Error: "strconv.ParseInt: parsing",
		},
		TestCase{Name: "lte",
			Input: []string{
				"1",
				"2",
			},
			Output: &Boolean{Value: true},
		},
		TestCase{Name: "lte",
			Input: []string{
				"2",
				"2",
			},
			Output: &Boolean{Value: true},
		},
		TestCase{Name: "lte",
			Input: []string{
				"3",
				"2",
			},
			Output: &Boolean{Value: false},
		},
		TestCase{Name: "gt",
			Input: []string{
				"1",
				"2",
			},
			Output: &Boolean{Value: false},
		},
		TestCase{Name: "gt",
			Input: []string{
				"1",
				"steve",
			},
			Error: "strconv.ParseInt: parsing",
		},
		TestCase{Name: "gt",
			Input: []string{
				"steve",
				"2",
			},
			Error: "strconv.ParseInt: parsing",
		},
		TestCase{Name: "gt",
			Input: []string{
				"2",
				"2",
			},
			Output: &Boolean{Value: false},
		},
		TestCase{Name: "gt",
			Input: []string{
				"3",
				"2",
			},
			Output: &Boolean{Value: true},
		},
		TestCase{Name: "gte",
			Input: []string{
				"1",
				"steve",
			},
			Error: "strconv.ParseInt: parsing",
		},
		TestCase{Name: "gte",
			Input: []string{
				"steve",
				"2",
			},
			Error: "strconv.ParseInt: parsing",
		},
		TestCase{Name: "gte",
			Input: []string{
				"1",
				"2",
			},
			Output: &Boolean{Value: false},
		},
		TestCase{Name: "gte",
			Input: []string{
				"2",
				"2",
			},
			Output: &Boolean{Value: true},
		},
		TestCase{Name: "gt",
			Input: []string{
				"3",
				"2",
			},
			Output: &Boolean{Value: true},
		},
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
		TestCase{Name: "field",
			Input: []string{
				"Steve Kemp",
				"0",
			},
			Output: &String{Value: "Steve"},
		},
		TestCase{Name: "field",
			Input: []string{
				"Steve Kemp",
				"1",
			},
			Output: &String{Value: "Kemp"},
		},
		TestCase{Name: "field",
			Input: []string{
				"Forename Surname",
				"10",
			},
			Output: &String{Value: ""},
		},
		TestCase{Name: "field",
			Input: []string{
				"Forename Surname",
				"Nope",
			},
			Error: "strconv",
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
		TestCase{Name: "matches",
			Input: []string{
				"password",
				"^pa[Ss]+word",
			},
			Output: &Boolean{Value: true},
		},
		TestCase{Name: "matches",
			Input: []string{
				"password",
				"^secret$",
			},
			Output: &Boolean{Value: false},
		},
		TestCase{Name: "matches",
			Input: []string{
				"password",
				"+",
			},
			Error: "error parsing regexp",
		},
		TestCase{Name: "md5",
			Input: []string{
				"password",
			},
			Output: &String{Value: "5f4dcc3b5aa765d61d8327deb882cf99"},
		},
		TestCase{Name: "prompt",
			Input: []string{
				"What is your name?",
			},
			StdIn:  "  STEVE  \n",
			Output: &String{Value: "STEVE"},
		},
		TestCase{Name: "prompt",
			Input: []string{
				"Foo",
				"Bar",
			},
			Error: "wrong number of args",
		},
		TestCase{Name: "prompt",
			Input: []string{
				"Empty",
			},
			StdIn: "",
			Error: "EOF",
		},
		TestCase{Name: "rand",
			Input: []string{
				"1",
				"100",
				"hostname",
			},
			Output: &String{Value: "96"},
		},
		TestCase{Name: "rand",
			Input: []string{
				"1",
				"steve",
			},
			Error: "strconv.Atoi",
		},
		TestCase{Name: "rand",
			Input: []string{
				"steve",
				"2",
			},
			Error: "strconv.Atoi",
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

		// Replace the contents of STDIN if we should
		if test.StdIn != "" {

			STDIN = bufio.NewReader(strings.NewReader(test.StdIn))
		}

		t.Run(fmt.Sprintf("%s(%s) -> %s", test.Name, test.Input, test.Output), func(t *testing.T) {

			// Find the function
			fun, ok := FUNCTIONS[test.Name]
			if !ok {
				t.Fatalf("failed to find test")
			}

			// Call the function
			ret, err := fun(nil, test.Input)

			// Got an error making the call
			if err != nil {

				// Should we have done?
				if test.Error != "" {

					if !strings.Contains(err.Error(), test.Error) {
						t.Fatalf("expected error (%s), but got different one (%s)", test.Error, err.Error())
					}
				} else {
					t.Fatalf("unexpected error calling %s(%v) %s", test.Name, test.Input, err)
				}
			} else {
				// Compare the results
				a := test.Output.String()
				b := ret.String()

				if a != b {
					t.Fatalf("error running test %s(%v) - %s != %s", test.Name, test.Input, a, b)
				}
			}
		})
	}

	STDIN = old
}
