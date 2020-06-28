// exists() method

package conditionals

import (
	"fmt"

	"github.com/skx/marionette/file"
)

// Exists takes a single argument, and returns true if the specified file exists.
func Exists(args []string) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("wrong number of args for 'exists': %d != 1", len(args))
	}

	if file.Exists(args[0]) {
		return true, nil
	}

	return false, nil
}

// init is used to dynamically register our conditional method.
func init() {
	Register("exists", Exists)
}
