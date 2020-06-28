// Package conditional contains the implementation of our conditional
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
	"sync"
)

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
