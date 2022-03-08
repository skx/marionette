//go:build go1.18
// +build go1.18

package parser

import (
	"strings"
	"testing"
)

func FuzzParser(f *testing.F) {

	// empty string
	f.Add([]byte(""))

	// invalid entries
	f.Add([]byte("let"))
	f.Add([]byte("3="))
	f.Add([]byte("let false=\"steve\" "))

	// assignments
	f.Add([]byte("let foo=\"bar\""))
	f.Add([]byte("let foo=3"))
	f.Add([]byte("let foo=true"))
	f.Add([]byte("let foo=false"))
	f.Add([]byte("let id=`/usr/bin/id -u`"))

	// blocks
	f.Add([]byte(`shell { command => "/usr/bin/uptime", shell => true } `))
	f.Add([]byte(`shell { command => "/usr/bin/uptime", shell => "true" } `))
	f.Add([]byte(`shell { command => [ "/usr/bin/uptime", "/usr/bin/id" ] } `))

	// block with conditions
	f.Add([]byte(`shell { command => "uptime", if => equal(\"one\",\"two\"); } `))
	f.Add([]byte(`shell { command => "uptime", unless => false(\"/bin/true\"); } `))

	// Assignments
	f.Add([]byte(`let a = "foo"`))
	f.Add([]byte(`let a = true`))
	f.Add([]byte(`let a = false;`))
	f.Add([]byte(`let a = 32`))
	f.Add([]byte(`let invalid =  [ "steve", "kemp"]`))

	// Known errors are listed here.
	//
	// The purpose of fuzzing is to find panics, or unexpected errors.
	//
	// Some programs are obviously invalid though, so we don't want to
	// report those known-bad things.
	known := []string{
		"expected",
		"expected identifier name after conditional",
		"assignment can only be made to identifiers",
		"illegal token",
		"end of file",
		"you cannot assign an array to a variable",
		"unterminated assignment",
		"strconv.ParseInt: parsing",
		"unexpected bare identifier",
	}

	f.Fuzz(func(t *testing.T, input []byte) {

		// Create a new parser
		c := New(string(input))

		// Parse, looking for errors
		_, err := c.Parse()
		if err != nil {

			// We got an error.  Is it a known one?
			for _, e := range known {

				// This is a known error, we expect to get
				if strings.Contains(err.Error(), e) {
					return
				}
			}

			// New error!  Fuzzers are magic, and this is a good
			// discovery :)
			t.Errorf("Input gave bad error: %s %s\n", input, err)
		}
	})
}
