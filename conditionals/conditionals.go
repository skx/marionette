// Package conditionals contains the implementation of our conditional
// functions.
//
// A conditional function is something that can be referred to within
// a rule in the magical `if` or `unless` keys.
//
// Although we only contain a small number of conditional functions
// they have been moved to a small package to make them self-contained
// and easy to extend without modifying our main package.
package conditionals

import (
	"fmt"
	"strings"
	"sync"
)

// ConditionCall holds the invokation of a conditional expression.
//
// Currently we support a few different types of conditional methods,
// which can only be used as the values for magical blocks "if" and "unless".
//
// New conditional-types can be implemented without touching the parser-code,
// or even the executor, just defining new self-registering classes in the
// conditionals package.
type ConditionCall struct {

	// Name stores the name of the conditional-functions to be called.
	Name string

	// Args contains the arguments to be used for the function invocation.
	Args []string
}

// String converts a ConditionCall to a string.
func (c ConditionCall) String() string {
	return fmt.Sprintf("%s(%s)", c.Name, strings.Join(c.Args, ","))
}

// This is a map of known conditional methods.
var handlers = struct {
	m map[string]Conditional
	sync.RWMutex
}{m: make(map[string]Conditional)}

// Conditional is the signature of an equality-method.
type Conditional func(args []string) (bool, error)

// Register records a new conditional-method, by name
func Register(id string, ptr Conditional) {
	handlers.Lock()
	handlers.m[id] = ptr
	handlers.Unlock()
}

// Lookup is a helper which returns the function implementing
// a given conditional-method, by name.
func Lookup(name string) (a Conditional) {
	handlers.RLock()
	fn, ok := handlers.m[name]
	handlers.RUnlock()
	if ok {
		a = fn
	}
	return
}
