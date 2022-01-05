// Package executor is the thing that will execute our rules.
//
// This means processing the rules, one by one, but also ensuring
// dependencies are handled.
//
// Variable assignments, and include-file inclusion, occur at
// run-time too, so they are handled here.
package executor

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/skx/marionette/ast"
	"github.com/skx/marionette/conditionals"
	"github.com/skx/marionette/config"
	"github.com/skx/marionette/environment"
	"github.com/skx/marionette/modules"
	"github.com/skx/marionette/parser"
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

	// included keeps track of which files we've already included.
	//
	// We use this to avoid issues with recursive file inclusions
	included map[string]bool

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
		cfg:      &config.Config{},
		env:      environment.New(),
		Program:  program,
		included: make(map[string]bool),
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
		switch r := r.(type) {

		case *ast.Assign:

			e.verbose(fmt.Sprintf("Processing assignment: %v\n", r))

			// variable assignment
			err := e.executeAssign(r)
			if err != nil {
				return err
			}

		case *ast.Include:

			e.verbose(fmt.Sprintf("Processing inclusion: %v\n", r))

			// include-file handling
			err := e.executeInclude(r)
			if err != nil {
				return err
			}

		case *ast.Rule:
			// rule execution

			e.verbose(fmt.Sprintf("Processing rule: %v\n", r))

			// Don't run rules that are only present to
			// be notified by a trigger.
			if r.Triggered {
				continue
			}

			// Have we executed this rule already?
			if seen[i] {
				continue
			}

			// Get the rule dependencies.
			deps := e.deps(r, "require")

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
			err := e.executeSingleRule(r)
			if err != nil {
				return err
			}

			// And mark this as executed too.
			seen[i] = true
		default:
			return fmt.Errorf("unknown node type! %t", r)
		}
	}
	return nil
}

// executeAssign executes an assignment node, updating the environment.
func (e *Executor) executeAssign(assign *ast.Assign) error {

	key := assign.Key
	val := assign.Value
	ret := ""
	var err error

	e.verbose(fmt.Sprintf("Setting variable %s -> %s", key, val))

	switch val.Type {
	case token.STRING:
		ret = os.Expand(val.Literal, e.mapper)
	case token.BACKTICK:
		ret, err = e.expand(val)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled type in executeAssign %v", val)
	}

	e.env.Set(key, ret)
	return nil
}

// executeInclude will handle a file inclusion node.
func (e *Executor) executeInclude(inc *ast.Include) error {

	// OK is this conditionally included?
	if inc.ConditionType != "" {

		cond := inc.ConditionType

		if cond == "if" {
			res, err := e.runConditional(inc.ConditionRule)
			if err != nil {
				return err
			}
			if !res {
				e.verbose(fmt.Sprintf("\tSkipping inclusion of %s condition was not true: %s", inc.Source, inc.ConditionRule))
				return nil
			}
		}

		if cond == "unless" {
			res, err := e.runConditional(inc.ConditionRule)
			if err != nil {
				return err
			}
			if res {
				e.verbose(fmt.Sprintf("\tSkipping inclusion of %s condition was not false: %s", inc.Source, inc.ConditionRule))
				return nil
			}
		}

	}

	// Expand any variables in the string.
	inc.Source = os.Expand(inc.Source, e.mapper)

	// If we've already included this path, return
	seen, ok := e.included[inc.Source]
	if ok && seen {
		e.verbose(fmt.Sprintf("Skipping include file %s - already seen\n",
			inc.Source))
		return nil
	}

	// Mark it as included now.
	e.included[inc.Source] = true

	// Now run the inclusion
	data, err := ioutil.ReadFile(inc.Source)
	if err != nil {
		return fmt.Errorf("failed to read include-source %s: %s", inc.Source, err)
	}

	// Create a new parser with our file content.
	p := parser.New(string(data))

	// Parse the rules
	out, err := p.Parse()
	if err != nil {
		return err
	}

	// Create the new executor
	ex := New(out.Recipe)

	// Set the configuration options.
	ex.SetConfig(e.cfg)

	// Propagate all the variables which we have in-scope.
	for k, v := range e.env.Variables() {
		ex.env.Set(k, v)
	}

	// Propagate all the include-files that have been seen
	for k, v := range e.included {
		ex.included[k] = v
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

	// Once the child executor has finished we'll copy back
	// the files that it has seen as included.
	// Propagate all the include-files that have been seen
	for k, v := range ex.included {
		e.included[k] = v
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

	// We want to ensure that we expand all arguments
	args := []string{}

	for _, arg := range test.Args {
		args = append(args, os.Expand(arg, e.mapper))
	}

	// Call the function, and return whatever result it gives us.
	return helper(args)
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
