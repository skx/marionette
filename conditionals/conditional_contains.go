// equals() method

package conditionals

import (
	"fmt"
	"strings"
)

// Contains takes a pair of arguments, and returns true if the first string
// contains the second one.
func Contains(args []string) (bool, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("wrong number of args for 'contains': %d != 2", len(args))
	}

	if strings.Contains(args[0], args[1]) {
		return true, nil
	}

	return false, nil
}

// init is used to dynamically register our conditional method.
func init() {
	Register("contains", Contains)
}
