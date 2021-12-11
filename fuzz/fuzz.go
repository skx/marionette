//go:build gofuzz
// +build gofuzz

package fuzz

import (
	"github.com/skx/marionette/environment"
	"github.com/skx/marionette/parser"
)

func Fuzz(data []byte) int {

	// Create a new execution environment
	env := environment.New()

	// Create a new parser, ensuring it uses the new environment.
	p := parser.NewWithEnvironment(string(data), env)

	// Parse the rules
	p.Parse()

	return 0
}
