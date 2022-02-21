package ast

import (
	"crypto/md5"
	"crypto/sha1"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"unicode/utf8"

	"github.com/skx/marionette/environment"
	"github.com/skx/marionette/file"
)

// fnContains returns true/false depending upon whether the first string
// contains the second one.
func fnContains(env *environment.Environment, args []string) (Node, error) {

	// Two arguments are required.
	if len(args) != 2 {
		return nil, fmt.Errorf("'contains' requires two arguments")
	}

	if strings.Contains(args[0], args[1]) {
		return &Boolean{Value: true}, nil
	}

	return &Boolean{Value: false}, nil
}

// fnEmpty returns true if the given string is unset/empty
func fnEmpty(env *environment.Environment, args []string) (Node, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of args for 'empty': %d != 1", len(args))
	}

	if len(args[0]) == 0 {
		return &Boolean{Value: true}, nil
	}

	return &Boolean{Value: false}, nil
}

// fnEqual returns true/false depending upon whether the two arguments
// are equal.
func fnEqual(env *environment.Environment, args []string) (Node, error) {

	// Two arguments are required.
	if len(args) != 2 {
		return nil, fmt.Errorf("'equal' requires two arguments")
	}

	// If the values differ then not-equal
	if args[0] != args[1] {
		return &Boolean{Value: false}, nil
	}

	// Same values?  Then equal
	return &Boolean{Value: true}, nil
}

// fnExists takes a single argument, and returns true if the specified file exists.
func fnExists(env *environment.Environment, args []string) (Node, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of args for 'exists': %d != 1", len(args))
	}

	if file.Exists(args[0]) {
		return &Boolean{Value: true}, nil
	}

	return &Boolean{Value: false}, nil
}

// fnFailure returns true if executing the given command fails.
func fnFailure(env *environment.Environment, args []string) (Node, error) {

	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of args for 'failure': %d != 1", len(args))
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
		return &Boolean{Value: true}, nil

	} else if errors.As(err, &pe) {
		// "no such file ...", "permission denied" etc.
		// log.Printf("os.PathError: %v", pe)
		return &Boolean{Value: true}, nil

	} else if err != nil {
		// something really bad happened!
		//log.Printf("general error: %v", err)
		return &Boolean{Value: true}, nil
	} else {
		// No failure
		return &Boolean{Value: false}, nil
	}
}

// fnLen returns the length of the given node.
func fnLen(env *environment.Environment, args []string) (Node, error) {

	// Only one argument is supported.
	if len(args) != 1 {
		return nil, fmt.Errorf("'len' requires a single argument")
	}

	return &Number{Value: int64(utf8.RuneCountInString(args[0]))}, nil
}

// fnLower converts the given node to lower-case.
func fnLower(env *environment.Environment, args []string) (Node, error) {

	// Only one argument is supported.
	if len(args) != 1 {
		return nil, fmt.Errorf("'lower' requires a single argument")
	}

	return &String{Value: strings.ToLower(args[0])}, nil
}

// fnMD5Sum returns the MD5 digest of the given input
func fnMD5Sum(env *environment.Environment, args []string) (Node, error) {

	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of args for 'md5sum': %d != 1", len(args))
	}

	h := md5.New()
	h.Write([]byte(args[0]))
	bs := h.Sum(nil)
	return &String{Value: fmt.Sprintf("%x", bs)}, nil
}

// fnNonEmpty returns true if the given string is not unset/empty
func fnNonEmpty(env *environment.Environment, args []string) (Node, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of args for 'nonempty': %d != 1", len(args))
	}

	if len(args[0]) > 0 {
		return &Boolean{Value: true}, nil
	}

	return &Boolean{Value: false}, nil
}

// fnOnPath returns true if the given binary can be found on the users' PATH
func fnOnPath(env *environment.Environment, args []string) (Node, error) {

	// only one argument is allowed
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of args for 'on_path': %d != 1", len(args))
	}

	arg := args[0]
	log.Printf("[DEBUG] Looking for %s on PATH", arg)

	// Do the lookup.
	path, err := exec.LookPath(arg)

	// got an error? return it
	if err != nil {
		log.Printf("[DEBUG] error running lookup %s", err)
		return &Boolean{Value: false}, nil
	}

	log.Printf("[DEBUG] lookup resulted in '%s'", path)

	// found a path?  Then success
	if path != "" {
		return &Boolean{Value: true}, nil
	}

	// Not found
	return &Boolean{Value: false}, nil
}

// fnSha1Sum returns the SHA1 digest of the given input
func fnSha1Sum(env *environment.Environment, args []string) (Node, error) {

	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of args for 'sha1sum': %d != 1", len(args))
	}

	h := sha1.New()
	h.Write([]byte(args[0]))
	bs := h.Sum(nil)
	return &String{Value: fmt.Sprintf("%x", bs)}, nil
}

// fnSuccess returns true if executing the given command succeeds.
func fnSuccess(env *environment.Environment, args []string) (Node, error) {

	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of args for 'success': %d != 1", len(args))
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
		return &Boolean{Value: false}, nil

	} else if errors.As(err, &pe) {
		// "no such file ...", "permission denied" etc.
		// log.Printf("os.PathError: %v", pe)
		return &Boolean{Value: false}, nil

	} else if err != nil {
		// something really bad happened!
		//log.Printf("general error: %v", err)
		return &Boolean{Value: false}, nil
	} else {
		// no error
		return &Boolean{Value: true}, nil
	}
}

// fnUpper converts the given node to upper-case.
func fnUpper(env *environment.Environment, args []string) (Node, error) {

	// Only one argument is supported.
	if len(args) != 1 {
		return nil, fmt.Errorf("'upper' requires a single argument")
	}

	return &String{Value: strings.ToUpper(args[0])}, nil
}

func init() {
	FUNCTIONS["contains"] = fnContains
	FUNCTIONS["empty"] = fnEmpty
	FUNCTIONS["equal"] = fnEqual
	FUNCTIONS["equals"] = fnEqual
	FUNCTIONS["exists"] = fnExists
	FUNCTIONS["failure"] = fnFailure
	FUNCTIONS["len"] = fnLen
	FUNCTIONS["lower"] = fnLower
	FUNCTIONS["md5sum"] = fnMD5Sum
	FUNCTIONS["md5"] = fnMD5Sum
	FUNCTIONS["nonempty"] = fnNonEmpty
	FUNCTIONS["on_path"] = fnOnPath
	FUNCTIONS["set"] = fnNonEmpty
	FUNCTIONS["sha1sum"] = fnSha1Sum
	FUNCTIONS["sha1"] = fnSha1Sum
	FUNCTIONS["success"] = fnSuccess
	FUNCTIONS["unset"] = fnEmpty
	FUNCTIONS["upper"] = fnUpper
}
