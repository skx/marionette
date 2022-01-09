//go:build darwin || windows

package modules

import (
	"fmt"
	"log"
	"runtime"

	"github.com/skx/marionette/environment"
)

// Execute is part of the module-api, and is invoked to run a rule.
func (g *GroupModule) Execute(env *environment.Environment, args map[string]interface{}) (bool, error) {

	message := "the 'group' module is not implemented on this platform"

	log.Printf("[ERROR] %s: %s", message, runtime.GOOS)

	return false, fmt.Errorf("%s: %s", message, runtime.GOOS)
}
