// equals() method

package conditionals

import (
	"fmt"
)

// Equals takes a pair of arguments, and returns true they are equal.
func Equals(args []string) (bool, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("wrong number of args for 'equals': %d != 2", len(args))
	}

	if args[0] == args[1] {
		return true, nil
	}

	return false, nil
}

// init is used to dynamically register our conditional method.
func init() {
	Register("equals", Equals)
	Register("equal", Equals)
}
