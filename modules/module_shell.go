package modules

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/skx/marionette/config"
)

// ShellModule stores our state
type ShellModule struct {

	// cfg contains our configuration object.
	cfg *config.Config
}

// Check is part of the module-api, and checks arguments.
func (f *ShellModule) Check(args map[string]interface{}) error {

	// Ensure we have a command to run.
	_, ok := args["command"]
	if !ok {
		return fmt.Errorf("missing 'command' parameter")
	}

	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (f *ShellModule) Execute(args map[string]interface{}) (bool, error) {

	// Get the command
	str := StringParam(args, "command")
	if str == "" {
		return false, fmt.Errorf("missing 'command' parameter")
	}

	// If we see redirection we're good
	if strings.Contains(str, ">") || strings.Contains(str, "<") || strings.Contains(str, "|") {
		str = "bash -c \"" + str + "\""
	}

	// Split on space to execute
	var bits []string
	bits = strings.Split(str, " ")

	// but if we see redirection use the shell instead
	if strings.Contains(str, ">") || strings.Contains(str, "<") {
		bits = []string{"bash", "-c", str}
	}

	// Now run
	cmd := exec.Command(bits[0], bits[1:]...)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("error running command '%s' %s", str, err.Error())
	}

	return true, nil
}

// init is used to dynamically register our module.
func init() {
	Register("shell", func(cfg *config.Config) ModuleAPI {
		return &ShellModule{cfg: cfg}
	})
}
