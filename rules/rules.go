// Package rules contains the definition of the user supplied rules.
//
// A rule has a type, and then a series of key=value pairs.  Sometimes
// the values are strings, other times they are arrays.
package rules

import "fmt"

// Rule is the structure which contains a single rule
type Rule struct {

	// Type contains the rule-type.
	Type string

	// Name contains the name of the rule.
	Name string

	// Triggered is true if this rule is only triggered by
	// another rule notifying it.
	//
	// Triggered rules are ignored when processing our list.
	Triggered bool

	// Parameters contains the params supplied by the user.
	Params map[string]interface{}
}

// NewRule creates a new rule.
func NewRule(t string, n string, p map[string]interface{}) *Rule {

	r := &Rule{}
	r.Type = t
	r.Name = n
	r.Params = p

	return r
}

// String converts the given rule to a string representation.
func (r Rule) String() string {

	// $MODULE [triggered] {
	out := r.Type + " "
	if r.Triggered {
		out += "triggered "
	}
	out += "{\n"

	// Now the parameters
	for key, val := range r.Params {

		//
		// Pad the keys
		//
		k := key
		for len(k) < 12 {
			k = " " + k
		}

		if key == "if" || key == "unless" {
			out += fmt.Sprintf("%s => %s\n", k, val)
			continue
		}

		// string?
		str, valid := val.(string)
		if valid {
			out += fmt.Sprintf("%s => \"%s\",\n", k, str)
		}

		// array
		strs, ok := val.([]string)
		if ok {
			out += fmt.Sprintf("%s => [ \n ", k)
			for _, x := range strs {
				out += fmt.Sprintf("    \"%s\",\n", x)
			}
			out += "  ],\n"
		}
	}

	out += "}\n"

	return out
}
