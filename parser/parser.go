// Package parser contains a parser for our input language.
//
// We consume tokens from the lexer, and attempt to process
// them into either:
//
//  1. A series of rules.
//
//  2. A series of variable assignments.
//
//  We support the inclusion of other files, and command
// expansion via backticks, but we're otherwise pretty
// minimal.
//
package parser

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/skx/marionette/ast"
	"github.com/skx/marionette/conditionals"
	"github.com/skx/marionette/lexer"
	"github.com/skx/marionette/token"
)

// Parser holds our state.
type Parser struct {
	// l is the handle to our lexer
	l *lexer.Lexer

	// curToken holds the current token from our lexer.
	curToken token.Token

	// peekToken holds the next token which will come from the lexer.
	//
	// We need lookahead for parsing (conditional) inclusion.
	peekToken token.Token
}

// New creates a new parser from the given input.
func New(input string) *Parser {

	// Create our object, and lexer
	p := &Parser{}
	p.l = lexer.New(input)

	// Ensure we're ready to process tokens.
	p.nextToken()

	return p
}

// Process parses our input, returning the AST which we will walk
// for the evaluation.
func (p *Parser) Process() (ast.Program, error) {

	// The program we return, and any error
	var program ast.Program
	var err error

	// Parse forever
	for {
		// Get the next token
		tok := p.nextToken()

		// Error-checking
		if tok.Type == token.ILLEGAL {
			return program, fmt.Errorf("illegal token: %v", tok)
		}
		if tok.Type == token.EOF {
			break
		}

		// Now parse the various tokens

		// Is this an assignment?
		if tok.Literal == "let" {

			// name
			name := p.nextToken()

			// =
			t := p.nextToken()
			if t.Type != token.ASSIGN {
				return program, fmt.Errorf("expected '=', got %v", t)
			}

			// value
			val := p.nextToken()

			// Error-checking.
			if val.Type == token.ILLEGAL || val.Type == token.EOF {
				return program, fmt.Errorf("unterminated assignment")
			}

			// assignment only handles strings/command-ouptut
			if val.Type != token.STRING && val.Type != token.BACKTICK {
				return program, fmt.Errorf("unexpected value for variable assignment; expected string or backtick, got %v", val)
			}

			// Add the node to our program, and continue
			program.Recipe = append(program.Recipe,
				&ast.Assign{Key: name.Literal, Value: val})
			continue
		}

		// Is this an include-file?
		if tok.Literal == "include" {

			// Get the thing we should include.
			t := p.nextToken()

			// We allow strings/backticks to be used
			if t.Type != token.STRING {
				return program, fmt.Errorf("only strings are supported for include statements; got %v", t)
			}

			// The include-command
			inc := &ast.Include{Source: t.Literal}

			// Look at the next token and see if it is a
			// conditional inclusion
			if p.peekTokenIs("if") || p.peekTokenIs("unless") {

				// skip the token - after saving it
				nxt := p.peekToken.Literal
				p.nextToken()

				// Get the name/arguments of the function call
				// we expect to come next.
				fname, args, error := p.parseFunctionCall()

				// error? then return that
				if error != nil {
					return program, error
				}

				// Otherwise save the condition.
				inc.ConditionType = nxt
				inc.ConditionRule = &conditionals.ConditionCall{Name: fname, Args: args}
			}

			// Add our rule onto the program, and continue
			program.Recipe = append(program.Recipe, inc)
			continue
		}

		// Otherwise it should be a block, which we need to parse
		// now.
		var tmp *ast.Rule
		tmp, err = p.parseBlock(tok.Literal)
		if err != nil {
			return program, err
		}

		// Add our rule onto the program, and continue
		program.Recipe = append(program.Recipe, tmp)
		continue
	}

	// No error
	return program, nil
}

