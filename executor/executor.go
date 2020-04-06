// Package executor is the thing that will execute our rules.
//
// This means processing the rules, one by one, but also ensuring
// dependencies are handled.
//
package executor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/skx/marionette/modules"
	"github.com/skx/marionette/rules"
)

// Executor holds our internal state.
type Executor struct {

	// Rules are the things we'll execute
	Rules []rules.Rule

	// Index is a mapping between rule-name and index
	index map[string]int
}

// New creates a new executor
func New(r []rules.Rule) *Executor {
	return &Executor{Rules: r}
}

// Get the rules a rule depends upon, via the given key.
//
// This is used to find any `requires` or `notify` rules.
func (e *Executor) deps(rule rules.Rule, key string) []string {

	var res []string

	requires, ok := rule.Params[key]

	// no requirements?  Awesome
	if !ok {
		return res
	}

	// OK the requirements might be a single rule, or
	// an array of rules
	str, ok := requires.(string)
	if ok {
		res = append(res, str)
		return res
	}

	strs, ok := requires.([]string)
	if ok {
		return strs
	}

	return res
}

// Check ensures the rules make sense
func (e *Executor) Check() error {

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
	e.index = make(map[string]int)

	for i, r := range e.Rules {

		_, ok := e.index[r.Name]
		if ok {
			return fmt.Errorf("rule names must be unique; we've already seen '%s'", r.Name)
		}

		e.index[r.Name] = i
	}

	//
	// Look at dependencies
	//
	for _, r := range e.Rules {

		// Get the dependencies
		deps := e.deps(r, "requires")

		// no requirements?  Awesome
		if len(deps) < 1 {
			continue
		}

		for _, dep := range deps {

			// Does the requirement exist?
			_, found := e.index[dep]
			if !found {
				return fmt.Errorf("rule '%s' has dependency '%s' which doesn't exist", r.Params["name"], dep)
			}
		}
	}

	return nil
}

// Execute runs the rules in turn, handling any dependency ordering.
func (e *Executor) Execute() error {

	// For each rule ..
	for _, r := range e.Rules {

		// Don't run rules that are only present to
		// be notified by a trigger.
		if r.Triggered {
			continue
		}

		// Get the rule dependencies
		deps := e.deps(r, "requires")

		// Process each one
		for _, dep := range deps {

			// get the actual rule, by index
			dr := e.Rules[e.index[dep]]
			err := e.ExecuteRule(dr)
			if err != nil {
				return err
			}
		}

		// Now the rule itself
		err := e.ExecuteRule(r)
		if err != nil {
			return err
		}
	}
	return nil
}

// ExecuteRule creates the appropriate module, and runs the single rule.
func (e *Executor) ExecuteRule(rule rules.Rule) error {

	// Show what we're doing
	fmt.Printf("Running %s-module rule: %s\n", rule.Type, rule.Name)

	// Did this rule-execution result in a change?
	var changed bool

	// Create the instance of the module
	helper := modules.Lookup(rule.Type)
	if helper != nil {

		// Check the arguments
		err := helper.Check(rule.Params)
		if err != nil {
			return fmt.Errorf("error validating %s-module rule '%s' %s",
				rule.Type, rule.Name, err.Error())
		}

		// Run the change
		changed, err = helper.Execute(rule.Params)
		if err != nil {
			return fmt.Errorf("error running %s-module rule '%s' %s",
				rule.Type, rule.Name, err.Error())
		}
	} else {

		path := os.Getenv("HOME") + "/.marionette/plugins/" + rule.Type
		cmd := strings.Split(path, " ")
		login := exec.Command(cmd[0], cmd[1:]...)

		buffer := bytes.Buffer{}
		result := bytes.Buffer{}
		input, _ := json.Marshal(rule.Params)
		buffer.Write(input)

		login.Stdout = &result
		login.Stdin = &buffer

		err := login.Run()
		if err != nil {
			fmt.Printf("Error running plugin %s: %s\n", rule.Type, err)
			return err
		}

		// What did we get ?
		res := result.String()
		if res == "changed" {
			changed = true
		}
	}

	if changed {

		// Now call any rules that we should notify.
		notify := e.deps(rule, "notify")

		// Process each one
		for _, child := range notify {

			// get the actual rule, by index
			dr := e.Rules[e.index[child]]

			// Execute the rule.
			err := e.ExecuteRule(dr)
			if err != nil {
				return err
			}
		}
	}

	// All done
	return nil
}
