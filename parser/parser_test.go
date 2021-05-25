//
// Test-cases for our parser.
//

package parser

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/skx/marionette/conditionals"
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
		"let m = `fkldjfdf/sdfsd/fsd/fds/fsdf/sf`",
	}

	for _, test := range broken {

		// Create a new parser
		p := New(test)

		// Count the number of variables which exist in an
		// empty parser
		vars := p.e.Variables()
		vlen := len(vars)

		// Parse the input
		_, err := p.Parse()
		if err == nil {
			t.Errorf("expected error parsing broken assign '%s' - got none", test)
		}

		// Count the variables which are set,
		// there should be no increase.
		v := p.e.Variables()
		if len(v) != vlen {
			t.Errorf("unexpected variables present")
		}
	}

	// valid tests
	valid := []string{`let a = "foo"`,
		"let a = `/bin/ls`",
	}

	for _, test := range valid {

		// Create the parser
		p := New(test)

		// Count the number of variables which exist in an
		// empty parser
		vars := p.e.Variables()
		vlen := len(vars)

		// Parse
		_, err := p.Parse()
		if err != nil {
			t.Errorf("unexpected error parsing '%s' %s", test, err.Error())
		}

		// Now we should have one more variable.
		v := p.e.Variables()
		if len(v) != (vlen + 1) {
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
                                 command => "echo Comparison Worked!",
                                 if =>        }`,
			Error: "expected identifier"},

		{Input: ` => `,
			Error: "expected identifier"},

		{Input: `shell { name => "OK",
                                 command => "echo Comparison Worked!",
                                 if => equal(
        }`,
			Error: "unexpected EOF in conditional"},
		{Input: `shell { name => "OK",
                                 command => "echo Comparison Worked!",
                                 unless`,
			Error: "expected => after conditional"},
		{Input: `shell { name => "OK",
                                 command => "echo Comparison Worked!",
                                 unless => foo foo`,
			Error: "expected ( after conditional"},
		{Input: "shell { command => \"echo OK\", if => equals( \"test\", `/missing/file/here`,     ) }",
			Error: "error running command"},
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
                          command => "echo Comparison Worked!",
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

	res, ok := out[0].Params["if"].(*conditionals.ConditionCall)
	if !ok {
		t.Errorf("we didn't parse a conditional")
	}

	formatted := res.String()
	if formatted != "equal(foo,foo)" {
		t.Errorf("failed to stringify valid comparison")
	}
}

// TestInclude handles some simple include-file things.
func TestInclude(t *testing.T) {

	// Invalid tests
	invalid := []string{
		"include ",
		"include {",
		"include (",
	}

	for _, txt := range invalid {
		p := New(txt)
		_, err := p.Parse()

		if err == nil {
			t.Errorf("expected error parsing '%s', got none", txt)
		}
		if !strings.Contains(err.Error(), "supported for include statements") {
			t.Errorf("got error (expected) but wrong value (%s)", err.Error())
		}
	}

	//
	// Attempt to include file that doesn't exist
	//
	txt := `include "/path/to/fil/not/found"`
	p := New(txt)
	_, err := p.Parse()
	if err == nil {
		t.Fatalf("including a file that wasn't found worked!")
	}

	//
	// Attempt to include file, via a broken backtick.
	//
	txt = "include `/path/to/fil/not/found`"
	p = New(txt)
	_, err = p.Parse()
	if err == nil {
		t.Fatalf("including a file that wasn't found worked!")
	}

	//
	// Now write out a temporary file
	//
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		t.Fatalf("create a temporary file failed")
	}

	// Write the input
	_, err = tmpfile.Write([]byte("# This is a comment\n"))
	if err != nil {
		t.Fatalf("error writing temporary file")
	}

	//
	// Parse the file that includes this
	//
	txt = `include "` + tmpfile.Name() + `"`
	p = New(txt)
	_, err = p.Parse()
	if err != nil {
		t.Fatalf("got error reading include file %s", err.Error())
	}

	//
	// Second attempt, update our include file to include
	// broken syntax
	//
	_, err = tmpfile.Write([]byte("let f => ff \n"))
	if err != nil {
		t.Fatalf("error writing temporary file")
	}

	//
	// Parse the file that includes this
	//
	txt = `include "` + tmpfile.Name() + `"`
	p = New(txt)
	_, err = p.Parse()
	if err == nil {
		t.Fatalf("got error reading include file %s", err.Error())
	}

	//
	// Cleanup
	//
	os.Remove(tmpfile.Name())

}

// TestMapper handles mapper-invokation
func TestMapper(t *testing.T) {

	txt := `# Comment
let foo = "bar"

shell { command => "x" }
`
	p := New(txt)
	_, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error parsing file")
	}

	//
	// Now we should have a variable "foo"
	//
	x := p.mapper("foo")
	if x != "bar" {
		t.Fatalf("${foo} had wrong value 'bar' != '%s'", x)
	}

	//
	// Now getenv
	//
	if p.mapper("USER") != os.Getenv("USER") {
		t.Fatalf("getenv failed")
	}
}
