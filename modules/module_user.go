// Common code for the "user" module.
//
// Execute is implemented in a per-OS fashion.

package modules

import (
	"fmt"
	"regexp"

	mcfg "github.com/skx/marionette/config"
)

// UserModule stores our state
type UserModule struct {

	// cfg contains our configuration object.
	cfg *mcfg.Config

	// Regular expression for testing if parameters are safe
	// and won't cause shell injection issues.
	reg *regexp.Regexp
}

// Check is part of the module-api, and checks arguments.
func (g *UserModule) Check(args map[string]interface{}) error {

	// Required keys for this module
	required := []string{"login", "state"}

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

// verbose will show the message if the verbose flag is set
func (g *UserModule) verbose(msg string) {
	if g.cfg.Verbose {
		fmt.Printf("%s\n", msg)
	}
}

// init is used to dynamically register our module.
func init() {
	Register("user", func(cfg *mcfg.Config) ModuleAPI {
		return &UserModule{
			cfg: cfg,
			reg: regexp.MustCompile(`^[-_/a-zA-Z0-9]+$`),
		}
	})
}
