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

	"github.com/skx/marionette/environment"
	"github.com/skx/marionette/token"
)

// BuiltIn is the signature of a built-in function
type BuiltIn func(env *environment.Environment, args []string) (Node, error)

// FUNCTIONS contains our list of built-in functions, as a map.
//
// The key is the name of the function, and the value is the pointer to the
// function which is used to implement it.
var FUNCTIONS map[string]BuiltIn

func init() {
	FUNCTIONS = make(map[string]BuiltIn)
}

//
// Node represents a node that we can process.
type Node interface {

	// String will convert this Node object to a human-readable form.
	String() string
}

// Literal is an interface which must be implemented by any of our
// core primitive types.
//
// When executed a primitive will return a string-version of the contents.
type Literal interface {

	// Evaluate returns the value of the literal.
	//
	// The environment is made available because we want to
	// allow variable expansion within strings and backticks.
	Evaluate(env *environment.Environment) (string, error)
}

//
// Primitive values follow
//

// Backtick is a value which returns the output of executing a command.
type Backtick struct {
	// Node is our parent object.
	Node

	// Value is the command we're to execute.
	Value string
}

// String returns our object as a string.
func (b *Backtick) String() string {
	return fmt.Sprintf("Backtick{Command:%s}", b.Value)
}

// Evaluate returns the value of the Backtick object.
func (b *Backtick) Evaluate(env *environment.Environment) (string, error) {
	ret, err := env.ExpandBacktick(b.Value)
	return ret, err
}

// Boolean represents a true/false value
type Boolean struct {
	// Node is our parent object.
	Node

	// Value is the literal value we hold
	Value bool
}

// String returns our object as a string.
func (b *Boolean) String() string {
	if b.Value {
		return ("Boolean{true}")
	}
	return ("Boolean{false}")
}

// Evaluate returns the value of the Boolean object.
func (b *Boolean) Evaluate(env *environment.Environment) (string, error) {
	if b.Value {
		return "true", nil
	}
	return "false", nil
}

// Funcall represents a function-call.
type Funcall struct {
	// Node is our parent object.
	Node

	// Name is the name of the function to be invoked.
	Name string

	// Arguments are the arguments to be passed to the call.
	Args []Node
}

// Evaluate returns the value of the function call.
func (f *Funcall) Evaluate(env *environment.Environment) (string, error) {

	// Lookup the function
	fn, ok := FUNCTIONS[f.Name]
	if !ok {
		return "", fmt.Errorf("function %s not defined", f.Name)
	}

	// Convert each argument to a string
	args := []string{}

	for _, arg := range f.Args {

		// expand the argument
		obj, ok := arg.(Literal)
		if !ok {
			return "", fmt.Errorf("failed to convert %v to string", arg)
		}
		val, err := obj.Evaluate(env)
		if err != nil {
			return "", err
		}

		args = append(args, val)
	}

	// Call the function
	ret, err := fn(env, args)
	if err != nil {
		return "", err
	}

	// convert the result to an object
	obj, ok2 := ret.(Literal)
	if ok2 {
		val, err2 := obj.Evaluate(env)
		if err2 != nil {
			return "", err2
		}

		return val, nil
	}

	return "", fmt.Errorf("return value wasn't a literal object")
}

// String returns our object as a string
func (f *Funcall) String() string {
	return fmt.Sprintf("Funcall{%s}", f.Name)
}

// Number represents an integer/hexadecimal/octal number.
//
// Note that we support integers only, not floating-point numbers.
type Number struct {
	// Node is our parent object.
	Node

	// Value is the literal string we've got
	Value int64
}

// String returns our object as a string
func (n *Number) String() string {
	return fmt.Sprintf("Number{%d}", n.Value)
}

// Evaluate returns the value of the Number object.
func (n *Number) Evaluate(env *environment.Environment) (string, error) {
	return fmt.Sprintf("%d", n.Value), nil
}

// String represents a string literal
type String struct {
	// Node is our parent object.
	Node

	// Value is the literal string we've got
	Value string
}

// String returns our object as a string.
func (s *String) String() string {
	return fmt.Sprintf("String{%s}", s.Value)
}

// Evaluate returns the value of the String object.
//
// This means expanding the variables contained within the string.
func (s *String) Evaluate(env *environment.Environment) (string, error) {
	return env.ExpandVariables(s.Value), nil

}

//
// "Real" nodes follow
//

// Assign represents a variable assignment.
type Assign struct {
	// Node is our parent object.
	Node

	// Key is the name of the variable.
	Key string

	// Value is the value which will be set.
	Value Node

	// ConditionType holds "if" or "unless" if this assignment
	// action is to be carried out conditionally.
	ConditionType string

	// Function holds a function to call, if this is a conditional
	// action.
	Function Node
}

// String turns an Assign object into a decent string.
func (a *Assign) String() string {
	if a == nil {
		return "<nil>"
	}
	// No condition?
	if a.ConditionType == "" {
		return (fmt.Sprintf("Assign{Key:%s Value:%s}", a.Key, a.Value))
	}

	return (fmt.Sprintf("Assign{Key:%s Value:%s ConditionType:%s Condition:%s}", a.Key, a.Value, a.ConditionType, a.Function))
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

	// Function holds a function to call, if this is a conditional
	// action.
	Function *Funcall
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
		i.Source, i.ConditionType, i.Function))
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
	// a single ast Node, or an array of them.
	//
	Params map[string]interface{}

	// ConditionType holds "if" or "unless" if this rule should
	// be executed only conditionally.
	ConditionType string

	// Function holds a function to call, if this is a conditional
	// action.
	Function *Funcall
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

	return fmt.Sprintf("Rule %s{%s ConditionType:%s Condition:%s}", r.Type, args, r.ConditionType, r.Function)

}

// Program contains a program
type Program struct {

	// Recipe contains the list of rule/assignment/include
	// statements we're going to process.
	Recipe []Node
}
