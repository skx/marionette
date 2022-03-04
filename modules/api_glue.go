package modules

import (
	"sync"

	"github.com/skx/marionette/config"
	"github.com/skx/marionette/environment"
)

// This is a map of known modules.
var handlers = struct {
	m map[string]ModuleConstructor
	sync.RWMutex
}{m: make(map[string]ModuleConstructor)}

// Register records a new module.
func Register(id string, newfunc ModuleConstructor) {
	handlers.Lock()
	handlers.m[id] = newfunc
	handlers.Unlock()
}

// RegisterAlias allows a new name to refer to an existing implementation.
func RegisterAlias(alias string, impl string) {
	handlers.Lock()
	handlers.m[alias] = handlers.m[impl]
	handlers.Unlock()
}

// Lookup is the factory-method which looks up and returns
// an object of the given type - if possible.
func Lookup(id string, cfg *config.Config, env *environment.Environment) (a ModuleAPI) {
	handlers.RLock()
	ctor, ok := handlers.m[id]
	handlers.RUnlock()
	if ok {
		a = ctor(cfg, env)
	}
	return
}

// Modules returns the names of all the registered module-names.
func Modules() []string {
	var result []string

	// For each handler save the name
	handlers.RLock()
	for index := range handlers.m {
		result = append(result, index)
	}
	handlers.RUnlock()

	// And return the result
	return result

}
