package modules

import (
	"fmt"
	"os"

	"github.com/skx/marionette/config"
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

	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (f *FailModule) Execute(args map[string]interface{}) (bool, error) {

	// Get the message
	str := StringParam(args, "message")
	if str == "" {
		return false, fmt.Errorf("missing 'message' parameter")
	}

	// Show it, and terminate
	fmt.Fprintf(os.Stderr, "FAIL: %s\n", str)
	os.Exit(1)

	return true, nil
}

// init is used to dynamically register our module.
func init() {
	Register("fail", func(cfg *config.Config) ModuleAPI {
		return &FailModule{cfg: cfg}
	})
}
