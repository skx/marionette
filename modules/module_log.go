package modules

import (
	"fmt"
	"log"

	"github.com/skx/marionette/config"
	"github.com/skx/marionette/environment"
)

// LogModule stores our state
type LogModule struct {

	// cfg contains our configuration object.
	cfg *config.Config
}

// Check is part of the module-api, and checks arguments.
func (f *LogModule) Check(args map[string]interface{}) error {

	// Ensure we have a message to log.
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
func (f *LogModule) Execute(env *environment.Environment, args map[string]interface{}) (bool, error) {

	// Get the message
	str := StringParam(args, "message")
	if str == "" {
		return false, fmt.Errorf("missing 'message' parameter")
	}

	// Show the message
	log.Print("[USER] " + str)

	// Log always results in a change
	return true, nil

}

// init is used to dynamically register our module.
func init() {
	Register("log", func(cfg *config.Config) ModuleAPI {
		return &LogModule{cfg: cfg}
	})
}
