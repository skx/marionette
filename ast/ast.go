// Package AST contains a simple AST for our scripts.
//
// The intention is that the parser will process a list of
// rules, and will generate a Program which will be executed.
//
// The program will consist of an arbitrary number of
// assignments, inclusions, and rules.
package ast

import (
	"github.com/skx/marionette/conditionals"
	"github.com/skx/marionette/token"
)

// Node represents a node.
type Node interface {
}

// Assign represents a variable assignment.
type Assign struct {
	// Node is our parent node.
	Node

	// Key is the name of the variable.
	Key string

	// Value is the value of our variable.
	//
	// This is a token so that we can execute commands,
	// via backticks.
	Value token.Token
}

// Include represents a file inclusion.
//
// This is produced by the parser by include statements.
type Include struct {
	// Node is our parent node.
	Node

	// Source holds the location to include.
	Source string

	// ConditionType holds "if" or "unless" if this inclusion is conditional
	ConditionType string

	// ConditionRule holds a conditional-rule to match, if ConditionType is non-empty.
	ConditionRule *conditionals.ConditionCall
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

// Program contains a program
type Program struct {

	// Recipe contains the list of rule/assignment/include
	// statements we're going to process.
	Recipe []Node
}
