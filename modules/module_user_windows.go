//go:build windows

package modules

import (
	"fmt"

	"github.com/skx/marionette/environment"
)

// Execute is part of the module-api, and is invoked to run a rule.
func (g *UserModule) Execute(env *environment.Environment, args map[string]interface{}) (bool, error) {

	if g.cfg.Verbose {
		fmt.Printf("'user' module is not implemented upon Windows\n")
	}

	return false, nil
}
