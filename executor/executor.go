// Package executor is the thing that will execute our rules.
//
// This means processing the rules, one by one, but also ensuring
// dependencies are handled.
//
// Variable assignments, and the inclusion of files, occur at
// run-time too, so they are handled here.
package executor

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/skx/marionette/ast"
	"github.com/skx/marionette/config"
	"github.com/skx/marionette/environment"
	"github.com/skx/marionette/modules"
	"github.com/skx/marionette/parser"
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

	// Keep track of which rules we've executed.
	executed map[string]bool

	// included keeps track of which files we've already included.
	//
	// We use this to avoid issues with recursive file inclusions.
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
		executed: make(map[string]bool),
		index:    make(map[string]int),
	}

	return e
}

// SetConfig updates the executor with the specified configuration object.
func (e *Executor) SetConfig(cfg *config.Config) {
	e.cfg = cfg
}

// MarkSeen marks the given file as having already been seen.
func (e *Executor) MarkSeen(path string) {
	e.included[path] = true
}

// SetMagicIncludeVars sets magic environment variables.
func (e *Executor) SetMagicIncludeVars(path string) error {
	abspath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	e.env.Set("INCLUDE_FILE", abspath)
	log.Printf("[DEBUG] Set include path variable INCLUDE_FILE -> %s\n", abspath)

	dirpath := filepath.Dir(abspath)
	e.env.Set("INCLUDE_DIR", dirpath)
	log.Printf("[DEBUG] Set include path variable INCLUDE_DIR -> %s\n", dirpath)

	return nil
}

// Get the rules a rule depends upon, via the given key.
//
// This is used to find any `require` or `notify` rules.
func (e *Executor) deps(rule *ast.Rule, key string) ([]string, error) {

	var res []string

	// Get the value from the map, if it exists.
	requires, ok := rule.Params[key]

	// no requirements/dependencies?  Then we're done.
	if !ok {
		return res, nil
	}

	//
	// OK the requirements might be a single object, or
	// an array of objects
	//
	// Handle both cases.
	//
	// Is this an array-object?
	array, ok := requires.(ast.Array)
	if ok {
		// For each of the children
		for _, tmp := range array.Values {

			// Evaluate it, and store
			val, err := tmp.Evaluate(e.env)
			if err != nil {
				return res, err
			}
			res = append(res, val)
		}
		return res, nil
	}

	// Is this a single object?
	dep, ok2 := requires.(ast.Object)
	if ok2 {

		// Is it a single node, which we can convert?
		val, err := dep.Evaluate(e.env)
		if err != nil {
			return res, err
		}
		res = append(res, val)
		return res, nil
	}

	return nil, fmt.Errorf("unknown object at deps - %v %t", requires, requires)
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
		deps, dErr := e.deps(rule, "require")
		if dErr != nil {
			return dErr
		}

		notify, nErr := e.deps(rule, "notify")
		if nErr != nil {
			return nErr
		}

		// Log these.
		log.Printf("[DEBUG] Rule %s require:[%s] notify:[%s]\n",
			rule.Name,
			strings.Join(deps, ","),
			strings.Join(notify, ","))

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

	// For each node in our program
	for _, r := range e.Program {

		// Test the type to see what we should do.
		switch r := r.(type) {

		case *ast.Assign:

			log.Printf("[DEBUG] Processing assignment: %s", r)

			// variable assignment
			err := e.executeAssign(r)
			if err != nil {
				return err
			}

		case *ast.Include:

			log.Printf("[DEBUG] Processing inclusion: %v\n", r)

			// include-file handling
			err := e.executeInclude(r)
			if err != nil {
				return err
			}

		case *ast.Rule:

			log.Printf("[DEBUG] Processing rule: %s", r)

			// rule execution
			err := e.executeSingleRule(r, false)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("unknown node type! %t", r)
		}
	}
	return nil
}

// executeAssign executes an assignment node, updating the environment.
func (e *Executor) executeAssign(assign *ast.Assign) error {

	// OK is this conditionally assigned?
	if assign.ConditionType != "" {

		// Should we execute the assignment?
		ret, err := e.shouldExecute(assign.ConditionType, assign.Function)

		// Error?  Then return that
		if err != nil {
			return err
		}

		// If we didn't get a "true" then we should skip this action.
		if !ret {
			return nil
		}
	}

	// The key
	key := assign.Key

	// Execute the literal object (be it a number, string, backtick or bool)
	val, err := assign.Value.Evaluate(e.env)
	if err != nil {
		return err
	}

	// Show what we're going to do.
	log.Printf("[DEBUG] Set '%s' -> '%s'", key, val)

	// Set the value
	e.env.Set(key, val)
	return nil
}

