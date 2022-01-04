// This is the simple driver to execute the named file(s).
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/skx/marionette/ast"
	"github.com/skx/marionette/config"
	"github.com/skx/marionette/parser"
)

func runFile(filename string, cfg *config.Config) error {

	// Read the file contents.
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	// Create a new parser with our file content.
	p := parser.New(string(data))

	// Parse the rules
	out, err := p.Process()
	if err != nil {
		return err
	}

	//
	// At this point we have a list of AST-nodes.
	//
	// We should process them, in order, however we're going to
	// just dump them to the console for the moment.
	//
	for _, node := range out.Recipe {

		switch node.(type) {

		case *ast.Assign:
			set := node.(*ast.Assign)
			fmt.Printf("Assignment: %s -> %s\n", set.Key, set.Value)
		case *ast.Include:
			inc := node.(*ast.Include)
			fmt.Printf("Include: %s\n", inc.Source)
		case *ast.Rule:
			rul := node.(*ast.Rule)
			fmt.Printf("RULE: %s ..\n", rul.Type)
		default:
			return fmt.Errorf("unknown node type! %t", node)
		}
	}

	// // Now we'll create an executor with the rules
	// ex := executor.New(rules)

	// // Set the configuration options.
	// ex.SetConfig(cfg)

	// // Check for broken dependencies
	// err = ex.Check()
	// if err != nil {
	// 	return err
	// }

	// // Now execute!
	// err = ex.Execute()
	// if err != nil {
	// 	return err
	// }

	return nil
}

// main is our entry-point
func main() {

	// Parse our command-line flags.
	verbose := flag.Bool("verbose", false, "Be verbose in execution")
	flag.Parse()

	// Create our configuration object
	cfg := &config.Config{Verbose: *verbose}

	// Ensure we got at least one recipe to execute.
	if len(flag.Args()) < 1 {
		fmt.Printf("Usage %s file1 file2 .. fileN\n", os.Args[0])
		return
	}

	// Process each given file.
	for _, file := range flag.Args() {
		err := runFile(file, cfg)
		if err != nil {
			fmt.Printf("Error:%s\n", err.Error())
			return
		}
	}

}
