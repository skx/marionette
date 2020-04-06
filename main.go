// This is the simple test-driver which attempts to parse a fixed file,
// and output the rules.
package main

import (
	"fmt"
	"io/ioutil"

	"github.com/skx/marionette/parser"
	"github.com/skx/marionette/rules"
)

// TODO: Here we need to lookup rules by name.
func processRule(rule rules.Rule) {

	fmt.Printf("Processing rule type %s - %s\n", rule.Type, rule.Name)
}

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

	// OK at this point we have a list of rules.
	//
	// We want to loop over each one and create a map so that
	// we can lookup rules by name.
	//
	// i.e. If a rule 1 depends upon rule 10 we want to find
	// that out in advance.
	//
	// We'll also make sure we don't try to notify/depend upon
	// a rule that we can't find.
	index := make(map[string]int)

	for i, r := range rules {
		index[r.Name] = i
	}

	//
	// Look at dependencies
	//
	for _, r := range rules {

		requires, ok := r.Params["requires"]

		// no requirements?  Awesome
		if !ok {
			continue
		}

		// OK the requirements might be a single rule, or
		// an array of rules
		str, ok := requires.(string)
		if ok {

			// Does the single requirement exist?
			_, found := index[str]
			if !found {
				fmt.Printf("rule '%s' has dependency '%s' which doesn't exist", r.Params["name"], str)
				return
			}
		}

		// Might have an array of strings
		strs, ok := requires.([]string)
		if ok {

			for _, str := range strs {

				// Does the requirement exist?
				_, found := index[str]
				if !found {
					fmt.Printf("rule '%s' has dependency '%s' which doesn't exist", r.Params["name"], str)
					return
				}
			}
		}
	}

	//
	// OK at this point we have a list of rules.
	//
	// We can process them in order.
	//
	// Except we have to handle any dependencies first.
	//
	for _, r := range rules {
		processRule(r)
	}
}
