package ast

import (
	"strings"
	"testing"
)

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
