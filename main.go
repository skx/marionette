// This is the simple test-driver which attempts to parse a fixed file,
// and output the rules.
package main

import (
	"fmt"
	"io/ioutil"

	"github.com/skx/marionette/executor"
	"github.com/skx/marionette/parser"
)

func main() {

	// Read a file and create a parser from its contents
	data, err := ioutil.ReadFile("input.txt")
	if err != nil {
		panic(err)
	}

	p := parser.New(string(data))

	// Parse the reuls
	rules, err := p.Parse()

	if err != nil {
		fmt.Printf("Error parsing file:%v\n", err.Error())
		return
	}

	// Now we'll create an executor with the rules
	ex := executor.New(rules)

	// Check for broken dependencies
	err = ex.Check()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}

	// Now execute!
	err = ex.Execute()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}
}
