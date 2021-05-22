// false() method - Return true if the argument is an empty string

package conditionals

import (
	"fmt"
)

// True takes a single argument and returns true if that argument was an empty
// string.
func False(args []string) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("wrong number of args for 'false': %d != 1", len(args))
	}

	if len(args[0]) == 0 {
		return true, nil
	}

	return false, nil
}

// init is used to dynamically register our conditional method.
func init() {
	Register("false", False)
}
