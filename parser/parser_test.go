//
// Test-cases for our parser.
//
// The parser is designed to consume tokens from our lexer so we have to
// fake-feed them in.  We do this via the `FakeLexer` helper.
//

package parser

import (
	"testing"
)

// TestAssign tests we can assign variables
func TestAssign(t *testing.T) {

	// Broken tests
	broken := []string{"moi",
		"let",
		"let f",
		"let f =",
		"let f => ff "}

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
		`let a = foo`,
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
