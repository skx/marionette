package executor

import (
	"database/sql"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/skx/marionette/ast"
	"github.com/skx/marionette/config"
	"github.com/skx/marionette/file"
	"github.com/skx/marionette/parser"
)

// TestSimpleRule tests that running a simple rule succeeds
func TestSimpleRule(t *testing.T) {

	//
	// Create a temporary file, which we'll populate
	//
	tmpfile, err := ioutil.TempFile("", "marionette-")
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
	params["target"] = ast.String{Value: tmpfile.Name()}
	params["content"] = ast.String{Value: expected}

	//
	// Create a simple rule
	//
	r := []ast.Node{
		&ast.Rule{Type: "file",
			Name:      "test",
			Triggered: false,
			Params:    params},
	}

	//
	// Create the executor
	//
	ex := New(r)
	ex.SetConfig(&config.Config{Verbose: true})

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

// TestCheckFail - calls a rule without a mandatory parameter.
func TestCheckFail(t *testing.T) {

	//
	// Setup the parameters - empty.  Bogus.
	//
	params := make(map[string]interface{})

	//
	// Create a simple rule
	//
	r := []ast.Node{
		&ast.Rule{
			Type:   "file",
			Name:   "test",
			Params: params,
		},
	}

	//
	// Create the executor
	//
	ex := New(r)
	ex.SetConfig(&config.Config{Verbose: true})

	err := ex.Check()
	if err != nil {
		t.Errorf("unexpected error checking rules")
	}

	err = ex.Execute()
	if err == nil {
		t.Errorf("expected error running rules, got none")
	}
	if !strings.Contains(err.Error(), "error validating") {
		t.Errorf("got an error, but the wrong one: %s", err.Error())
	}

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
	r := []ast.Node{
		&ast.Rule{Type: "file",
			Name:      "test",
			Triggered: false,
			Params:    params},
		&ast.Rule{Type: "file",
			Name:      "test",
			Triggered: false,
			Params:    params},
	}

	//
	// Create the executor
	//
	ex := New(r)
	ex.SetConfig(&config.Config{Verbose: true})

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
	params["require"] = ast.String{Value: "foo"}

	//
	// Create a rule with a single dependency
	//
	r1 := []ast.Node{
		&ast.Rule{Type: "file",
			Name:      "test",
			Triggered: false,
			Params:    params},
	}

	//
	// Create a rule with a pair of dependencies
	params["require"] = ast.Array{
		Values: []ast.Object{
			ast.String{Value: "foo"},
			ast.String{Value: "bar"},
		},
	}

	r2 := []ast.Node{
		&ast.Rule{Type: "file",
			Name:      "test",
			Triggered: false,
			Params:    params},
	}

	//
	// Create the executor
	//
	ex := New(r1)
	ex.SetConfig(&config.Config{Verbose: true})

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
	ex.SetConfig(&config.Config{Verbose: true})

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

	//
	// Create our rule.
	//
	r1 := []ast.Node{
		&ast.Rule{
			Type:          "file",
			Name:          "test",
			Triggered:     false,
			Params:        params,
			ConditionType: "if",
			Function: ast.Funcall{
				Name: "equal",
				Args: []ast.Object{
					ast.String{Value: "foo"},
					ast.String{Value: "bar"},
				},
			},
		},
	}

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
		t.Errorf("unexpected error running rules: %s", err.Error())
	}

	//
	// Now we try to run with the wrong type for our conditional
	//
	// change params
	tmp := r1[0].(*ast.Rule)
	tmp.ConditionType = "foo"

	ex = New(r1)
	err = ex.Execute()
	if err == nil {
		t.Errorf("expected error running rules, got none")
	}
	if !strings.Contains(err.Error(), "unknown condition-type") {
		t.Errorf("got an error, but not the right kind: %s", err.Error())
	}

	//
	// Finally we try to run with an unknown conditional
	//
	// change params
	tmpt := r1[0].(*ast.Rule)
	tmpt.Params = params
	tmpt.ConditionType = "if"
	tmpt.Function = ast.Funcall{
		Name: "agrees",
		Args: []ast.Object{
			ast.String{Value: "foo"},
			ast.String{Value: "bar"},
		},
	}

	ex = New(r1)
	err = ex.Execute()
	if err == nil {
		t.Errorf("expected error running rules, got none")
	}
	if !strings.Contains(err.Error(), "not defined") {
		t.Errorf("got an error, but not the right kind: %s", err.Error())
	}

}

