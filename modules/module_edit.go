package modules

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/skx/marionette/config"
	"github.com/skx/marionette/file"
)

// EditModule stores our state.
type EditModule struct {

	// cfg contains our configuration object.
	cfg *config.Config
}

// Check is part of the module-api, and checks arguments.
func (e *EditModule) Check(args map[string]interface{}) error {

	// Ensure we have a target (i.e. file to operate upon).
	_, ok := args["target"]
	if !ok {
		return fmt.Errorf("missing 'target' parameter")
	}

	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (e *EditModule) Execute(args map[string]interface{}) (bool, error) {

	var ret bool

	// Get the target
	target := StringParam(args, "target")
	if target == "" {
		return false, fmt.Errorf("failed to convert target to string")
	}

	//
	// Now look at our actions
	//

	// Append a line if missing
	append := StringParam(args, "append_if_missing")
	if append != "" {
		changed, err := e.Append(target, append)
		if err != nil {
			return false, err
		}
		if changed {
			ret = true
		}
	}

	// Remove lines matching a regexp.
	remove := StringParam(args, "remove_lines")
	if remove != "" {
		changed, err := e.RemoveLines(target, remove)
		if err != nil {
			return false, err
		}

		if changed {
			ret = true
		}
	}

	return ret, nil
}

// Append the given line to the file, if it is missing.
func (e *EditModule) Append(path string, text string) (bool, error) {

	// If the target file doesn't exist create it
	if !file.Exists(path) {

		f, err := os.OpenFile(path,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return false, err
		}
		defer f.Close()
		if _, err := f.WriteString("\n" + text); err != nil {
			return false, err
		}
		return true, nil
	}

	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Did we find what we're looking for?
	found := false

	// Process line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if text == line {
			found = true
		}
	}

	if err = scanner.Err(); err != nil {
		return false, err
	}

	// If we found the line we do nothing
	if found {
		return false, nil
	}

	// Otherwise we need to append the text
	f, err := os.OpenFile(path,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false, err
	}
	defer f.Close()
	if _, err := f.WriteString("\n" + text); err != nil {
		return false, err
	}

	return true, nil
}

// RemoveLines remove any lines from the file which match the given
// regular expression.
func (e *EditModule) RemoveLines(path string, pattern string) (bool, error) {

	// If the target file doesn't exist then we cannot
	// remove content from it.
	if !file.Exists(path) {
		return false, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	// Open the input file
	in, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer in.Close()

	// Open a temporary file
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		return false, err
	}

	// Process the input file line by line
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {

		// Get the line
		line := scanner.Text()

		// If it doesn't match the regexp, write to the temporary file
		if !re.MatchString(line) {
			tmpfile.WriteString(line + "\n")
		}
	}

	identical, err := file.Identical(tmpfile.Name(), path)
	if err != nil {
		return false, err
	}

	if identical {
		return false, nil
	}

	// otherwise change
	err = file.Copy(tmpfile.Name(), path)
	return true, err
}

// init is used to dynamically register our module.
func init() {
	Register("edit", func(cfg *config.Config) ModuleAPI {
		return &EditModule{cfg: cfg}
	})
}
