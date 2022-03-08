//
// Test-cases for our parser.
//

package parser

import (
	"os"
	"strings"
	"testing"

	"github.com/skx/marionette/ast"
)

// TestAssignment performs basic assignment-statement testing
func TestAssignment(t *testing.T) {

	// Broken statements
	broken := []string{
		"let",
		"let foo",
		"let foo=",
		"let foo=bar",
		"let a=\"b\" unless",
		"let a=\"b\" unless false",
		"let a=\"b\" unless false(",
		"let a=\"b\" unless false(/bin/ls,",
		"let a=\"3\" if     ",
		"let a=\"3\" if     true",
		"let a=\"3\" if     true(",
		"let a=\"3\" if     true(/bin/ls",
		"let a=\"3\" if     true(/bin/ls,",
		"let 3=\"3\"",
		"let false=\"3\"",
		"let true=\"3\"",
	}

	// Ensure each one fails
	for _, test := range broken {

		// Create a sub-test for this input
		t.Run(test, func(t *testing.T) {

			p := New(test)
			_, err := p.Parse()

			if err == nil {
				t.Errorf("expected error parsing broken assign '%s' - got none", test)
			}
		})
	}

	// Now test valid assignments
	valid := []string{
		"let x = `/bin/true`",
		"let x = `/bin/true` if equal(\"a\",\"a\")",
		"let a = \"boo\"",
		"let _false_ = \"ok\"",
		"let _true_like = \"ok\"",
	}

	// Ensure each one succeeds
	for _, test := range valid {

		// Create a sub-test for this input
		t.Run(test, func(t *testing.T) {
			p := New(test)
			_, err := p.Parse()

			if err != nil {
				t.Errorf("unexpected error parsing assignment '%s': %s", test, err)
			}
		})
	}
}

// TestBlock performs basic block-parsing.
func TestBlock(t *testing.T) {

	// Broken tests
	broken := []string{
		`"foo`,
		"foo",
		`foo { name : "test" }`,
		`foo { name = "steve"}`,
		`foo { name = "steve`,
		`foo { name = { "steve" } }`,
		`foo { name =`,
		`foo { name `,
		`foo { "name" `,
		`foo { "unterminated `,
		`foo { `,
		`foo { name => "unterminated `,
		`foo { name => `,
		`foo { name => { "steve, "kemp"] }`,
		`foo { name => [ "steve`,
		`foo { name => [ ,,, "",,,`,
	}

	for _, test := range broken {

		// Create a sub-test for this input
		t.Run(test, func(t *testing.T) {
			p := New(test)
			_, err := p.Parse()

			if err == nil {
				t.Errorf("expected error parsing broken block '%s' - got none", test)
			}
		})
	}

	// valid tests
	valid := []string{`file { target => "steve", name => "steve" }`,
		`moi triggered { test => "steve", name => "steve" }`,
		`foo { name => [ "one", "two", ] }`,
	}

	for _, test := range valid {

		// Create a sub-test for this input
		t.Run(test, func(t *testing.T) {

			p := New(test)
			rules, err := p.Parse()

			if err != nil {
				t.Errorf("unexpected error parsing '%s' %s", test, err.Error())
			}

			if len(rules.Recipe) != 1 {
				t.Errorf("expected a single rule")
			}
		})
	}
}

