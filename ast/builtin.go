package ast

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/skx/marionette/environment"
)

// fnEqual returns true/false depending upon whether the two arguments
// are equal.
func fnEqual(env *environment.Environment, args []string) (Node, error) {

	// Two arguments are required.
	if len(args) != 2 {
		return nil, fmt.Errorf("'equal' requires two arguments")
	}

	// If the values differ then not-equal
	if args[0] != args[1] {
		return &Boolean{Value: false}, nil
	}

	// Same values?  Then equal
	return &Boolean{Value: true}, nil
}

// fnLen returns the length of the given node.
func fnLen(env *environment.Environment, args []string) (Node, error) {

	// Only one argument is supported.
	if len(args) != 1 {
		return nil, fmt.Errorf("'len' requires a single argument")
	}

	return &Number{Value: int64(utf8.RuneCountInString(args[0]))}, nil
}

// fnLower converts the given node to lower-case.
func fnLower(env *environment.Environment, args []string) (Node, error) {

	// Only one argument is supported.
	if len(args) != 1 {
		return nil, fmt.Errorf("'lower' requires a single argument")
	}

	return &String{Value: strings.ToLower(args[0])}, nil
}

// fnUpper converts the given node to upper-case.
func fnUpper(env *environment.Environment, args []string) (Node, error) {

	// Only one argument is supported.
	if len(args) != 1 {
		return nil, fmt.Errorf("'upper' requires a single argument")
	}

	return &String{Value: strings.ToUpper(args[0])}, nil
}

func init() {
	FUNCTIONS["equal"] = fnEqual
	FUNCTIONS["equals"] = fnEqual
	FUNCTIONS["len"] = fnLen
	FUNCTIONS["lower"] = fnLower
	FUNCTIONS["upper"] = fnUpper
}
