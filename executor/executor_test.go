package executor

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/skx/marionette/config"
	"github.com/skx/marionette/parser"
	"github.com/skx/marionette/rules"
)

// TestSimpleRule tests that running a simple rule succeeds
func TestSimpleRule(t *testing.T) {

	//
	// Create a temporary file, which we'll populate
	//
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		t.Fatalf("create a temporary file failed")
	}
	os.Remove(tmpfile.Name())

	//
	// Content we'll write to a file
	//
	expected := "This is a test\n"

	//
	// Setup the parameters
	//
	params := make(map[string]interface{})
	params["target"] = tmpfile.Name()
	params["content"] = expected

	//
	// Create a simple rule
	//
	r := []rules.Rule{rules.Rule{Type: "file",
		Name:      "test",
		Triggered: false,
		Params:    params}}

	//
	// Create the executor
	//
	ex := New(r)
	ex.SetConfig(&config.Config{})

	err = ex.Check()
	if err != nil {
		t.Errorf("unexpected error checking rules")
	}

	err = ex.Execute()
	if err != nil {
		t.Errorf("unexpected error running rules")
	}

	//
	// At this point our rule has run, so we should have
	// a temporary-file.
	//
	content, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		t.Errorf("failed to read rule output")
	}
	if string(content) != expected {
		t.Errorf("post-execution file has wrong content")
	}
	os.Remove(tmpfile.Name())
}

// TestRepeatedNames ensures non-unique names are detected
func TestRepeatedNames(t *testing.T) {

	//
	// Setup the parameters
	//
	params := make(map[string]interface{})

	//
	// Create a pair of rules with identical names.
	//
	r := []rules.Rule{
		rules.Rule{Type: "file",
			Name:      "test",
			Triggered: false,
			Params:    params},
		rules.Rule{Type: "file",
			Name:      "test",
			Triggered: false,
			Params:    params}}

	//
	// Create the executor
	//
	ex := New(r)
	ex.SetConfig(&config.Config{})

	err := ex.Check()
	if err == nil {
		t.Errorf("expected error checking rules, got none")
	}
	if !strings.Contains(err.Error(), "rule names must be unique") {
		t.Errorf("received an error, but not the one we expected: %s", err.Error())
	}
}

// TestBrokenDependencies ensures that we can find rules that refer to
// others that don't exist
func TestBrokenDependencies(t *testing.T) {
	//
	// Setup the parameters
	//
	params := make(map[string]interface{})
	// -> missing rule
	params["requires"] = "foo"

	//
	// Create a rule with a single dependency
	//
	r1 := []rules.Rule{rules.Rule{Type: "file",
		Name:      "test",
		Triggered: false,
		Params:    params}}

	//
	// Create a rule with a pair of dependencies
	params["requires"] = []string{"foo", "bar"}
	r2 := []rules.Rule{rules.Rule{Type: "file",
		Name:      "test",
		Triggered: false,
		Params:    params}}

	//
	// Create the executor
	//
	ex := New(r1)
	ex.SetConfig(&config.Config{})

	err := ex.Check()
	if err == nil {
		t.Errorf("expected error checking rules, got none")
	}
	if !strings.Contains(err.Error(), "has reference to") {
		t.Errorf("received an error, but not the one we expected: %s", err.Error())
	}

	//
	// Create the executor, again
	//
	ex = New(r2)
	ex.SetConfig(&config.Config{})

	err = ex.Check()
	if err == nil {
		t.Errorf("expected error checking rules, got none")
	}
	if !strings.Contains(err.Error(), "has reference to") {
		t.Errorf("received an error, but not the one we expected: %s", err.Error())
	}
}

// TestIf tests the support for our `if` conditional handling.
func TestIf(t *testing.T) {

	//
	// Setup the parameters
	//
	params := make(map[string]interface{})
	params["name"] = "foo"
	params["if"] = &parser.Condition{Name: "equals",
		Args: []string{"foo", "bar"}}

	//
	// Create our rule.
	//
	r1 := []rules.Rule{rules.Rule{Type: "file",
		Name:      "test",
		Triggered: false,
		Params:    params}}

	//
	// Create the executor
	//
	ex := New(r1)
	ex.SetConfig(&config.Config{Verbose: true})

	err := ex.Check()
	if err != nil {
		t.Errorf("unexpected error checking rules")
	}
	err = ex.Execute()
	if err != nil {
		t.Errorf("unexpected error running rules")
	}

	//
	// Now we try to run with the wrong type for our conditional
	//
	params["if"] = "foo"
	r1[0].Params = params
	ex = New(r1)
	err = ex.Execute()
	if err == nil {
		t.Errorf("expected error running rules, got none")
	}
	if !strings.Contains(err.Error(), "xpected a conditional structure") {
		t.Errorf("got an error, but not the right kind: %s", err.Error())
	}

	//
	// Finally we try to run with an unknown conditional
	//
	params["if"] = &parser.Condition{Name: "agrees",
		Args: []string{"foo", "bar"}}
	r1[0].Params = params
	ex = New(r1)
	err = ex.Execute()
	if err == nil {
		t.Errorf("expected error running rules, got none")
	}
	if !strings.Contains(err.Error(), "not available") {
		t.Errorf("got an error, but not the right kind: %s", err.Error())
	}

}

// TestUnless tests the support for our `unless` conditional handling.
func TestUnless(t *testing.T) {

	//
	// Setup the parameters
	//
	params := make(map[string]interface{})
	params["name"] = "foo"
	params["unless"] = &parser.Condition{Name: "equals",
		Args: []string{"bar", "bar"}}

	//
	// Create our rule.
	//
	r1 := []rules.Rule{rules.Rule{Type: "file",
		Name:      "test",
		Triggered: false,
		Params:    params}}

	//
	// Create the executor
	//
	ex := New(r1)
	ex.SetConfig(&config.Config{Verbose: true})

	err := ex.Check()
	if err != nil {
		t.Errorf("unexpected error checking rules")
	}
	err = ex.Execute()
	if err != nil {
		t.Errorf("unexpected error running rules")
	}

	//
	// Now we try to run with the wrong type for our conditional
	//
	params["unless"] = "foo"
	r1[0].Params = params
	ex = New(r1)
	err = ex.Execute()
	if err == nil {
		t.Errorf("expected error running rules, got none")
	}
	if !strings.Contains(err.Error(), "xpected a conditional structure") {
		t.Errorf("got an error, but not the right kind: %s", err.Error())
	}

	//
	// Finally we try to run with an unknown conditional
	//
	params["unless"] = &parser.Condition{Name: "agrees",
		Args: []string{"foo", "bar"}}
	r1[0].Params = params
	ex = New(r1)
	err = ex.Execute()
	if err == nil {
		t.Errorf("expected error running rules, got none")
	}
	if !strings.Contains(err.Error(), "not available") {
		t.Errorf("got an error, but not the right kind: %s", err.Error())
	}

}
