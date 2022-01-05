//go:build go1.18
//+build go1.18

package parser

import (
	"strings"
	"testing"
)

func FuzzParser(f *testing.F) {
	f.Add([]byte(""))
	f.Add([]byte("let foo=\"bar\""))

	// Known errors
	known := []string{
		"expected",
		"illegal token",
		"end of file",
		"unterminated assignment",
	}

	f.Fuzz(func(t *testing.T, input []byte) {

		// Create a new parser
		c := New(string(input))

		// Parse, looking for errors
		_, err := c.Parse()
		if err != nil {

			for _, e := range known {

				// This is a known error, we expect to get
				if strings.Contains(err.Error(), e) {
					return
				}
			}
			t.Errorf("Input gave bad error: %s %s\n", input, err)
		}

	})
}
