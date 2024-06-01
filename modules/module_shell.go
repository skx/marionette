package modules

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/skx/marionette/config"
	"github.com/skx/marionette/environment"
)

// ShellModule stores our state
type ShellModule struct {

	// cfg contains our configuration object.
	cfg *config.Config

	// env holds our environment
	env *environment.Environment

	// Saved copy of STDOUT.
	stdout []byte

	// Saved copy of STDERR.
	stderr []byte
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
func (f *ShellModule) Execute(args map[string]interface{}) (bool, error) {

	// get command(s)
	cmds := ArrayCastParam(args, "command")

	// Ensure we have one or more commands to run.
	if len(cmds) < 1 {
		return false, fmt.Errorf("missing 'command' parameter")
	}

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

	//
	// Should we run using a shell?
	//
	useShell := false

	//
	// Does the user explicitly request the use of a shell?
	//
	shell := StringParam(args, "shell")
	if strings.ToLower(shell) == "true" {
		useShell = true
	}

	//
	// If the user didn't explicitly specify a shell must be used
	// we must do so anyway if we see a redirection, or the use of
	// a pipe.
	//
	if strings.Contains(command, ">") || strings.Contains(command, "&") || strings.Contains(command, "|") || strings.Contains(command, "<") {
		useShell = true
	}

	//
	// By default we split on space to find the things to execute.
	//
	var bits []string
	bits = strings.Split(command, " ")

	//
	// But
	//
	//   If the user explicitly specified the need to use a shell.
	//
	// or
	//
	//   We found a redirection/similar then we must run via a shell.
	//
	if useShell {
		bits = []string{"bash", "-c", command}
	}

	// Show what we're executing.
	log.Printf("[DEBUG] CMD: %s", strings.Join(bits, " "))

	// Now run
	cmd := exec.Command(bits[0], bits[1:]...)

	// Setup buffers for saving STDOUT/STDERR.
	var execOut bytes.Buffer
	var execErr bytes.Buffer

	// Wire up the output buffers.
	//
	// In the past we'd output these to the console, but now we're
	// implementing the GetOutput interface they'll be shown when
	// running with -debug anyway.
	//
	cmd.Stdout = &execOut
	cmd.Stderr = &execErr

	// Run the command
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running command '%s' %s", command, err.Error())
	}

	// Save the outputs
	f.stdout = execOut.Bytes()
	f.stderr = execErr.Bytes()

	return nil
}

// GetOutputs is an optional interface method which allows the
// module to return values to the caller - prefixed by the rule-name.
func (f *ShellModule) GetOutputs() map[string]string {

	// Prepare a map of key->values to return
	m := make(map[string]string)

	// Populate with information from our execution.
	m["stdout"] = strings.TrimSpace(string(f.stdout))
	m["stderr"] = strings.TrimSpace(string(f.stderr))

	return m
}

// init is used to dynamically register our module.
func init() {
	Register("shell", func(cfg *config.Config, env *environment.Environment) ModuleAPI {
		return &ShellModule{
			cfg: cfg,
			env: env,
		}
	})
}
