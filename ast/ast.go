// Package AST contains a simple AST for our input values.
//
// The intention is that the parser will process a list of rules, and
// will generate a Program which will be executed.  The program will
// consist of an arbitrary number of assignments, inclusions, and rules.
package ast

import (
	"github.com/skx/marionette/token"
)

// Node represents a node.
type Node interface {

	// String returns this object as a string.
	String() string
}

// Include represents a file inclusion.
//
// This is produced by the parser when an include statement is seen.
type Include struct {
	// Node is our parent node.
	Node

	// Source holds the the source to include
	Source string
}

// Rule represents a parsed rule.
type Rule struct {
	// Node is our parent node.
	Node

	// Type contains the rule-type.
	Type string

	// Name contains the name of the rule.
	Name string

	// Triggered is true if this rule is only triggered by
	// another rule notifying it.
	//
	// Triggered rules are ignored when processing our list.
	Triggered bool

	// Parameters contains the params supplied by the user.
	Params map[string]interface{}
}

// Assign represents a variable assignment.
//
// TODO This is not (yet) used.
type Assign struct {
	// Node is our parent node.
	Node

	// Key is the name of the variable
	Key string

	// Value is the value of our variable
	Value token.Token
}

// Program contains a program
type Program struct {

	// Recipe contains the list of rule/assignment/include statements
	// we're going to process.
	Recipe []Node
}
