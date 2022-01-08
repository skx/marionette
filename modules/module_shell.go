package modules

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/skx/marionette/config"
	"github.com/skx/marionette/environment"
)

// ShellModule stores our state
type ShellModule struct {

	// cfg contains our configuration object.
	cfg *config.Config
}

// Check is part of the module-api, and checks arguments.
func (f *ShellModule) Check(args map[string]interface{}) error {

	// Ensure we have one or more commands to run.
	_, ok := args["command"]
	if !ok {
		return fmt.Errorf("missing 'command' parameter")
	}

	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (f *ShellModule) Execute(env *environment.Environment, args map[string]interface{}) (bool, error) {

	// Ensure we have one or more commands to run.
	_, ok := args["command"]
	if !ok {
		return false, fmt.Errorf("missing 'command' parameter")
	}

	// Get the argument
	arg := args["command"]

	// if it is a string process it
	str, ok := arg.(string)
	if ok {

		// we return an error if the command failed
		err := f.executeSingle(str, args)
		if err != nil {
			return false, err
		}

		// otherwise we always assume a change was made
		return true, nil
	}

	// otherwise we assume it is an array of commands
	cmds := arg.([]string)

	// process each argument
	for _, cmd := range cmds {

		// Run this command
		err := f.executeSingle(cmd, args)

		// process any error
		if err != nil {
			return false, err
		}
	}

	// shell commands always result in a change
	return true, nil
}

// executeSingle executes a single command.
//
// All parameters are available, as is the string command to run.
func (f *ShellModule) executeSingle(command string, args map[string]interface{}) error {

	// Show what we're doing.
	log.Printf("[INFO] Executing: %s", command)

	// Split on space to execute
	var bits []string
	bits = strings.Split(command, " ")

	// but if we see redirection, or the use of a pipe, use the shell instead
	if strings.Contains(command, ">") || strings.Contains(command, "&") || strings.Contains(command, "|") || strings.Contains(command, "<") {
		bits = []string{"bash", "-c", command}
	}

	// Now run
	cmd := exec.Command(bits[0], bits[1:]...)

	// If we're hiding the output we'll write it here.
	var execOut bytes.Buffer
	var execErr bytes.Buffer

	// Show to the console if we should
	if f.cfg.Debug {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	} else {
		// Otherwise pipe to the buffer, and ignore it.
		cmd.Stdout = &execOut
		cmd.Stderr = &execErr
	}

	// Run the command
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running command '%s' %s", command, err.Error())
	}

	return nil
}

// init is used to dynamically register our module.
func init() {
	Register("shell", func(cfg *config.Config) ModuleAPI {
		return &ShellModule{cfg: cfg}
	})
}
