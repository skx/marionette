// on_path() method

package conditionals

import (
	"fmt"
	"log"
	"os/exec"
)

// OnPath takes a single argument and looks for the given name as a binary on
// the users' path.
func OnPath(args []string) (bool, error) {

	// only one argument is allowed
	if len(args) != 1 {
		return false, fmt.Errorf("wrong number of args for 'on_path': %d != 1", len(args))
	}

	arg := args[0]
	log.Printf("[DEBUG] Looking for %s on PATH", arg)

	// Do the lookup.
	path, err := exec.LookPath(arg)

	// got an error? return it
	if err != nil {
		log.Printf("[DEBUG] error running lookup %s", err)
		return false, nil
	}

	log.Printf("[DEBUG] lookup resulted in '%s'", path)

	// found a path?  Then success
	if path != "" {
		return true, nil
	}

	// Not found
	return false, nil
}

// init is used to dynamically register our conditional method.
func init() {
	Register("on_path", OnPath)
}
