// Package parser contains a parser for our input language.
//
// We consume tokens from the lexer, and return a "Program".
//
// The program consists of different types of things that we
// support at runtime:
//
//  1. A rule to process with one of our modules.
//
//  2. Variable assignments.
//
//  3. File inclusions.
//
//  4. AST nodes for various primitive types (strings, numbers, etc).
//
// Expansion of variables, handling of include-files, and command
// execution for the case of variable assignments, will all happen
// at run-time not within this parser.
//
// Importantly we ensure we don't leak the token.Token package outside
// the scope of our internals here - consumers of the parsed-programs
// shouldn't need to know/use that package.
package parser

import (
	"fmt"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/skx/marionette/ast"
	"github.com/skx/marionette/lexer"
	"github.com/skx/marionette/token"
)

// Parser holds our state.
type Parser struct {
	// Should we output our program?
	debug bool

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

	// Create our object, and the lexer it will use.
	p := &Parser{
		debug: false,
		l:     lexer.New(input),
	}

	// Should we output things to the console as we parse?
	if os.Getenv("DEBUG_PARSER") == "true" {
		p.debug = true
	}

	// Ensure we're ready to process tokens.
	p.nextToken()

	return p
}

// Parse parses our input, returning the AST which will be walked
// during program execution and evaluation.
func (p *Parser) Parse() (ast.Program, error) {

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

		// Now parse the various logical program-things we have.

		// Is this an assignment?
		if tok.Literal == "let" {

			// Parse the assignment-statement
			var let *ast.Assign
			let, err = p.parseLet()
			if err != nil {
				return program, err
			}

			if p.debug {
				fmt.Printf("%v\n", let)
			}

			// Add our rule onto the program, and continue
			program.Recipe = append(program.Recipe, let)
			continue
		}

		// Is this an include-file?
		if tok.Literal == "include" {

			// Parse the include-statement
			var inc *ast.Include
			inc, err = p.parseInclude()
			if err != nil {
				return program, err
			}

			if p.debug {
				fmt.Printf("%v\n", inc)
			}

			// Add our rule onto the program, and continue
			program.Recipe = append(program.Recipe, inc)
			continue
		}

		// Otherwise it should be a block, which we need to parse.
		var tmp *ast.Rule
		tmp, err = p.parseBlock(tok.Literal)
		if err != nil {
			return program, err
		}

		// If we're debugging then show what we produced.
		if p.debug {
			fmt.Printf("%v\n", tmp)
		}

		// Add our rule onto the program, and continue
		program.Recipe = append(program.Recipe, tmp)
	}

	// No error
	return program, nil
}

// parseLet parses an assignment statement
func (p *Parser) parseLet() (*ast.Assign, error) {

	// The statement we'll return
	let := &ast.Assign{}

	// name of the variable to which assignment is being made.
	name := p.nextToken()

	// name must be an identifier - not a string, number, boolean, etc.
	if name.Type != token.IDENT {
		return let, fmt.Errorf("assignment can only be made to identifiers, got %v", name)
	}

	// =
	t := p.nextToken()
	if t.Type != token.ASSIGN {
		return let, fmt.Errorf("expected '=', got %v", t)
	}

	// get the value, and parse it
	t = p.nextToken()
	val, err := p.parsePrimitive(t)

	// Error-checking.
	if err != nil {
		return let, err
	}

	// Look at the next token and see if it is a
	// conditional assignment.
	if p.peekTokenIs("if") || p.peekTokenIs("unless") {

		// skip the token - after saving it
		nxt := p.peekToken.Literal
		tok := p.nextToken()

		// Parse the function
		tok = p.nextToken()
		action, err := p.parsePrimitive(tok)

		if err != nil {
			return let, err
		}

		// Confirm the action is a Funcall
		faction, ok := action.(*ast.Funcall)
		if !ok {
			return let, fmt.Errorf("expected function-call after %s, got %v", nxt, action)
		}

		// Otherwise save the condition.
		let.ConditionType = nxt
		let.Function = faction
	}

	// Update the assignment node, and return it.
	let.Key = name.Literal
	let.Value = val

	return let, nil
}

