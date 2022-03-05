// Common code for the "group" module.
//
// Execute is implemented in a per-OS fashion.

package modules

import (
	"fmt"
	"regexp"

	mcfg "github.com/skx/marionette/config"
	"github.com/skx/marionette/environment"
)

// GroupModule stores our state
type GroupModule struct {

	// cfg contains our configuration object.
	cfg *mcfg.Config

	// env holds our environment
	env *environment.Environment

	// Regular expression for testing if parameters are safe
	// and won't cause shell injection issues.
	reg *regexp.Regexp
}

// Check is part of the module-api, and checks arguments.
func (g *GroupModule) Check(args map[string]interface{}) error {

	// Required keys for this module
	required := []string{"group", "state"}

	// Ensure they exist.
	for _, key := range required {

		// Get the param
		_, ok := args[key]
		if !ok {
			return fmt.Errorf("missing '%s' parameter", key)
		}

		// Ensure it is a simple string
		val := StringParam(args, key)
		if val == "" {
			return fmt.Errorf("parameter '%s' wasn't a simple string", key)
		}

		// Ensure it has decent characters
		if !g.reg.MatchString(val) {
			return fmt.Errorf("parameter '%s' failed validation", key)
		}

	}

	// Ensure state is one of "present"/"absent"
	state := StringParam(args, "state")
	if state == "absent" {
		return nil
	}
	if state == "present" {
		return nil
	}

	return fmt.Errorf("state must be one of 'absent' or 'present'")
}

// init is used to dynamically register our module.
func init() {
	Register("group", func(cfg *mcfg.Config, env *environment.Environment) ModuleAPI {
		return &GroupModule{
			cfg: cfg,
			env: env,
			reg: regexp.MustCompile(`^[-_/a-zA-Z0-9]+$`),
		}
	})
}
