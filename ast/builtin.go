package ast

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/skx/marionette/environment"
)

// fnEqual returns true/false depending upon whether the two arguments
// are equal.
func fnEqual(env *environment.Environment, args []Node) (Node, error) {

	// Two arguments are required.
	if len(args) != 2 {
		return nil, fmt.Errorf("'equal' requires two arguments")
	}

	// Get the argument types
	a := fmt.Sprintf("%T", args[0])
	b := fmt.Sprintf("%T", args[1])

	// If the types differ then not-equal
	if a != b {
		return &Boolean{Value: false}, nil
	}

	// Now look at the contents
	aa := args[0].String()
	bb := args[1].String()

	// If the values differ then not-equal
	if aa != bb {
		return &Boolean{Value: false}, nil
	}

	// Same types and same values?  Then equal
	return &Boolean{Value: true}, nil
}

// fnLen returns the length of the given node.
func fnLen(env *environment.Environment, args []Node) (Node, error) {

	// Only one argument is supported.
	if len(args) != 1 {
		return nil, fmt.Errorf("'len' requires a single argument")
	}

	// expand the argument
	obj, ok := args[0].(Literal)
	if ok {
		val, err := obj.Evaluate(env)
		if err != nil {
			return nil, err
		}

		return &Number{Value: int64(utf8.RuneCountInString(val))}, nil
	}
	return nil, fmt.Errorf("Failed to convert %v to string", args[0])
}

// fnLower converts the given node to lower-case.
func fnLower(env *environment.Environment, args []Node) (Node, error) {

	// Only one argument is supported.
	if len(args) != 1 {
		return nil, fmt.Errorf("'lower' requires a single argument")
	}

	// expand the argument
	obj, ok := args[0].(Literal)
	if ok {
		val, err := obj.Evaluate(env)
		if err != nil {
			return nil, err
		}

		return &String{Value: strings.ToLower(val)}, nil
	}
	return nil, fmt.Errorf("Failed to convert %v to string", args[0])
}

// fnUpper converts the given node to upper-case.
func fnUpper(env *environment.Environment, args []Node) (Node, error) {

	// Only one argument is supported.
	if len(args) != 1 {
		return nil, fmt.Errorf("'upper' requires a single argument")
	}

	// expand the argument
	obj, ok := args[0].(Literal)
	if ok {
		val, err := obj.Evaluate(env)
		if err != nil {
			return nil, err
		}

		return &String{Value: strings.ToUpper(val)}, nil
	}
	return nil, fmt.Errorf("Failed to convert %v to string", args[0])
}

func init() {
	FUNCTIONS["equal"] = fnEqual
	FUNCTIONS["len"] = fnLen
	FUNCTIONS["lower"] = fnLower
	FUNCTIONS["upper"] = fnUpper
}