// parseInclude parses an include-statement
func (p *Parser) parseInclude() (*ast.Include, error) {

	// The include statement we'll return
	inc := &ast.Include{}

	// Get the thing we should include.
	t := p.nextToken()

	// We only allow strings to be used for inclusion.
	if t.Type != token.STRING {
		return inc, fmt.Errorf("only strings are supported for include statements; got %v", t)
	}

	// The include-command
	inc.Source = t.Literal

	// Look at the next token and see if it is a
	// conditional inclusion
	if p.peekTokenIs("if") || p.peekTokenIs("unless") {

		// skip the token - after saving it
		nxt := p.peekToken.Literal
		tok := p.nextToken()

		// Parse the function
		tok = p.nextToken()
		action, err := p.parsePrimitive(tok)

		if err != nil {
			return inc, err
		}

		// Confirm the action is a Funcall
		faction, ok := action.(*ast.Funcall)
		if !ok {
			return inc, fmt.Errorf("expected function-call after %s, got %v", nxt, action)
		}

		// Otherwise save the condition.
		inc.ConditionType = nxt
		inc.Function = faction

	}

	return inc, nil

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
// The values of the keys can either be collections of our primitive
// types, or arrays of them.
//
// The two keys `if` and `unless` have unquoted expressions
// as arguments.
//
func (p *Parser) parseBlock(ty string) (*ast.Rule, error) {

	// Create an empty rule, which we'll populate
	r := &ast.Rule{}
	r.Name = ""
	r.Params = make(map[string]interface{})
	r.Type = ty

	// We should find either "triggered" or "{".
	t := p.nextToken()
	if t.Literal == "triggered" {

		// OK it is a triggered-rule.
		// record that and skip to the next token
		r.Triggered = true
		t = p.nextToken()
	}

	// "{"
	if t.Type != token.LBRACE {
		return r, fmt.Errorf("expected '{', got %v", t)
	}

	// Now loop until we find the end of the block, which is "}".
	//
	// The block will contain:
	//
	//    key1 => VALUE,
	//    key2 => VALUE2,
	//    key3 => [ VAL1, VAL2, VAL3, ],
	//
	// etc.
	//
	// So "key", then "=>", followed by either single values, or arrays
	// enclosed in "[" + "]" pairs.
	//
	// We can ignore commas, and semi-colons, whenever they appear.
	//
	for {

		// Get the next token.
		// "KEY"
		t = p.nextToken()

		// error checking?
		if t.Type == token.ILLEGAL {
			return r, fmt.Errorf("found illegal token:%v", t)
		}

		// end of file?
		if t.Type == token.EOF {
			return r, fmt.Errorf("found end of file")
		}

		// end of block?
		if t.Type == token.RBRACE {
			break
		}

		// commas are always skipped over.
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
		next := p.nextToken()
		if next.Literal != token.LASSIGN {
			return r, fmt.Errorf("expected => after conditional %s, got %v", name, next)
		}

		//
		// Is this a conditional key?
		//
		// We expect either "if" or "unless", and then a function
		//
		//   "if|unless" =>  FOO ( arg1, arg2 .. )
		if name == "if" || name == "unless" {

			// Parse the function
			tok := p.nextToken()
			action, err := p.parsePrimitive(tok)

			if err != nil {
				return r, err
			}

			// Confirm the action is a Funcall
			faction, ok := action.(*ast.Funcall)
			if !ok {
				return r, fmt.Errorf("expected function-call after %s, got %v", next, action)
			}

			// Otherwise save the condition.
			r.ConditionType = name
			r.Function = faction
			continue
		}

		//
		// OK at this point we've parsed:
		//
		//  KEY =>
		//
		// We need to find the value, which is either a single
		// token, or an array of tokens.
		//
		if p.peekToken.Type == token.LSQUARE {

			// OK this is an array of primitive values,
			// separated by commas.
			values, err := p.parseMultiplePrimitives()
			if err != nil {
				return r, err
			}

			// Save it
			r.Params[name] = values

		} else {

			// Single value.
			next = p.nextToken()
			value, err := p.parsePrimitive(next)
			if err != nil {
				return r, err
			}

			// Save it
			r.Params[name] = value
		}
	}

	// If there was no name setup for the rule then we
	// generate one.
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
		str, ok := n.(*ast.String)
		if ok {
			return str.Value
		}

		// Show a warning.
		if p.debug {
			fmt.Printf("WARNING: Name of rule is not *ast.String, got %T\n", n)
		}
	}

	// Return a random UUID for this rule.
	return uuid.New().String()
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

// parsePrimitive attempts to parse the given token as one of our
// primitive-types.
//
// This is used for block-values, be they single or multiple.
func (p *Parser) parsePrimitive(tok token.Token) (ast.Object, error) {

	// Return the appropriate AST-node, if we can.
	switch tok.Type {

	case token.BACKTICK:
		return &ast.Backtick{Value: tok.Literal}, nil

	case token.BOOLEAN:
		if tok.Literal == "true" {
			return &ast.Boolean{Value: true}, nil
		}
		return &ast.Boolean{Value: false}, nil

	case token.IDENT:
		// if this is an identifier and the next token is "("
		// then we've got a function-call
		name := tok.Literal

		if p.peekTokenIs(token.LPAREN) {

			// Arguments we'll build up
			var args []ast.Object

			// skip "("
			p.nextToken()

			// look for arguments
			t := p.nextToken()
			for t.Literal != ")" && t.Type != token.EOF {

				if t.Type != token.COMMA {

					val, err := p.parsePrimitive(t)
					if err != nil {
						return nil, err
					}
					args = append(args, val)
				}

				t = p.nextToken()
			}

			if t.Type == token.EOF {
				return nil, fmt.Errorf("unexpected EOF in function-call")
			}

			return &ast.Funcall{Name: name, Args: args}, nil

		}

	case token.NUMBER:
		val, err := strconv.ParseInt(tok.Literal, 0, 64)
		if err != nil {
			return nil, err
		}
		return &ast.Number{Value: val}, nil

	case token.STRING:
		return &ast.String{Value: tok.Literal}, nil

	}
	return nil, fmt.Errorf("unexpected type parsing primitive:%v", tok)
}

// parseMultiplePrimitives attempts to parse multiple values within a
// "[" + "]" separated block.
func (p *Parser) parseMultiplePrimitives() ([]ast.Object, error) {

	var ret []ast.Object
	var val ast.Object
	var err error

	for {
		// get the next token
		tok := p.nextToken()

		// error checking?
		if tok.Type == token.ILLEGAL {
			return ret, fmt.Errorf("found illegal token:%v", tok)
		}

		// end of file?
		if tok.Type == token.EOF {
			return ret, fmt.Errorf("found end of file")
		}

		// end of values?
		if tok.Type == token.RSQUARE {
			return ret, nil
		}

		// if it is a comma ignore it
		if tok.Type == token.COMMA {
			continue
		}

		// if it was the opening "[" ignore it
		if tok.Type == token.LSQUARE {
			continue
		}

		// OK then it is a single primitive value
		val, err = p.parsePrimitive(tok)
		if err != nil {
			return ret, err
		}

		ret = append(ret, val)
	}
}
