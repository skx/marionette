// Package executor is the thing that will execute our rules.
//
// This means processing the rules, one by one, but also ensuring
// dependencies are handled.
//
package executor

import (
	"fmt"

	"github.com/skx/marionette/ast"
	"github.com/skx/marionette/conditionals"
	"github.com/skx/marionette/config"
	"github.com/skx/marionette/environment"
	"github.com/skx/marionette/modules"
)

// Executor holds our internal state.
type Executor struct {

	// Program is the series of AST-nodes we'll interpret.
	Program []ast.Node

	// Index is a mapping between rule-name and index.
	//
	// This is required because we expect users to refer to
	// dependencies by name, but when we search for them in
	// our Rules array above we need to efficiently lookup
	// their index.
	index map[string]int

	// cfg holds our configuration options.
	cfg *config.Config

	// env holds the environment.
	env *environment.Environment
}

// New creates a new executor, using the array of AST nodes we should
// execute, which was produced by the parser.
func New(program []ast.Node) *Executor {

	//
	// Setup our state
	//
	e := &Executor{
		env:     environment.New(),
		Program: program,
	}

	//
	// Default configuration
	//
	e.cfg = &config.Config{}

	return e
}

// verbose will output a message if running verbosely
func (e *Executor) verbose(msg string) {
	if e.cfg.Verbose {
		fmt.Printf("%s\n", msg)
	}
}

// SetConfig updates the executor with a configuration object.
func (e *Executor) SetConfig(cfg *config.Config) {
	e.cfg = cfg
}

// Get the rules a rule depends upon, via the given key.
//
// This is used to find any `require` or `notify` rules.
func (e *Executor) deps(rule *ast.Rule, key string) []string {

	var res []string

	requires, ok := rule.Params[key]

	// no requirements?  Awesome
	if !ok {
		return res
	}

	//
	// OK the requirements might be a single rule, or
	// an array of rules.
	//
	// Handle both cases.
	//

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

// Check ensures the rules make sense.
//
// In short this means that we check the dependencies/notifiers listed
// for every rule, and raise an error if they contain references to
// rules which don't exist.
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
	//
	e.index = make(map[string]int)

	//
	// Walk over all the rules we've got
	//
	for i, r := range e.Program {

		// Skip nodes which are not ast.Rules
		rule, ok := r.(*ast.Rule)
		if !ok {
			continue
		}

		_, ok2 := e.index[rule.Name]
		if ok2 {
			return fmt.Errorf("rule names must be unique; we've already seen '%s'", rule.Name)
		}

		e.index[rule.Name] = i
	}

	//
	// For every rule.
	//
	for _, r := range e.Program {

		// Skip nodes which are not ast.Rules
		rule, ok := r.(*ast.Rule)
		if !ok {
			continue
		}

		//
		// Get the dependencies of that rule, and the things
		// it will notify in the event it is triggered.
		//
		deps := e.deps(rule, "require")
		notify := e.deps(rule, "notify")

		// Join the pair of rules
		var all []string
		all = append(all, deps...)
		all = append(all, notify...)

		// nothing to check?  Awesome
		if len(all) < 1 {
			continue
		}

		// for each rule-reference
		for _, dep := range all {

			// Does the requirement exist?
			_, found := e.index[dep]
			if !found {
				return fmt.Errorf("rule '%s' has reference to '%s' which doesn't exist", rule.Params["name"], dep)
			}
		}
	}

	return nil
}

