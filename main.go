// This is the simple driver to execute the named file(s).
package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/skx/marionette/executor"
	"github.com/skx/marionette/parser"
)

func runFile(filename string) error {

	// Read the file contents.
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	// Create a new parser.
	p := parser.New(string(data))

	// Parse the rules
	rules, err := p.Parse()
	if err != nil {
		return err
	}

	// Now we'll create an executor with the rules
	ex := executor.New(rules)

	// Check for broken dependencies
	err = ex.Check()
	if err != nil {
		return err
	}

	// Now execute!
	err = ex.Execute()
	if err != nil {
		return err
	}

	return nil
}

// main is our entry-point
func main() {

	// Ensure we got an argument, or two.
	if len(os.Args) < 1 {
		fmt.Printf("Usage %s file1 file2 .. fileN\n", os.Args[0])
		return
	}

	// Process each given file.
	for _, file := range os.Args[1:] {
		err := runFile(file)
		if err != nil {
			fmt.Printf("Error:%s\n", err.Error())
			return
		}
	}

}
