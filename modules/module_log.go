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

	// env holds our environment
	env *environment.Environment
}

// Check is part of the module-api, and checks arguments.
func (f *LogModule) Check(args map[string]interface{}) error {

	// Ensure we have a message, or messages, to log.
	_, ok := args["message"]
	if !ok {
		return fmt.Errorf("missing 'message' parameter")
	}

	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (f *LogModule) Execute(args map[string]interface{}) (bool, error) {

	// Get the message/messages to log.
	arg, ok := args["message"]

	// Ensure that we've got something
	if !ok {
		return false, fmt.Errorf("missing 'message' parameter")
	}

	// string?
	str, ok := arg.(string)
	if ok {
		log.Print("[USER] " + str)
		return true, nil
	}

	// otherwise we assume it is an array of messages
	strs := arg.([]string)

	// process each argument
	for _, str = range strs {
		log.Print("[USER] " + str)
	}

	return true, nil
}

// init is used to dynamically register our module.
func init() {
	Register("log", func(cfg *config.Config, env *environment.Environment) ModuleAPI {
		return &LogModule{
			cfg: cfg,
			env: env,
		}
	})
}
