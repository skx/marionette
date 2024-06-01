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
	strs := ArrayCastParam(args, "message")

	// Ensure that we've got something
	if len(strs) < 1 {
		return false, fmt.Errorf("missing 'message' parameter")
	}

	// process each argument
	complete := ""
	for _, str := range strs {
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
