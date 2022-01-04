// Package executor is the thing that will execute our rules.
//
// This means processing the rules, one by one, but also ensuring
// dependencies are handled.
//
package executor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/skx/marionette/ast"
	"github.com/skx/marionette/conditionals"
	"github.com/skx/marionette/config"
	"github.com/skx/marionette/environment"
	"github.com/skx/marionette/modules"
	"github.com/skx/marionette/token"
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
		cfg:     &config.Config{},
		env:     environment.New(),
		Program: program,
	}

	return e
}

// verbose will output a message only if running verbosely.
func (e *Executor) verbose(msg string) {
	if e.cfg.Verbose {
		fmt.Printf("%s\n", msg)
	}
}

// SetConfig updates the executor with the specified configuration object.
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
	// Walk over all the nodes we've got
	//
	for i, r := range e.Program {

		//
		// Skip nodes which are not ast.Rules
		//
		rule, ok := r.(*ast.Rule)
		if !ok {
			continue
		}

		//
		// Find the index of the name, to ensure it is unique.
		//
		_, ok2 := e.index[rule.Name]
		if ok2 {
			return fmt.Errorf("rule names must be unique; we've already seen '%s'", rule.Name)
		}

		//
		// Save the index away
		//
		e.index[rule.Name] = i
	}

	//
	// For every node in our program.
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
				return fmt.Errorf("rule '%s' has reference to '%s' which doesn't exist", rule.Name, dep)
			}
		}
	}

	return nil
}

// Execute runs the rules in turn, handling any dependency ordering.
func (e *Executor) Execute() error {

	// Keep track of which rules we've executed
	seen := make(map[int]bool)

	// For each node in our program
	for i, r := range e.Program {

		// Test the type to see what we should do.
		switch r.(type) {

		case *ast.Assign:
			err := e.execute_Assign(r.(*ast.Assign))
			if err != nil {
				return err
			}
		case *ast.Include:
			err := e.execute_Include(r.(*ast.Include))
			if err != nil {
				return err
			}
		case *ast.Rule:
			rule := r.(*ast.Rule)
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
			return nil
		default:
			return fmt.Errorf("unknown node type! %t", r)
		}
	}
	return nil
}

// execute_Assign executes an assignment
func (e *Executor) execute_Assign(assign *ast.Assign) error {

	key := assign.Key
	val := assign.Value
	ret := ""
	var err error

	e.verbose(fmt.Sprintf("Setting variable %s -> %s", key, val))

	switch val.Type {
	case token.STRING:
		ret = val.Literal
	case token.BACKTICK:
		ret, err = e.expand(val)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled type in execute_Assign %t", val)
	}

	e.env.Set(key, ret)
	return nil
}
func (e *Executor) execute_Include(inc *ast.Include) error {
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

	// Expand all params
	params := make(map[string]interface{})

	for k, v := range rule.Params {

		// param is a string?  expand it
		str, ok := v.(string)
		if ok {
			params[k] = os.Expand(str, e.mapper)
			continue
		}

		// param is a string array?  expand them
		strs, ok2 := v.([]string)
		if ok2 {
			var tmp []string
			var t string

			for _, x := range strs {
				t = os.Expand(x, e.mapper)
				tmp = append(tmp, t)
			}
			params[k] = tmp
			continue
		}
	}

	// Run the change
	changed, err := helper.Execute(e.env, params)
	if err != nil {
		return false, fmt.Errorf("error running %s-module rule '%s' %s",
			rule.Type, rule.Name, err.Error())
	}

	return changed, nil
}

// expand processes a token returned from the parser, returning
// the appropriate value.
//
// The expansion really means two things:
//
// 1. If the string contains variables ${foo} replace them.
//
// 2. If the token is a backtick operation then run the command
//    and return the value.
//
func (e *Executor) expand(tok token.Token) (string, error) {

	// Get the argument, and expand variables
	value := tok.Literal
	value = os.Expand(value, e.mapper)

	// If this is a backtick we replace the value
	// with the result of running the command.
	if tok.Type == token.BACKTICK {

		tmp, err := e.runCommand(value)
		if err != nil {
			return "", fmt.Errorf("error running %s: %s", value, err)
		}

		value = tmp
	}

	// Return the value we've found.
	return value, nil
}

// mapper is a helper to expand variables.
//
// ${foo} will be converted to the contents of the variable named foo
// which was created with `let foo = "bar"`, or failing that the contents
// of the environmental variable named `foo`.
//
func (e *Executor) mapper(val string) string {

	// Lookup a variable which exists?
	res, ok := e.env.Get(val)
	if ok {
		return res
	}

	// Lookup an environmental variable?
	return os.Getenv(val)
}

// runCommand returns the output of the specified command
func (e *Executor) runCommand(command string) (string, error) {

	// Are we running under a fuzzer?  If so disable this
	if os.Getenv("FUZZ") == "FUZZ" {
		return command, nil
	}

	// Build up the thing to run, using a shell so that
	// we can handle pipes/redirection.
	toRun := []string{"/bin/bash", "-c", command}

	// Run the command
	cmd := exec.Command(toRun[0], toRun[1:]...)

	// Get the output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running command '%s' %s", command, err.Error())
	}

	// Strip trailing newline.
	ret := strings.TrimSuffix(string(output), "\n")
	return ret, nil
}
