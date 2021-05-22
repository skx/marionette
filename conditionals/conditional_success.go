// success() method - Return true the command executed has a zero-exit-code

package conditionals

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// Success takes a single argument and returns true the command executed had a zero-exit-code.
func Success(args []string) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf("wrong number of args for 'success': %d != 1", len(args))
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
		return false, nil

	} else if errors.As(err, &pe) {
		// "no such file ...", "permission denied" etc.
		// log.Printf("os.PathError: %v", pe)
		return false, nil

	} else if err != nil {
		// something really bad happened!
		//log.Printf("general error: %v", err)
		return false, nil
	} else {
		return true, nil
	}
}

// init is used to dynamically register our conditional method.
func init() {
	Register("success", Success)
}