// TestTriggered uses a rule which is "triggered", and thus shouldn't be
// executed normally.
func TestTriggered(t *testing.T) {

	//
	// Create our rule.
	//
	r1 := []ast.Node{
		&ast.Rule{Type: "file",
			Name:          "bob",
			Triggered:     false,
			ConditionType: "if",
			Function: ast.Funcall{
				Name: "equal",
				Args: []ast.Object{
					ast.String{Value: "foo"},
					ast.String{Value: "bar"},
				},
			},
			Params: map[string]interface{}{
				"require": ast.String{Value: "test"},
			},
		},
		&ast.Rule{Type: "file",
			Name:      "test",
			Triggered: true,
			Params: map[string]interface{}{
				"target":  "/tmp/foo",
				"ensure":  "present",
				"content": "foo",
			},
		},
	}

	//
	// Create the executor
	//
	ex := New(r1)
	ex.SetConfig(&config.Config{Verbose: true})

	err := ex.Check()
	if err != nil {
		t.Errorf("unexpected error checking rules: %s", err.Error())
	}
	err = ex.Execute()
	if err != nil {
		t.Errorf("unexpected error running rules: %s", err.Error())
	}

}

// TestUnless tests the support for our `unless` conditional handling.
func TestUnless(t *testing.T) {

	//
	// Setup the parameters
	//
	params := make(map[string]interface{})
	params["name"] = "foo"
	params["target"] = "foo"
	params["source_url"] = "https://example.com/"

	//
	// Create our rule.
	//
	r1 := []ast.Node{
		&ast.Rule{
			Type:          "file",
			Name:          "test",
			Triggered:     false,
			Params:        params,
			ConditionType: "unless",
			Function: ast.Funcall{
				Name: "equal",
				Args: []ast.Object{
					ast.String{Value: "bar"},
					ast.String{Value: "bar"},
				},
			},
		},
	}

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
	// change params
	tmp := r1[0].(*ast.Rule)
	tmp.ConditionType = "foo"

	ex = New(r1)
	err = ex.Execute()
	if err == nil {
		t.Errorf("expected error running rules, got none")
	}
	if !strings.Contains(err.Error(), "unknown condition-type") {
		t.Errorf("got an error, but not the right kind: %s", err.Error())
	}

	//
	// Finally we try to run with an unknown conditional
	//
	// change params
	tmpt := r1[0].(*ast.Rule)
	tmpt.Function = ast.Funcall{
		Name: "tervetulo",
		Args: []ast.Object{
			ast.String{Value: "foo"},
			ast.String{Value: "bar"},
		},
	}

	ex = New(r1)
	err = ex.Execute()
	if err == nil {
		t.Errorf("expected error running rules, got none")
	}
	if !strings.Contains(err.Error(), "not defined") {
		t.Errorf("got an error, but not the right kind: %s", err.Error())
	}

}

// Test a moderately complex program
func TestModerateExample(t *testing.T) {

	// Write some content to a file
	trash, err := WriteContent("this will get deleted")
	if err != nil {
		t.Fatalf("failed to write file to be deleted")
	}

	// Write out a rule to delete that file
	del, err := WriteContent("file { target => \"" + trash + "\", state => \"absent\", require => \"logger\" }   log { name => \"logger\", message => \"hello\" }")
	if err != nil {
		t.Fatalf("failed to write file-rule")
	}

	// Now write out a proper file
	main, err := WriteContent(`let a = "1" if equal("one", "one");` +
		`include "` + del + `" if equal("one", "one");` +
		`include "` + del + `" if equal("one", "one");`)

	if err != nil {
		t.Fatalf("failed to write main file")
	}

	// Run the whole thing.
	data, err := ioutil.ReadFile(main)
	if err != nil {
		t.Fatalf("failed to read main file")
	}
	// Create a new parser with our file content.
	p := parser.New(string(data))

	// Parse the rules
	out, err := p.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %s", err)
	}

	// Execute
	ex := New(out.Recipe)

	// Check for broken dependencies
	err = ex.Check()
	if err != nil {
		t.Fatalf("failed to check rules:%s", err)
	}

	// Now execute!
	err = ex.Execute()
	if err != nil {
		t.Fatalf("failed to run rules:%s", err)
	}

	// Now we've run the temporary file we created at the start
	// should have been removed.
	if file.Exists(trash) {
		t.Fatalf("file should have been removed, but wasn't!")
	}

	os.Remove(main)
	os.Remove(del)
	// os.Remove(trash) - removed already :)

}

func WriteContent(input string) (string, error) {

	tmpfile, err := ioutil.TempFile("", "marionette-")
	if err != nil {
		return "", err
	}

	d1 := []byte(input)
	err = os.WriteFile(tmpfile.Name(), d1, 0644)
	if err != nil {
		return "", err
	}
	return tmpfile.Name(), nil

}

