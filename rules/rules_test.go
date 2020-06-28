package rules

import (
	"fmt"
	"strings"
	"testing"
)

// Ensure a rule doesn't have the "triggered" value by default.
func TestTriggered(t *testing.T) {

	// create a random rule and format as string
	r := NewRule("shell", "rule-name", nil)
	out := r.String()

	if r.Triggered {
		t.Fatalf("rule shouldn't have a 'triggered' marker")
	}
	if strings.Contains(out, " triggered ") {
		t.Fatalf("stringified rule shouldn't have a 'triggered' marker")
	}

	// Now make it triggered
	r.Triggered = true
	out = r.String()
	if !r.Triggered {
		t.Fatalf("rule should have a 'triggered' marker now")
	}
	if !strings.Contains(out, " triggered ") {
		t.Fatalf("stringified rule should have a 'triggered' marker now")
	}

}

// TestSimpleRule ensures the string-formatted version of a rule has
// expected content.
func TestSimpleRule(t *testing.T) {

	params := make(map[string]interface{})
	params["foo"] = "bar"
	params["name"] = "steve"

	// create a new rule and format as string
	r := NewRule("shell", "my-rule-name", params)
	out := r.String()

	// We expect `name => "Steve"` in our output, etc.
	for k, v := range params {

		if !strings.Contains(out, k) {
			t.Fatalf("didn't find expected key %s in output", k)
		}
		if !strings.Contains(out, fmt.Sprintf("\"%s\"", v)) {
			t.Fatalf("didn't find expected value %s in output", k)
		}
	}
}

// TestArrayRule ensures the string-formatted version of a rule has
// expected content.
func TestArrayRule(t *testing.T) {

	params := make(map[string]interface{})
	params["children"] = []string{"bart", "lisa", "maggie"}

	// create a new rule and format as string
	r := NewRule("shell", "my-rule-name", params)
	out := r.String()

	// We expect `name => "Steve"` in our output, etc.
	for k, v := range params {

		if !strings.Contains(out, k) {
			t.Fatalf("didn't find expected key %s in output", k)
		}

		values, ok := v.([]string)
		if !ok {
			t.Fatalf("test value wasn't a string array!")
		}

		for _, x := range values {
			if !strings.Contains(out, fmt.Sprintf("\"%s\"", x)) {
				t.Fatalf("didn't find expected value %s in output", x)
			}
		}
	}
}
