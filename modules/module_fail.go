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
}

// Check is part of the module-api, and checks arguments.
func (f *FailModule) Check(args map[string]interface{}) error {

	// Ensure we have a message to abort with.
	_, ok := args["message"]
	if !ok {
		return fmt.Errorf("missing 'message' parameter")
	}

	// Ensure the message is a string
	msg := StringParam(args, "message")
	if msg == "" {
		return fmt.Errorf("failed to convert 'message' to string")
	}

	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (f *FailModule) Execute(env *environment.Environment, args map[string]interface{}) (bool, error) {

	// Get the message
	str := StringParam(args, "message")
	if str == "" {
		return false, fmt.Errorf("missing 'message' parameter")
	}

	// Show it, and terminate
	fmt.Fprintf(os.Stderr, "FAIL: %s\n", str)

	// Return an error
	return false, fmt.Errorf("%s", str)

}

// init is used to dynamically register our module.
func init() {
	Register("fail", func(cfg *config.Config) ModuleAPI {
		return &FailModule{cfg: cfg}
	})
}