// executeInclude will handle a file inclusion node.
func (e *Executor) executeInclude(inc *ast.Include) error {

	// OK is this conditionally assigned?
	if inc.ConditionType != "" {

		// Should we execute the inclusion?
		ret, err := e.shouldExecute(inc.ConditionType, inc.Function)

		// Error?  Then return that
		if err != nil {
			return err
		}

		// If we didn't get a "true" then we should skip this action.
		if !ret {
			return nil
		}
	}

	// We now need to handle the things that we should include
	//
	// We might have:
	//
	//   include "path/to/file"
	//   include true
	//   include [ "one.txt", "two.txt" ]
	//
	// Because the array value will handle multiple values we'll
	// expand them as we go.
	includes := []string{}

	//
	// Is this an array?
	//
	array, ok := inc.Source.(ast.Array)
	if ok {

		// If so evaluate each node and save it
		// in our list of things to include.
		for _, p := range array.Values {

			val, err2 := p.Evaluate(e.env)
			if err2 != nil {
				return err2
			}

			// save into our array of strings
			includes = append(includes, val)
		}

	} else {

		// OK this isn't an array, so we can just
		// handle it as a single-thing.
		val, err2 := inc.Source.Evaluate(e.env)
		if err2 != nil {
			return err2
		}

		includes = append(includes, val)
	}

	// For each thing to include ..
	for _, path := range includes {

		// If we've already included this path, skip it
		seen, ok := e.included[path]
		if ok && seen {
			log.Printf("[INFO] Skipping inclusion of %s - already seen", path)
			continue
		}

		// Mark it as included now.
		e.MarkSeen(path)

		// And read/run it.
		err := e.executeIncludeReal(path)
		if err != nil {
			return fmt.Errorf("failed to execute included file %s: %s", path, err)
		}
	}

	return nil
}

