// failure() method - Return true the command executed has a non-zero exit-code

package conditionals

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// Failure takes a single argument and returns true the command executed had a non-zero exit-code.
func Failure(args []string) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("wrong number of args for 'failure': %d != 1", len(args))
	}

	// Build up the thing to run, using a shell so that
	// we can handle pipes/redirection.
	toRun := []string{"/bin/bash", "-c", args[0]}

	// Run the command
	cmd := exec.Command(toRun[0], toRun[1:]...)

	err := cmd.Run()

	var (
		ee *exec.ExitError
		pe *os.PathError
	)

	if errors.As(err, &ee) {
		// ran, but non-zero exit code
		// log.Println("exit code error:", ee.ExitCode())
		return true, nil

	} else if errors.As(err, &pe) {
		// "no such file ...", "permission denied" etc.
		// log.Printf("os.PathError: %v", pe)
		return true, nil

	} else if err != nil {
		// something really bad happened!
		//log.Printf("general error: %v", err)
		return true, nil
	} else {
		return false, nil
	}
}

// init is used to dynamically register our conditional method.
func init() {
	Register("failure", Failure)
}
