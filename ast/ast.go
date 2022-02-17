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
	"strings"

	"github.com/skx/marionette/conditionals"
	"github.com/skx/marionette/token"
)

// Node represents a node that we can process.
type Node interface {

	// String will convert this Node object to a human-readable form.
	String() string
}

// Assign represents a variable assignment.
type Assign struct {
	// Node is our parent object.
	Node

	// Key is the name of the variable.
	Key string

	// Value is the value of our variable.
	//
	// This is a token so that we can execute commands,
	// via backticks.
	Value token.Token

	// ConditionType holds "if" or "unless" if this assignment
	// action is to be carried out conditionally.
	ConditionType string

	// ConditionRule holds a conditional-rule to match,
	// if the ConditionType is non-empty.
	ConditionRule *conditionals.ConditionCall
}

// String turns an Assign object into a decent string.
func (a *Assign) String() string {
	if a == nil {
		return "<nil>"
	}
	t := "unknown"

	switch a.Value.Type {
	case token.BACKTICK:
		t = "backtick"
	case token.BOOLEAN:
		t = "boolean"
	case token.NUMBER:
		t = "number"
	case token.STRING:
		t = "string"
	}

	// No condition?
	if a.ConditionType == "" {
		return (fmt.Sprintf("Assign{Key:%s Value:%s Type:%s}", a.Key, a.Value.Literal, t))
	}

	return (fmt.Sprintf("Assign{Key:%s Value:%s Type:%s ConditionType:%s Condition:%s}", a.Key, a.Value.Literal, t, a.ConditionType, a.ConditionRule))
}

// Include represents a file inclusion.
//
// This is produced by the parser by include statements.
type Include struct {
	// Node is our parent object.
	Node

	// Source holds the location to include.
	Source string

	// ConditionType holds "if" or "unless" if this inclusion is to
	// be executed conditionally.
	ConditionType string

	// ConditionRule holds a conditional-rule to match, if
	// ConditionType is non-empty.
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
	//
	// The keys will be strings, with the values being either
	// a single token, or an array of tokens.
	//
	// (We need to store the tokens here, because we need to
	// be able to differentiate later whether we received a
	// string or a backtick-string which should be expanded
	// at runtime.)
	Params map[string]interface{}

	// ConditionType holds "if" or "unless" if this rule should
	// be executed only conditionally.
	ConditionType string

	// ConditionRule holds a conditional-rule to match, if
	// ConditionType is non-empty.
	ConditionRule *conditionals.ConditionCall
}

// String turns a Rule object into a useful string
func (r *Rule) String() string {
	if r == nil {
		return "<nil>"
	}

	args := ""
	for k, v := range r.Params {

		// try to format the value
		val := ""

		str, ok := v.(token.Token)
		if ok {
			val = fmt.Sprintf("\"%s\"", str)
		}

		array, ok2 := v.([]token.Token)
		if ok2 {
			for _, s := range array {
				val += fmt.Sprintf(", \"%s\"", s)
			}
			val = strings.TrimPrefix(val, ", ")
			val = "[" + val + "]"
		}

		// now add on the value(s)
		args += fmt.Sprintf(", %s->%s", k, val)
	}

	// trip prefix
	args = strings.TrimPrefix(args, ", ")

	if r.ConditionType == "" {
		return fmt.Sprintf("Rule %s{%s}", r.Type, args)
	}

	return fmt.Sprintf("Rule %s{%s ConditionType:%s Condition:%s}", r.Type, args, r.ConditionType, r.ConditionRule)

}

// Program contains a program
type Program struct {

	// Recipe contains the list of rule/assignment/include
	// statements we're going to process.
	Recipe []Node
}