// parseBlock parses the contents of modules' block.
//
// A block has the general form:
//
//  type [triggered] {
//       key1   => "value",
//       key2   => [ "foo", "bar", "baz" ],
//       unless => expression(),
//       if     => expression,
//  }
//
// The values of the keys can either be quoted strings,
// backtick-strings, or arrays of the same.
//
// The two keys `if` and `unless` have unquoted expressions
// as arguments.
//
func (p *Parser) parseBlock(ty string) (*ast.Rule, error) {

	r := &ast.Rule{}
	r.Name = ""
	r.Params = make(map[string]interface{})
	r.Type = ty

	// We should find either "triggered" or "{".
	t := p.nextToken()
	if t.Literal == "triggered" {
		r.Triggered = true
		t = p.nextToken()
	}
	if t.Type != token.LBRACE {
		return r, fmt.Errorf("expected '{', got %v", t)
	}

	// Now loop until we find the end of the block, which is "}".
	for {

		t = p.nextToken()

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

		//
		// Skip the assign
		//
		//  We expect:
		//
		//   "if|unless|blah" =>  FOO ( arg1, arg2 .. )
		//
		next := p.nextToken()
		if next.Literal != token.LASSIGN {
			return r, fmt.Errorf("expected => after conditional %s, got %v", name, next)
		}

		//
		// Is this a conditional key?
		//
		if name == "if" || name == "unless" {

			//
			// Get the name/arguments of the function call we
			// expect to come next.
			//
			fname, args, error := p.parseFunctionCall()

			if error != nil {
				return r, error
			}

			//
			// OK at this point we should have:
			//
			//  1. A rule-type (tType) such as "exists"
			//
			//  2. A collection of arguments.
			//
			// Save those values away in our interface map.
			//
			r.Params[name] = &conditionals.ConditionCall{Name: fname, Args: args}
			continue
		}

		// Read the value of the key
		value, err := p.readValue(name)
		if err != nil {
			return r, err
		}

		// Save it
		r.Params[name] = value
	}

	r.Name = p.getName(r.Params)
	return r, nil
}

// getName returns the name from the specified map, if one wasn't
// set then we return a random UUID.
func (p *Parser) getName(params map[string]interface{}) string {

	// Did we get a name parameter?
	n, ok := params["name"]
	if ok {

		// OK we did.  Was it a string?
		str, ok := n.(string)
		if ok {

			// Yes.  Use it.
			return str
		}
	}

	// Otherwise return a random UUID
	return uuid.New().String()
}

// parseFunctionCall is invoked for the `if` & `unless` handlers.
//
// It parsers Foo(Val,val,val...) and returns the argument collection.
func (p *Parser) parseFunctionCall() (string, []string, error) {

	var name string
	var args []string

	//
	// The function-call "exists", "equal", etc
	//
	tType := p.nextToken()
	if tType.Type != token.IDENT {
		return name, args, fmt.Errorf("expected identifier name after conditional %s, got %v", tType, tType)
	}
	name = tType.Literal

	//
	// Skip the opening bracket
	//
	open := p.nextToken()
	if open.Type != token.LPAREN {
		return name, args, fmt.Errorf("expected ( after conditional name %s, got %v", open, open)
	}

	//
	// Collect the arguments, until we get a close-bracket
	//
	t := p.nextToken()
	for t.Literal != ")" && t.Type != token.EOF {

		//
		// Append the argument, unless it is a comma
		//
		if t.Type != token.COMMA {

			args = append(args, t.Literal)
		}
		t = p.nextToken()
	}

	if t.Type == token.EOF {
		return name, args, fmt.Errorf("unexpected EOF in conditional")
	}

	return name, args, nil
}

// readValue returns the value associated with a name.
//
// The value is either a string, or an array of strings.
func (p *Parser) readValue(name string) (interface{}, error) {

	var a []string

	t := p.nextToken()

	// error checking
	if t.Type == token.ILLEGAL {
		return nil, fmt.Errorf("found illegal token:%v processing block %s", t, name)
	}
	if t.Type == token.EOF {
		return nil, fmt.Errorf("found end of file processing block %s", name)
	}

	// string or backticks
	if t.Type == token.STRING || t.Type == token.BACKTICK {
		return t.Literal, nil
	}

	// array?
	if t.Type != token.LSQUARE {
		return nil, fmt.Errorf("not a string or an array for value in block %s", name)
	}

	for {
		t := p.nextToken()

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

// nextToken moves to our next token from the lexer.
func (p *Parser) nextToken() token.Token {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()

	return p.curToken
}

// peekTokenIs tests if the next token has the given value.
func (p *Parser) peekTokenIs(t string) bool {
	return p.peekToken.Literal == t
}
