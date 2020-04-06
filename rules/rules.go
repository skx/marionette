// Package rules contains the definition of the user supplied rules.
//
// A rule has a type, and then a series of key=value pairs.  Sometimes
// the values are strings, other times the are arrays
package rules

// Rule is the structure which contains a single rule
type Rule struct {

	// Type contains the rule-type
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

// NewRule creates a new rules.
func NewRule(t string, n string, p map[string]interface{}) *Rule {

	r := &Rule{}
	r.Type = t
	r.Name = n
	r.Params = p

	return r
}