// TestConditionalErrors performs some sanity-checks that broken conditionals
// result in expected errors.
func TestConditionalErrors(t *testing.T) {

	type TestCase struct {
		Input string
		Error string
	}

	// Broken tests
	broken := []TestCase{
		{Input: `shell { name => "OK1",
                                 command => "echo Comparison Worked!",
                                 if =>        }`,
			Error: "unexpected type parsing primitive"},

		{Input: `shell { name => "OK2",
                                 command => "echo Comparison Worked!",
                                 if => equal(
        }`,
			Error: "unexpected type parsing primitive:token"},
		{Input: `shell { name => "OK3",
                                 command => "echo Comparison Worked!",
                                 unless`,
			Error: "expected => after conditional unless"},
		{Input: `shell { name => "OK4",
                                 command => "echo Comparison Worked!",
                                 unless => foo foo`,
			Error: "unexpected bare identifier foo"},
	}

	for _, test := range broken {

		// Create a sub-test for this input
		t.Run(test.Input, func(t *testing.T) {

			p := New(test.Input)
			_, err := p.Parse()

			if err == nil {
				t.Errorf("expected error parsing broken input '%s' - got none", test.Input)
			} else {
				if !strings.Contains(err.Error(), test.Error) {
					t.Errorf("error '%s' did not match '%s' when hadnling %s", err.Error(), test.Error, test.Input)
				}
			}
		})
	}
}

// TestConditional performs a basic sanity-check that a conditional
// looks sane.
func TestConditional(t *testing.T) {

	input := `shell { name => "OK",
                          command => "echo Comparison Worked!",
                          if => equal( "foo", "foo" ),
                  }`

	p := New(input)
	out, err := p.Parse()

	if err != nil {
		t.Errorf("unexpected error parsing valid input '%s': %s", input, err.Error())
	}

	// We should have one result
	if len(out.Recipe) != 1 {
		t.Errorf("unexpected number of results")
	}

	rule := out.Recipe[0].(*ast.Rule)

	// Did we get the right type of condition?
	if rule.ConditionType != "if" {
		t.Errorf("we didn't parse a conditional")
	}

	// Does it look like the right test?
	formatted := rule.Function.String()
	if formatted != "Funcall{equal(String{foo},String{foo})}" {
		t.Errorf("failed to stringify valid comparison: %s", formatted)
	}
}

// TestInclude performs basic testing of our include-file handling
func TestInclude(t *testing.T) {

	// Broken statements
	broken := []string{
		"include",
		"include 22.2",
		"include \"test.inc\" unless false(/bin/ls",
		"include \"test.inc\" unless false(/bin/ls,",
		"include \"test.inc\" if true(/bin/ls,",
		"include \"test.inc\" if true(/bin/ls",
	}

	// Ensure each one fails
	for _, test := range broken {

		// Create a sub-test for this input
		t.Run(test, func(t *testing.T) {
			p := New(test)
			_, err := p.Parse()

			if err == nil {
				t.Errorf("expected error parsing broken include '%s' - got none", test)
			}
		})
	}

	// Now test valid includes
	valid := []string{
		"include \"test.inc\"",
		"include [ \"test.inc\", \"test.inc\"] ",
		"include \"test.inc\" unless failure(\"/bin/ls\")",
		"include \"test.inc\" if success(\"/bin/ls\")",
	}

	// Ensure each one succeeds
	for _, test := range valid {

		// Create a sub-test for this input
		t.Run(test, func(t *testing.T) {
			p := New(test)
			_, err := p.Parse()

			if err != nil {
				t.Errorf("unexpected error parsing include '%s': %s", test, err)
			}
		})
	}
}

// #86 - Test we can parse modules without spaces
func TestModuleSpace(t *testing.T) {

	input := `shell{command=>"id"}`

	p := New(input)
	out, err := p.Parse()

	// This should be error-free
	if err != nil {
		t.Errorf("unexpected error parsing input '%s': %s", input, err.Error())
	}

	// We should have one result
	if len(out.Recipe) != 1 {
		t.Errorf("unexpected number of results")
	}
}

// Test that we can output debug-strings
func TestDebug(t *testing.T) {

	// One example of each rule-type
	input := `
include "foo.txt"
let a = 3
shell{command=>"id"}
`
	os.Setenv("DEBUG_PARSER", "true")
	p := New(input)
	out, err := p.Parse()

	// This should be error-free
	if err != nil {
		t.Errorf("unexpected error parsing input '%s': %s", input, err.Error())
	}

	// We should have three results
	if len(out.Recipe) != 3 {
		t.Errorf("unexpected number of results")
	}
}
