package ast

import (
	"strings"
	"testing"

	"github.com/skx/marionette/environment"
)

// TestBrokenBacktick tests the basic backtick functions fail
func TestBrokenBacktick(t *testing.T) {

	e := environment.New()

	b := &Backtick{Value: "/no/such/binary-can-be/here"}

	_, err := b.Evaluate(e)
	if err == nil {
		t.Fatalf("expected error running missing binary")
	}
}

// TestSimpleFunction tests some simple functions.
func TestSimpleFunction(t *testing.T) {

	// Test calling a function that doesn't exist
	f := &Funcall{Name: "not-found.bogus"}
	_, err := f.Evaluate(nil)
	if err == nil {
		t.Fatalf("expected error calling missing function")
	}

	// Test calling a function with the wrong number of arguments
	f.Name = "matches"
	f.Args = []Object{
		&String{Value: "haystack"},
	}
	_, err = f.Evaluate(nil)
	if err == nil {
		t.Fatalf("expected error calling function with wrong args")
	}
	if !strings.Contains(err.Error(), "wrong number of args") {
		t.Fatalf("got error, but wrong one:%s", err.Error())
	}

	// Test calling a function with an arg that will fail
	f.Name = "matches"
	f.Args = []Object{
		&Backtick{Value: "`/f/not-found"},
		&Backtick{Value: "`/f/not-found"},
	}
	_, err = f.Evaluate(nil)
	if err == nil {
		t.Fatalf("expected error calling function with broken arg")
	}
	if !strings.Contains(err.Error(), "error running command") {
		t.Fatalf("got error, but wrong one:%s", err.Error())
	}

	// Test calling a function with no error
	f.Name = "len"
	f.Args = []Object{
		&String{Value: "Hello, World"},
	}
	out, err2 := f.Evaluate(nil)
	if err2 != nil {
		t.Fatalf("unexpected error calling 'len'")
	}

	if out != "12" {
		t.Fatalf("unexpected result for len(Hello, World) : %s", out)
	}
}

func TestStringification(t *testing.T) {

	// Backtick
	b := &Backtick{Value: "/usr/bin/id"}
	if !strings.Contains(b.String(), "Backtick") {
		t.Fatalf("stringified object is bogus")
	}
	if !strings.Contains(b.String(), "/usr/bin/id") {
		t.Fatalf("stringified object is bogus")
	}

	// Boolean
	bo := &Boolean{Value: true}
	if !strings.Contains(bo.String(), "Boolean") {
		t.Fatalf("stringified object is bogus")
	}
	if !strings.Contains(bo.String(), "t") {
		t.Fatalf("stringified object is bogus")
	}

	// Boolean: Evaluate - true
	boe, berr := bo.Evaluate(nil)
	if berr != nil {
		t.Fatalf("unexpected error evaluating object:%s", berr.Error())
	}
	if boe != "true" {
		t.Fatalf("wrong value evaluating bool:%s", boe)
	}

	// Boolean: Evaluate - false
	bo = &Boolean{Value: false}
	boe, berr = bo.Evaluate(nil)
	if berr != nil {
		t.Fatalf("unexpected error evaluating object:%s", berr.Error())
	}
	if boe != "false" {
		t.Fatalf("wrong value evaluating bool:%s", boe)
	}

	// Funcall
	f := &Funcall{Name: "equal", Args: []Object{
		&String{Value: "one"},
		&String{Value: "two"},
	}}
	if !strings.Contains(f.String(), "Funcall") {
		t.Fatalf("stringified object is bogus")
	}
	if !strings.Contains(f.String(), "equal") {
		t.Fatalf("stringified object is bogus")
	}
	if !strings.Contains(f.String(), "String{one},String{two}") {
		t.Fatalf("stringified object is bogus")
	}

	// Number
	n := &Number{Value: 323}
	if !strings.Contains(n.String(), "Number") {
		t.Fatalf("stringified object is bogus")
	}
	if !strings.Contains(n.String(), "323") {
		t.Fatalf("stringified object is bogus")
	}

	// Number: Evaluate
	no, err := n.Evaluate(nil)
	if err != nil {
		t.Fatalf("unexpected error evaluating object:%s", err.Error())
	}
	if no != "323" {
		t.Fatalf("wrong value evaluating number:%s", no)
	}
	// String
	tmp := &String{Value: "steve"}
	if !strings.Contains(tmp.String(), "steve") {
		t.Fatalf("stringified object is bogus")
	}
}
