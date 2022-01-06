//go:build windows

package modules

import (
	"fmt"
	"log"

	"github.com/skx/marionette/environment"
)

// Execute is part of the module-api, and is invoked to run a rule.
func (g *UserModule) Execute(env *environment.Environment, args map[string]interface{}) (bool, error) {

	log.Printf("[ERROR] the 'user' module is not implemented upon Windows")

	return false, fmt.Errorf("the 'user' module is not implemented upon Windows")
}
