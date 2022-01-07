package modules

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/skx/marionette/environment"
)

// Test the argument validation works
func TestLogArguments(t *testing.T) {

	e := environment.New()

	// Save our log writer
	before := log.Writer()
	defer log.SetOutput(before)

	// Chang logger to write to a temporary buffer.
	var buf bytes.Buffer
	log.SetOutput(&buf)

	// Create the object, and arguments.
	l := &LogModule{}
	args := make(map[string]interface{})
	empty := make(map[string]interface{})

	// Missing argument
	err := l.Check(args)
	if err == nil {
		t.Fatalf("expected error due to missing message-parameter")
	}
	if !strings.Contains(err.Error(), "missing 'message'") {
		t.Fatalf("got error - but wrong one : %s", err)
	}

	// Wrong kind of argument
	args["message"] = 3
	err = l.Check(args)
	if err == nil {
		t.Fatalf("expected error due to missing message-parameter")
	}
	if !strings.Contains(err.Error(), "failed to convert") {
		t.Fatalf("got error - but wrong one : %s", err)
	}

	// Valid argument
	args["message"] = "Hello, world"
	err = l.Check(args)
	if err != nil {
		t.Fatalf("unexpected error checking")
	}

	//
	// Try to execute - missing argument
	//
	c := false
	c, err = l.Execute(e, empty)
	if err == nil {
		t.Fatalf("expected an error with no message.")
	}

	//
	// Try to execute - valid argument
	//
	c, err = l.Execute(e, args)
	if err != nil {
		t.Fatalf("unexpected error executing")
	}
	if !c {
		t.Fatalf("logger should have resulted in a change")
	}

	//
	// Confirm we got our message in the log-output
	//
	output := buf.String()
	if !strings.Contains(output, "Hello, world") {
		t.Fatalf("log message wasn't found")
	}

}
