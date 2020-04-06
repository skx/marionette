// Package parser contains a rudimentary parser for our input language.
package parser

import (
	"fmt"

	"github.com/skx/marionette/lexer"
	"github.com/skx/marionette/rules"
	"github.com/skx/marionette/token"
)

// Parser holds our state.
type Parser struct {
	// l is the handle to our lexer
	l *lexer.Lexer
}

// New creates a new parser from the given input.
func New(input string) *Parser {
	p := &Parser{}

	p.l = lexer.New(input)
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

		// OK we expect an identifier, then a block
		if tok.Type == token.IDENT {

			tmp, err := p.ParseBlock(tok.Literal)
			if err != nil {
				return rules, err
			}

			rules = append(rules, tmp)
		}
	}

	return rules, err
}

// ParseBlock parses the contents of modules' block.
func (p *Parser) ParseBlock(ty string) (rules.Rule, error) {

	var r rules.Rule
	r.Name = ""
	r.Params = make(map[string]interface{})
	r.Type = ty

	// We should find "{"
	t := p.l.NextToken()
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

		// Now look for "="
		next := p.l.NextToken()
		if next.Literal != "=>" {
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
		return t.Literal, nil
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
			a = append(a, t.Literal)
		}
		if t.Type == token.RSQUARE {
			return a, nil
		}
	}
}