// executeIncludeReal handles the mechanics of launching a sub-executor,
// setting up the include-file history & etc.
func (e *Executor) executeIncludeReal(source string) error {

	// Read the source we're to include
	data, err := ioutil.ReadFile(source)
	if err != nil {
		return fmt.Errorf("failed to read include-source %s: %s", source, err)
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

	// Set "magic" variables for the current include file.
	err = ex.SetMagicIncludeVars(source)
	if err != nil {
		return err
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
	for k, v := range ex.included {
		e.included[k] = v
	}

	return nil
}

// shouldExecute tests whether the assignment/include/rule should be executed,
// based on the condition-type and the condition-rule.
func (e *Executor) shouldExecute(cType string, cRule ast.Funcall) (bool, error) {

	// Invoke it, and get the output
	ret, err := cRule.Evaluate(e.env)
	if err != nil {
		return false, err
	}

	// Function return-value
	retVal := true

	// Is the result "truthy"?
	if ret == "" ||
		ret == "false" ||
		ret == "0" {
		retVal = false
	}

	// Now see if this means the thing should execute
	switch cType {
	case "if":
		if !retVal {
			log.Printf("[INFO] Skipping because condition was not true: %s", cRule)
			return false, nil
		}
	case "unless":
		if retVal {
			log.Printf("[INFO] Skipping because condition was not false: %s", cRule)
			return false, nil
		}
	default:
		return false, fmt.Errorf("unknown condition-type %s", cType)
	}

	// OK the assignment/include/rule should be executed
	return true, nil
}

// executeSingleRule creates the appropriate module, and runs the single rule.
func (e *Executor) executeSingleRule(rule *ast.Rule, force bool) error {

	// Show what we're doing
	log.Printf("[INFO] Running %s-module rule: %s", rule.Type, rule.Name)

	// Don't run rules that are only present to
	// be notified by a trigger.
	if rule.Triggered {
		if force {
			log.Printf("[DEBUG] Forcing execution of rule due to notify action")
		} else {
			log.Printf("[DEBUG] Skipping rule because it has the triggered-modifier")
			return nil
		}
	}

	// Have we executed this rule already?
	if e.executed[rule.Name] {
		log.Printf("[DEBUG] Skipping rule because it has already executed")
		return nil
	}

	e.executed[rule.Name] = true

	// Get the rule dependencies.
	deps, dErr := e.deps(rule, "require")
	if dErr != nil {
		return dErr
	}

	// Process each one
	for _, dep := range deps {

		dr := e.Program[e.index[dep]].(*ast.Rule)
		log.Printf("[DEBUG] Running dependency for %s: %s\n", rule.Name, dr.Name)
		// Now the rule itself
		err := e.executeSingleRule(dr, false)
		if err != nil {
			return err
		}

	}

	// OK is this conditionally executed?
	if rule.ConditionType != "" {

		// Should we execute the rule?
		ret, err := e.shouldExecute(rule.ConditionType, rule.Function)

		// Error?  Then return that
		if err != nil {
			return err
		}

		// If we didn't get a "true" then we should skip this action.
		if !ret {
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
	helper := modules.Lookup(rule.Type, e.cfg, e.env)
	if helper == nil {
		return fmt.Errorf("unknown module type %s, from rule %v", rule.Type, rule)
	}

	// Run the module instance
	changed, err = e.runInternalModule(helper, rule)
	if err != nil {
		return err
	}

	if changed {

		log.Printf("[INFO] Rule resulted in a change being made.")

		// Now call any rules that we should notify.
		notify, nErr := e.deps(rule, "notify")
		if nErr != nil {
			return nErr
		}

		// Process each one
		for _, child := range notify {

			// get the actual rule, by index
			dr := e.Program[e.index[child]].(*ast.Rule)

			// Show what we're going to do.
			log.Printf("[INFO] Notifying rule: %s", dr.Name)

			// Execute the rule.
			err := e.executeSingleRule(dr, true)
			if err != nil {
				return err
			}
		}
	} else {
		log.Printf("[INFO] Rule resulted in no change being made.")
	}

	// All done
	return nil
}

// runInternalModule executes the given rule with the loaded internal module.
func (e *Executor) runInternalModule(helper modules.ModuleAPI, rule *ast.Rule) (bool, error) {

	var err error

	// Expand all params into strings/arrays of strings
	// into a new map.  We leave the rule-params alone.
	params := make(map[string]interface{})

	// So for each argument
	for k, v := range rule.Params {

		// Is this parameter value an array?
		//
		// If so expand each value it contains.
		array, ok := v.(ast.Array)
		if ok {

			// temporary values
			var tmp []string

			// for each node
			for _, p := range array.Values {

				val, err2 := p.Evaluate(e.env)
				if err2 != nil {
					return false, err2
				}

				// save into our array of strings
				tmp = append(tmp, val)
			}

			params[k] = tmp

			continue
		}

		// parameter contains a single node?
		p, ok := v.(ast.Object)
		if ok {

			// Is it a single node, which we can convert?
			val, err2 := p.Evaluate(e.env)
			if err2 != nil {
				return false, err2
			}
			params[k] = val

			continue
		}

		// We got a parameter which is unknown
		return false, fmt.Errorf("runInternalModule unknown object at deps - %V %T", v, v)

	}

	// Check the arguments, using the module-specific Check method.
	err = helper.Check(params)
	if err != nil {
		return false, fmt.Errorf("error validating %s-module rule '%s' %s",
			rule.Type, rule.Name, err.Error())
	}

	// Execute the module.
	changed, err := helper.Execute(params)
	if err != nil {
		return false, fmt.Errorf("error running %s-module rule '%s' %s",
			rule.Type, rule.Name, err.Error())
	}

	// Now that execution is complete it might be that the module
	// wishes to store variables in the environment.
	//
	// If the module implements the "ModuleOutput" interface then invoke
	// it, and update the environment appropriately.
	if outputs, ok := helper.(modules.ModuleOutput); ok {

		// Logging.
		log.Printf("[DEBUG] Rule %s [%s] implements ModuleOutput\n",
			rule.Name, rule.Type)

		// Get any output-variables which should be set.
		out := outputs.GetOutputs()

		// For each one then set variable
		for key, val := range out {

			//
			// NOTE: We set the variable scoped by the
			//       rule name.
			//
			name := rule.Name + "." + key

			log.Printf("[DEBUG] SetOutputVariable (%s) => %s\n",
				name, val)
			e.env.Set(name, val)
		}
	}

	// If the module resulted in a change record that too.
	//
	// Note this doesn't require the module to implement the
	// ModuleOutput interface - it is global, for all rule-types
	key := rule.Name + ".changed"
	val := "false"
	if changed {
		val = "true"
	}
	e.env.Set(key, val)

	// Finally return the value to the caller.
	return changed, nil
}