// TestSQLRun executes a marionette file, and will confirm it creates
// what we expect
func TestSQLRun(t *testing.T) {

	// Create a temporary file-name
	tmpfile, err := ioutil.TempFile("", "marionette-")
	if err != nil {
		t.Fatalf("create a temporary file failed")
	}
	os.Remove(tmpfile.Name())

	// Ensure that the file doesn't exist
	if file.Exists(tmpfile.Name()) {
		t.Fatalf("the file we removed is present still?")
	}

	//
	// Module source we're gonna execute.
	//
	src := `
sql {
     driver   => "sqlite3",
     dsn      => "file:#PATH#",
     sql      => "

CREATE TABLE IF NOT EXISTS contacts (
        contact_id INTEGER PRIMARY KEY,
        first_name TEXT NOT NULL,
        last_name TEXT NOT NULL,
        email TEXT NOT NULL
);

INSERT INTO contacts( first_name, last_name, email ) VALUES( 'steve', 'kemp', 'steve@steve.fi');
INSERT INTO contacts( first_name, last_name, email ) VALUES( 'nobody', 'special', 'steve@example.com');
",
}
`

	// FILE -> the temporary filename
	src = strings.ReplaceAll(src, "#PATH#", tmpfile.Name())

	// Create a new parser with our content.
	p := parser.New(string(src))

	// Parse the rules
	out, err := p.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %s", err)
	}

	// Execute
	ex := New(out.Recipe)

	// Check for broken dependencies
	err = ex.Check()
	if err != nil {
		t.Fatalf("failed to check rules:%s", err)
	}

	// Now execute!
	err = ex.Execute()
	if err != nil {
		t.Fatalf("failed to run rules:%s", err)
	}

	// Ensure that we now have a generated SQLite file
	if !file.Exists(tmpfile.Name()) {
		t.Fatalf("we expected SQLite file to be created")
	}

	// Right now open and find the contents.
	db, err := sql.Open("sqlite3", tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to open sqlite3 file")
	}

	row, err := db.Query("SELECT * FROM contacts WHERE email NOT LIKE '%exampl%'")
	if err != nil {
		t.Fatalf("failed to prepare SQL query")
	}
	defer row.Close()

	for row.Next() {

		var id string
		var first string
		var last string
		var mail string

		err := row.Scan(&id, &first, &last, &mail)
		if err != nil {
			t.Fatalf("error running row-scan")
		}

		if first != "steve" {
			t.Fatalf("unexpected SQL result:%s", first)
		}
		if last != "kemp" {
			t.Fatalf("unexpected SQL result:%s", last)
		}
		if mail != "steve@steve.fi" {
			t.Fatalf("unexpected SQL result:%s", mail)
		}
	}

	db.Close()
	os.Remove(tmpfile.Name())

}

// Ensure triggered rules are ignored
func TestIgnoreTriggered(t *testing.T) {

	src := `
# only created if notified
file triggered {
      name    => "one",
      target  => "one.tst",
      content => "OK"
}

# only created if notified
file triggered {
      name    => "two",
      target  => "two.tst",
      content => "OK"
}`

	// Before we begin neither file will exist.
	for _, f := range []string{"one.tst", "two.tst"} {

		// Ensure that we now have a generated SQLite file
		if file.Exists(f) {
			t.Fatalf("we did not expect the file to be present: %s", f)
		}
	}

	// Create a new parser with our content.
	p := parser.New(string(src))

	// Parse the rules
	out, err := p.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %s", err)
	}

	// Execute
	ex := New(out.Recipe)

	// Check for broken dependencies
	err = ex.Check()
	if err != nil {
		t.Fatalf("failed to check rules:%s", err)
	}

	// Now execute!
	err = ex.Execute()
	if err != nil {
		t.Fatalf("failed to run rules:%s", err)
	}

	// At this point those files should still not exist.
	for _, f := range []string{"one.tst", "two.tst"} {

		// Ensure that we now have a generated SQLite file
		if file.Exists(f) {
			t.Fatalf("file should not be created: %s", f)
		}

		os.Remove(f)
	}
}

// Notify two rules upon a change.
func TestNotifyMultiple(t *testing.T) {

	src := `
# always results in a change
log { message => "test",
      notify  => [ "one", "two" ],
}

# only created if notified
file triggered {
      name    => "one",
      target  => "one.tst",
      content => "OK"
}

# only created if notified
file triggered {
      name    => "two",
      target  => "two.tst",
      content => "OK"
}`

	// Before we begin neither file will exist.
	for _, f := range []string{"one.tst", "two.tst"} {

		// Ensure that we now have a generated SQLite file
		if file.Exists(f) {
			t.Fatalf("we did not expect the file to be present: %s", f)
		}
	}

	// Create a new parser with our content.
	p := parser.New(string(src))

	// Parse the rules
	out, err := p.Parse()
	if err != nil {
		t.Fatalf("failed to parse: %s", err)
	}

	// Execute
	ex := New(out.Recipe)

	// Check for broken dependencies
	err = ex.Check()
	if err != nil {
		t.Fatalf("failed to check rules:%s", err)
	}

	// Now execute!
	err = ex.Execute()
	if err != nil {
		t.Fatalf("failed to run rules:%s", err)
	}

	// At this point we should have two files created
	for _, f := range []string{"one.tst", "two.tst"} {

		// Ensure that we now have a generated SQLite file
		if !file.Exists(f) {
			t.Fatalf("we expected a file to be created: %s", f)
		}

		os.Remove(f)
	}
}
