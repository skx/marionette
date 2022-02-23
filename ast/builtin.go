// builtin.go - Contains our built-in primitives.

package ast

import (
	"crypto/md5"
	"crypto/sha1"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/skx/marionette/environment"
	"github.com/skx/marionette/file"
)

// BuiltIn is the signature of a built-in function
type BuiltIn func(env *environment.Environment, args []string) (Object, error)

// FUNCTIONS contains our list of built-in functions, as a map.
//
// The key is the name of the function, and the value is the pointer to the
// function which is used to implement it.
var FUNCTIONS map[string]BuiltIn

// FALSE is a global false-value, which simplifies our function returns
var FALSE = &Boolean{Value: false}

// TRUE is a global true-value, which simplifies our function returns.
var TRUE = &Boolean{Value: true}

// init is called on startup, and creates the FUNCTIONS map which will
// hold our built-in functions.
func init() {

	// create the map to hold function-references
	FUNCTIONS = make(map[string]BuiltIn)

	// Populate it.
	FUNCTIONS["contains"] = fnContains
	FUNCTIONS["empty"] = fnEmpty
	FUNCTIONS["equal"] = fnEqual
	FUNCTIONS["equals"] = fnEqual // duplicate
	FUNCTIONS["exists"] = fnExists
	FUNCTIONS["failure"] = fnFailure
	FUNCTIONS["field"] = fnField
	FUNCTIONS["gt"] = fnGt
	FUNCTIONS["gte"] = fnGte
	FUNCTIONS["len"] = fnLen
	FUNCTIONS["lower"] = fnLower
	FUNCTIONS["lt"] = fnLt
	FUNCTIONS["lte"] = fnLte
	FUNCTIONS["matches"] = fnMatches
	FUNCTIONS["md5"] = fnMD5Sum // duplicate
	FUNCTIONS["md5sum"] = fnMD5Sum
	FUNCTIONS["nonempty"] = fnNonEmpty
	FUNCTIONS["on_path"] = fnOnPath
	FUNCTIONS["set"] = fnNonEmpty // duplicate
	FUNCTIONS["sha1"] = fnSha1Sum // duplicate
	FUNCTIONS["sha1sum"] = fnSha1Sum
	FUNCTIONS["success"] = fnSuccess
	FUNCTIONS["unset"] = fnEmpty // duplicate
	FUNCTIONS["upper"] = fnUpper

}

//
// Now our built-in methods follow
//

// fnContains returns true/false depending upon whether the first string
// contains the second one.
func fnContains(env *environment.Environment, args []string) (Object, error) {

	// Two arguments are required.
	if len(args) != 2 {
		return nil, fmt.Errorf("'contains' requires two arguments")
	}

	if strings.Contains(args[0], args[1]) {
		return TRUE, nil
	}

	return FALSE, nil
}

// fnEmpty returns true if the given string is unset/empty
func fnEmpty(env *environment.Environment, args []string) (Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of args for 'empty': %d != 1", len(args))
	}

	if len(args[0]) == 0 {
		return TRUE, nil
	}

	return FALSE, nil
}

// fnEqual returns true/false depending upon whether the two arguments
// are equal.
func fnEqual(env *environment.Environment, args []string) (Object, error) {

	// Two arguments are required.
	if len(args) != 2 {
		return nil, fmt.Errorf("'equal' requires two arguments")
	}

	// If the values differ then not-equal
	if args[0] != args[1] {
		return FALSE, nil
	}

	// Same values?  Then equal
	return TRUE, nil
}

// fnExists takes a single argument, and returns true if the specified file exists.
func fnExists(env *environment.Environment, args []string) (Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of args for 'exists': %d != 1", len(args))
	}

	if file.Exists(args[0]) {
		return TRUE, nil
	}

	return FALSE, nil
}

// fnFailure returns true if executing the given command fails.
func fnFailure(env *environment.Environment, args []string) (Object, error) {

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
		return TRUE, nil

	} else if errors.As(err, &pe) {
		// "no such file ...", "permission denied" etc.
		// log.Printf("os.PathError: %v", pe)
		return TRUE, nil

	} else if err != nil {
		// something really bad happened!
		//log.Printf("general error: %v", err)
		return TRUE, nil
	} else {
		// No failure
		return FALSE, nil
	}
}

// fnField returns the numbered field from the given text.  Much like awk:
//   field( "Steve Kemp", 0) -> "Steve"
//   field( "Steve Kemp", 1) -> "Kemp"
func fnField(env *environment.Environment, args []string) (Object, error) {

	if len(args) != 2 {
		return nil, fmt.Errorf("wrong number of args for 'field': %d != 2", len(args))
	}

	// Split the first thing into tokens
	fields := strings.Fields(args[0])

	// Get the number to extract
	n, err := strconv.ParseInt(args[1], 0, 64)
	if err != nil {
		return FALSE, err
	}

	if int(n) <= len(fields) {
		return &String{Value: fields[n]}, nil
	}

	log.Printf("[DEBUG] Warning: Field %d out of bounds for input '%s' %d fields available", n, args[0], len(fields))

	return &String{Value: ""}, nil

}

