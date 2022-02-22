// object.go - Contains our "object" implementation.

package ast

import (
	"fmt"
	"log"
	"strings"

	"github.com/skx/marionette/environment"
)

// Object is an interface which must be implemented by anything which will
// be used as a core-primitive value.
//
// A primitive can be output as a string, and also self-evaluate to return
// a string-value.
type Object interface {

	// String returns the object contents as a string
	String() string

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
	// Object is our parent object.
	Object

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
	// Object is our parent object.
	Object

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
	// Object is our parent object.
	Object

	// Name is the name of the function to be invoked.
	Name string

	// Arguments are the arguments to be passed to the call.
	Args []Object
}

// Evaluate returns the value of the function call.
func (f *Funcall) Evaluate(env *environment.Environment) (string, error) {

	// Lookup the function
	fn, ok := FUNCTIONS[f.Name]
	if !ok {
		return "", fmt.Errorf("function %s not defined", f.Name)
	}

	// Holder for expanded arguments
	args := []string{}

	// Convert each argument to a string
	for _, arg := range f.Args {

		// Evaluate
		val, err := arg.Evaluate(env)
		if err != nil {
			return "", err
		}

		// Save the string-representation into our temporary
		// set of arguments.
		args = append(args, val)
	}

	log.Printf("[DEBUG] Invoking function - %s(%s)", f.Name, strings.Join(args, ","))

	// Call the function, with the stringified arguments.
	ret, err := fn(env, args)
	if err != nil {
		return "", err
	}

	log.Printf("[DEBUG] Function result - %s(%s) -> %s", f.Name, strings.Join(args, ","), ret)

	// Get the output of the return value as string
	return ret.Evaluate(env)
}

// String returns our object as a string.
func (f *Funcall) String() string {
	args := ""
	for _, a := range f.Args {
		if len(args) > 0 {
			args += ","
		}
		args += a.String()
	}
	return fmt.Sprintf("Funcall{%s(%s)}", f.Name, args)
}

// Number represents an integer/hexadecimal/octal number.
//
// Note that we support integers only, not floating-point numbers.
type Number struct {
	// Object is our parent object.
	Object

	// Value is the literal number we're holding.
	Value int64
}

// String returns our object as a string.
func (n *Number) String() string {
	return fmt.Sprintf("Number{%d}", n.Value)
}

// Evaluate returns the value of the Number object.
func (n *Number) Evaluate(env *environment.Environment) (string, error) {
	return fmt.Sprintf("%d", n.Value), nil
}

// String represents a string literal
type String struct {
	// Object is our parent object.
	Object

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
