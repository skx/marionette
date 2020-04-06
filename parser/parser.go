// Package parser contains a rudimentary parser for our input language.
package parser

import (
	"fmt"
	"os"

	"github.com/skx/marionette/lexer"
	"github.com/skx/marionette/rules"
	"github.com/skx/marionette/token"
)

// Parser holds our state.
type Parser struct {
	// l is the handle to our lexer
	l *lexer.Lexer

	// variables we've found
	vars map[string]string
}

// New creates a new parser from the given input.
func New(input string) *Parser {
	p := &Parser{}

	p.l = lexer.New(input)
	p.vars = make(map[string]string)
	return p
}

// Parse parses our input, returning an array of rules found,
// and any error which was encountered upon the way.
func (p *Parser) Parse() ([]rules.Rule, error) {

	// The rules we return
	var rules []rules.Rule
	var err error

	// Parse forever
	for {
		tok := p.l.NextToken()
		if tok.Type == token.ILLEGAL {
			return rules, fmt.Errorf("illegal token: %v", tok)
		}
		if tok.Type == token.EOF {
			break
		}

		// OK we expect an identifier.
		if tok.Type != token.IDENT {
			return nil, fmt.Errorf("unexpected input, expected identifier")
		}

		// Is this an assignment?
		if tok.Literal == "let" {
			//
			err := p.ParseVariable()
			if err != nil {
				return rules, err
			}
		} else {

			// OK then it must be a block statement
			tmp, err := p.ParseBlock(tok.Literal)
			if err != nil {
				return rules, err
			}

			rules = append(rules, tmp)
		}
	}

	return rules, err
}

// Variables returns any defined variables.
func (p *Parser) Variables() map[string]string {
	return p.vars
}

// ParseVariable parses a variable assignment, storing it in our map.
func (p *Parser) ParseVariable() error {

	// name
	name := p.l.NextToken()

	// =
	t := p.l.NextToken()
	if t.Type != token.ASSIGN {
		return fmt.Errorf("expected '=', got %v", t)
	}

	// value
	val := p.l.NextToken()

	p.vars[name.Literal] = val.Literal
	return nil
}

// ParseBlock parses the contents of modules' block.
func (p *Parser) ParseBlock(ty string) (rules.Rule, error) {

	var r rules.Rule
	r.Name = ""
	r.Params = make(map[string]interface{})
	r.Type = ty

	// We should find either "triggered" or "{"
	t := p.l.NextToken()
	if t.Literal == "triggered" {
		r.Triggered = true
		t = p.l.NextToken()
	}
	if t.Type != token.LBRACE {
		return r, fmt.Errorf("expected '{', got %v", t)
	}

	// Now loop until we find the end of the block find "}"
	for {

		t = p.l.NextToken()

		// error checking
		if t.Type == token.ILLEGAL {
			return r, fmt.Errorf("found illegal token:%v", t)
		}
		if t.Type == token.EOF {
			return r, fmt.Errorf("found end of file")
		}

		// end of block?
		if t.Type == token.RBRACE {
			break
		}

		// commas are skipped over
		if t.Type == token.COMMA {
			continue
		}

		// OK so we want "name = value"
		if t.Type != token.IDENT {
			return r, fmt.Errorf("expected literal in block, got %v", t)
		}

		// Record the name
		name := t.Literal

		// Now look for "=>"
		next := p.l.NextToken()
		if next.Literal != token.LASSIGN {
			return r, fmt.Errorf("expected => after name %s, got %v", name, next)
		}

		// Read the value
		value, err := p.ReadValue(name)
		if err != nil {
			return r, err
		}
		r.Params[name] = value
	}

	// If we got a name then place it in our struct
	n, ok := r.Params["name"]
	if ok {
		str, ok := n.(string)
		if ok {
			r.Name = str
		}
	}

	return r, nil
}

// ReadValue returns the value associated with a name.
//
// The value is either a string, or an array of strings.
func (p *Parser) ReadValue(name string) (interface{}, error) {

	// Helper to expand variables.
	mapper := func(val string) string {
		return p.vars[val]
	}

	var a []string

	t := p.l.NextToken()

	// error checking
	if t.Type == token.ILLEGAL {
		return nil, fmt.Errorf("found illegal token:%v processing block %s", t, name)
	}
	if t.Type == token.EOF {
		return nil, fmt.Errorf("found end of file processing block %s", name)
	}

	// string?
	if t.Type == token.STRING {
		return os.Expand(t.Literal, mapper), nil
	}

	// array?
	if t.Type != token.LSQUARE {
		return nil, fmt.Errorf("not a string or an array for value in block %s", name)
	}

	for {
		t := p.l.NextToken()

		// error checking
		if t.Type == token.ILLEGAL {
			return nil, fmt.Errorf("found illegal token:%v processing block %s", t, name)
		}
		if t.Type == token.EOF {
			return nil, fmt.Errorf("found end of file, processing block %s", name)
		}

		if t.Type == token.COMMA {
			continue
		}
		if t.Type == token.STRING {
			a = append(a, os.Expand(t.Literal, mapper))
		}
		if t.Type == token.RSQUARE {
			return a, nil
		}
	}
}
