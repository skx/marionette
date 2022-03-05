package modules

import (
	"fmt"
	"os"

	"github.com/skx/marionette/config"
	"github.com/skx/marionette/environment"
)

// FailModule stores our state.
type FailModule struct {

	// cfg contains our configuration object.
	cfg *config.Config

	// env holds our environment
	env *environment.Environment
}

// Check is part of the module-api, and checks arguments.
func (f *FailModule) Check(args map[string]interface{}) error {

	// Ensure we have a message to abort with.
	_, ok := args["message"]
	if !ok {
		return fmt.Errorf("missing 'message' parameter")
	}

	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (f *FailModule) Execute(args map[string]interface{}) (bool, error) {

	// Get the message/messages to log.
	arg, ok := args["message"]

	// Ensure that we've got something
	if !ok {
		return false, fmt.Errorf("missing 'message' parameter")
	}

	// A single string?  Show it, and return it as an error.
	str, ok := arg.(string)
	if ok {
		fmt.Fprintf(os.Stderr, "FAIL: %s\n", str)
		return false, fmt.Errorf("%s", str)
	}

	// otherwise we assume it is an array of strings
	strs := arg.([]string)

	// process each argument
	complete := ""
	for _, str = range strs {
		fmt.Fprintf(os.Stderr, "FAIL: %s\n", str)
		complete += str + "\n"
	}

	// Return the joined error-message
	return false, fmt.Errorf("%s", complete)

}

// init is used to dynamically register our module.
func init() {
	Register("fail", func(cfg *config.Config, env *environment.Environment) ModuleAPI {
		return &FailModule{
			cfg: cfg,
			env: env,
		}
	})
}
