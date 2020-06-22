//
// Test-cases for our parser.
//
// The parser is designed to consume tokens from our lexer so we have to
// fake-feed them in.  We do this via the `FakeLexer` helper.
//

package parser

import (
	"strings"
	"testing"
)

// TestAssign tests we can assign variables
func TestAssign(t *testing.T) {

	// Broken tests
	broken := []string{"moi",
		"let",
		"let f",
		"let f =",
		"let f => ff ",
		"let m = )",
	}

	for _, test := range broken {

		p := New(test)
		_, err := p.Parse()

		if err == nil {
			t.Errorf("expected error parsing broken assign '%s' - got none", test)
		}

		v := p.Variables()
		if len(v) != 0 {
			t.Errorf("unexpected variables present")
		}
	}

	// valid tests
	valid := []string{`let a = "foo"`,
		"let a = `/bin/ls`",
	}

	for _, test := range valid {

		p := New(test)
		_, err := p.Parse()

		if err != nil {
			t.Errorf("unexpected error parsing '%s' %s", test, err.Error())
		}

		v := p.Variables()
		if len(v) != 1 {
			t.Errorf("variable not present")
		}
	}
}

// TestBlock performs basic block-parsing.
func TestBlock(t *testing.T) {

	// Broken tests
	broken := []string{`"foo`,
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

		p := New(test)
		_, err := p.Parse()

		if err == nil {
			t.Errorf("expected error parsing broken assign '%s' - got none", test)
		}
	}

	// valid tests
	valid := []string{`file { target => "steve", name => "steve" }`,
		`moi triggered { test => "steve", name => "steve" }`,
		`foo { name => [ "one", "two", ] }`,
	}

	for _, test := range valid {

		p := New(test)
		rules, err := p.Parse()

		if err != nil {
			t.Errorf("unexpected error parsing '%s' %s", test, err.Error())
		}

		if len(rules) != 1 {
			t.Errorf("expected a single rule")
		}
	}
}

// TestConditionalErrors performs some sanity-checks that broken conditionals
// result in expected errors.
func TestConditinalErrors(t *testing.T) {

	type TestCase struct {
		Input string
		Error string
	}

	// Broken tests
	broken := []TestCase{
		{Input: `shell { name => "OK",
                                 command => "echo Comparision Worked!",
                                 if =>        }`,
			Error: "expected identifier"},

		{Input: `shell { name => "OK",
                                 command => "echo Comparision Worked!",
                                 if => equal(
        }`,
			Error: "unexpected EOF in conditional"},
		{Input: `shell { name => "OK",
                                 command => "echo Comparision Worked!",
                                 unless`,
			Error: "expected => after conditional"},
		{Input: `shell { name => "OK",
                                 command => "echo Comparision Worked!",
                                 unless => foo foo`,
			Error: "expected ( after conditional"},
	}

	for _, test := range broken {

		p := New(test.Input)
		_, err := p.Parse()

		if err == nil {
			t.Errorf("expected error parsing broken input '%s' - got none", test.Input)
		} else {
			if !strings.Contains(err.Error(), test.Error) {
				t.Errorf("error '%s' did not match '%s'", err.Error(), test.Error)
			}
		}
	}
}

// TestConditional performs a basic sanity-check that a conditional
// looks sane.
func TestConditional(t *testing.T) {

	input := `shell { name => "OK",
                          command => "echo Comparision Worked!",
                          if => equal( "foo", "foo" ),
                  }`

	p := New(input)
	out, err := p.Parse()

	if err != nil {
		t.Errorf("unexpected error parsing valid input '%s': %s", input, err.Error())
	}

	// We should have one result
	if len(out) != 1 {
		t.Errorf("unexpected number of results")
	}

	res, ok := out[0].Params["if"].(*Condition)
	if !ok {
		t.Errorf("we didn't parse a conditional")
	}

	formatted := res.String()
	if formatted != "equal(foo,foo)" {
		t.Errorf("failed to stringify valid comparison")
	}
}