// fnGt compares two numbers to see if the first is greater than the second.
func fnGt(env *environment.Environment, args []string) (Object, error) {

	// Two arguments are required.
	if len(args) != 2 {
		return nil, fmt.Errorf("'gt' requires two arguments")
	}

	a, errA := strconv.ParseInt(args[0], 0, 64)
	if errA != nil {
		return FALSE, errA
	}
	b, errB := strconv.ParseInt(args[1], 0, 64)
	if errB != nil {
		return FALSE, errB
	}

	if a > b {
		return TRUE, nil
	}
	return FALSE, nil

}

// fnGt compares two numbers to see if the first is greater than, or equal
// to the second.
func fnGte(env *environment.Environment, args []string) (Object, error) {

	// Two arguments are required.
	if len(args) != 2 {
		return nil, fmt.Errorf("'gte' requires two arguments")
	}

	a, errA := strconv.ParseInt(args[0], 0, 64)
	if errA != nil {
		return FALSE, errA
	}
	b, errB := strconv.ParseInt(args[1], 0, 64)
	if errB != nil {
		return FALSE, errB
	}

	if a >= b {
		return TRUE, nil
	}
	return FALSE, nil

}

// fnLen returns the length of the given node.
func fnLen(env *environment.Environment, args []string) (Object, error) {

	// Only one argument is supported.
	if len(args) != 1 {
		return nil, fmt.Errorf("'len' requires a single argument")
	}

	return &Number{Value: int64(utf8.RuneCountInString(args[0]))}, nil
}

// fnLt compares two numbers to see if the first is less than the second.
func fnLt(env *environment.Environment, args []string) (Object, error) {

	// Two arguments are required.
	if len(args) != 2 {
		return nil, fmt.Errorf("'lt' requires two arguments")
	}

	a, errA := strconv.ParseInt(args[0], 0, 64)
	if errA != nil {
		return FALSE, errA
	}
	b, errB := strconv.ParseInt(args[1], 0, 64)
	if errB != nil {
		return FALSE, errB
	}

	if a < b {
		return TRUE, nil
	}
	return FALSE, nil

}

// fnLt compares two numbers to see if the first is less than, or equal to,
// the second.
func fnLte(env *environment.Environment, args []string) (Object, error) {

	// Two arguments are required.
	if len(args) != 2 {
		return nil, fmt.Errorf("'lte' requires two arguments")
	}

	a, errA := strconv.ParseInt(args[0], 0, 64)
	if errA != nil {
		return FALSE, errA
	}
	b, errB := strconv.ParseInt(args[1], 0, 64)
	if errB != nil {
		return FALSE, errB
	}

	if a <= b {
		return TRUE, nil
	}
	return FALSE, nil
}

// fnLower converts the given node to lower-case.
func fnLower(env *environment.Environment, args []string) (Object, error) {

	// Only one argument is supported.
	if len(args) != 1 {
		return nil, fmt.Errorf("'lower' requires a single argument")
	}

	return &String{Value: strings.ToLower(args[0])}, nil
}

// fnMatches returns true if the first string matches the regular expression
// specified as the second argument.
func fnMatches(env *environment.Environment, args []string) (Object, error) {

	if len(args) != 2 {
		return nil, fmt.Errorf("wrong number of args for 'matches': %d != 2", len(args))
	}

	reg, err := regexp.Compile(args[1])
	if err != nil {
		return FALSE, err
	}
	res := reg.FindStringSubmatch(args[0])

	if len(res) > 0 {
		return TRUE, nil
	}

	return FALSE, nil
}

// fnMD5Sum returns the MD5 digest of the given input
func fnMD5Sum(env *environment.Environment, args []string) (Object, error) {

	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of args for 'md5sum': %d != 1", len(args))
	}

	h := md5.New()
	h.Write([]byte(args[0]))
	bs := h.Sum(nil)
	return &String{Value: fmt.Sprintf("%x", bs)}, nil
}

// fnNonEmpty returns true if the given string is not unset/empty
func fnNonEmpty(env *environment.Environment, args []string) (Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of args for 'nonempty': %d != 1", len(args))
	}

	if len(args[0]) > 0 {
		return TRUE, nil
	}

	return FALSE, nil
}

// fnOnPath returns true if the given binary can be found on the users' PATH
func fnOnPath(env *environment.Environment, args []string) (Object, error) {

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
		return FALSE, nil
	}

	log.Printf("[DEBUG] lookup resulted in '%s'", path)

	// found a path?  Then success
	if path != "" {
		return TRUE, nil
	}

	// Not found
	return FALSE, nil
}

// fnSha1Sum returns the SHA1 digest of the given input
func fnSha1Sum(env *environment.Environment, args []string) (Object, error) {

	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of args for 'sha1sum': %d != 1", len(args))
	}

	h := sha1.New()
	h.Write([]byte(args[0]))
	bs := h.Sum(nil)
	return &String{Value: fmt.Sprintf("%x", bs)}, nil
}

// fnSuccess returns true if executing the given command succeeds.
func fnSuccess(env *environment.Environment, args []string) (Object, error) {

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
		return FALSE, nil

	} else if err != nil {
		// something really bad happened!
		//log.Printf("general error: %v", err)
		return FALSE, nil
	} else {
		// no error
		return TRUE, nil
	}
}

// fnUpper converts the given node to upper-case.
func fnUpper(env *environment.Environment, args []string) (Object, error) {

	// Only one argument is supported.
	if len(args) != 1 {
		return nil, fmt.Errorf("'upper' requires a single argument")
	}

	return &String{Value: strings.ToUpper(args[0])}, nil
}
