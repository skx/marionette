//go:build windows

package modules

import (
	"fmt"
	"log"

	"github.com/skx/marionette/environment"
)

// Execute is part of the module-api, and is invoked to run a rule.
func (g *GroupModule) Execute(env *environment.Environment, args map[string]interface{}) (bool, error) {

	log.Printf("[ERROR] the 'group' module is not implemented upon Windows")

	return false, fmt.Errorf("the 'group' module is not implemented upon Windows")
}