// Execute runs the rules in turn, handling any dependency ordering.
func (e *Executor) Execute() error {

	// Keep track of which rules we've executed
	seen := make(map[int]bool)

	// For each rule ..
	for i, r := range e.Program {

		// Skip nodes which are not ast.Rules
		rule, ok := r.(*ast.Rule)
		if !ok {
			continue
		}

		// Don't run rules that are only present to
		// be notified by a trigger.
		if rule.Triggered {
			continue
		}

		// Have we executed this rule already?
		if seen[i] {
			continue
		}

		// Get the rule dependencies.
		deps := e.deps(rule, "require")

		// Process each one
		for i, dep := range deps {

			// Have we executed this rule already?
			if seen[i] {
				continue
			}

			// get the actual rule, by index
			dr := e.Program[e.index[dep]].(*ast.Rule)

			// Don't run rules that are only present to
			// be notified by a trigger.
			if dr.Triggered {
				continue
			}

			err := e.executeSingleRule(dr)
			if err != nil {
				return err
			}

			// Now we've executed the rule.
			seen[i] = true
		}

		// Now the rule itself
		err := e.executeSingleRule(rule)
		if err != nil {
			return err
		}

		// And mark this as executed too.
		seen[i] = true
	}
	return nil
}

// runConditional returns true if the given conditional is true.
//
// The conditionals are implemented in their own package, and can be
// looked up by name.  We lookup the conditional, and if it exists
// invoke it dynamically returning the appropriate result.
func (e *Executor) runConditional(cond interface{}) (bool, error) {

	// Get the value as an instance of our Conditional struct
	test, ok := cond.(*conditionals.ConditionCall)
	if !ok {
		return false, fmt.Errorf("we expected a conditional structure, but got %v", cond)
	}

	// Look for the implementation of the conditional-method.
	helper := conditionals.Lookup(test.Name)
	if helper == nil {
		return false, fmt.Errorf("conditional-function %s not available", test.Name)
	}

	// Call the function, and return whatever result it gives us.
	res, err := helper(test.Args)
	return res, err
}

// executeSingleRule creates the appropriate module, and runs the single rule.
func (e *Executor) executeSingleRule(rule *ast.Rule) error {

	// Show what we're doing
	e.verbose(fmt.Sprintf("Running %s-module rule: %s", rule.Type, rule.Name))

	//
	// Are there conditionals present?
	//
	if rule.Params["if"] != nil {
		res, err := e.runConditional(rule.Params["if"])
		if err != nil {
			return err
		}
		if !res {
			e.verbose(fmt.Sprintf("\tSkipping rule condition was not true: %s", rule.Params["if"]))
			return nil
		}
	}

	if rule.Params["unless"] != nil {
		res, err := e.runConditional(rule.Params["unless"])
		if err != nil {
			return err
		}
		if res {
			e.verbose(fmt.Sprintf("\tSkipping rule condition was true: %s", rule.Params["unless"]))
			return nil
		}
	}

	// Did this rule-execution result in a change?
	//
	// If so then we'd notify any rules which should be executed
	// as a result of that change.
	var changed bool
	var err error

	// Create the instance of the module
	helper := modules.Lookup(rule.Type, e.cfg)
	if helper == nil {
		return fmt.Errorf("unknown module type %s, from rule %v", rule.Type, rule)
	}

	// Run the module instance
	changed, err = e.runInternalModule(helper, rule)
	if err != nil {
		return err
	}

	if changed {

		e.verbose("\tRule resulted in a change being made.")

		// Now call any rules that we should notify.
		notify := e.deps(rule, "notify")

		// Process each one
		for _, child := range notify {

			// get the actual rule, by index
			dr := e.Program[e.index[child]].(*ast.Rule)

			// report upon it if we're being verbose
			e.verbose(fmt.Sprintf("\t\tNotifying rule: %s", dr.Name))

			// Execute the rule.
			err := e.executeSingleRule(dr)
			if err != nil {
				return err
			}
		}
	}

	// All done
	return nil
}

// runInternalModule executes the given rule with the loaded internal
// module.
func (e *Executor) runInternalModule(helper modules.ModuleAPI, rule *ast.Rule) (bool, error) {

	// Check the arguments
	err := helper.Check(rule.Params)
	if err != nil {
		return false, fmt.Errorf("error validating %s-module rule '%s' %s",
			rule.Type, rule.Name, err.Error())
	}

	// Run the change
	changed, err := helper.Execute(e.env, rule.Params)
	if err != nil {
		return false, fmt.Errorf("error running %s-module rule '%s' %s",
			rule.Type, rule.Name, err.Error())
	}

	return changed, nil
}
