// Package ast contains a simple AST for our scripts.
//
// The intention is that the parser will process a list of
// rules, and will generate a Program which will be executed.
//
// The program will consist of an arbitrary number of
// assignments, inclusions, and rules.
package ast

import (
	"fmt"

	"github.com/skx/marionette/conditionals"
	"github.com/skx/marionette/token"
)

// Node represents a node.
type Node interface {
	String() string
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

// String turns an Assign object into a decent string.
func (a *Assign) String() string {
	if a == nil {
		return "<nil>"
	}
	t := "string"
	switch a.Value.Type {
	case token.BACKTICK:
		t = "backtick"
	case token.STRING:
		t = "string"
	default:
		t = "unknown"
	}
	return (fmt.Sprintf("Assign{Key:%s Value:%s Type:%s}", a.Key, a.Value.Literal, t))
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

// String turns an Include object into a useful string.
func (i *Include) String() string {
	if i == nil {
		return "<nil>"
	}
	if i.ConditionType == "" {
		return (fmt.Sprintf("Include{ Source:%s }", i.Source))
	}
	return (fmt.Sprintf("Include{ Source:%s  ConditionType:%s Condition:%s}",
		i.Source, i.ConditionType, i.ConditionRule))
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

// String turns a Rule object into a useful string
func (r *Rule) String() string {
	if r == nil {
		return "<nil>"
	}

	return fmt.Sprintf("Rule %s{ name:%s}", r.Type, r.Name)
}

// Program contains a program
type Program struct {

	// Recipe contains the list of rule/assignment/include
	// statements we're going to process.
	Recipe []Node
}
