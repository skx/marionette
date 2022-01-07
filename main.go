// This is the simple driver to execute the named file(s).
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/hashicorp/logutils"
	"github.com/skx/marionette/config"
	"github.com/skx/marionette/executor"
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
	out, err := p.Parse()
	if err != nil {
		return err
	}

	// Now we'll create an executor with the program
	ex := executor.New(out.Recipe)

	// Set the configuration options.
	ex.SetConfig(cfg)

	// Mark the file as having been processed.
	ex.MarkSeen(filename)

	// // Set "magic" variables for the current include file.
	err = ex.SetMagicIncludeVars(filename)
	if err != nil {
		return err
	}

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

	// Parse our command-line flags.
	debug := flag.Bool("debug", false, "Be very verbose in logging.")
	verbose := flag.Bool("verbose", false, "Show logs when executing")
	flag.Parse()

	// By default we set the log-level to "USER", which will
	// allow the user-generated messages from our log-module
	// to be visible.
	//
	// If we're running with -verbose we'll show "INFO", and
	// if we're called with -debug we'll show DEBUG
	// running verbosely we'll show info.
	dbg := logutils.LogLevel("DEBUG")
	inf := logutils.LogLevel("INFO")
	usr := logutils.LogLevel("USER")

	// default to user
	lvl := usr
	if *verbose {
		lvl = inf
	}
	if *debug {
		lvl = dbg
	}

	// Setup the filter
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "USER", "ERROR"},
		MinLevel: lvl,
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)

	// Create our configuration object
	cfg := &config.Config{
		Debug:   *debug,
		Verbose: *verbose,
	}

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
