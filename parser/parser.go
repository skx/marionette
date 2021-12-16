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
package parser

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/google/uuid"
	"github.com/skx/marionette/conditionals"
	"github.com/skx/marionette/environment"
	"github.com/skx/marionette/lexer"
	"github.com/skx/marionette/rules"
	"github.com/skx/marionette/token"
)

// Parser holds our state.
type Parser struct {
	// l is the handle to our lexer
	l *lexer.Lexer

	// curToken holds the current token from our lexer.
	curToken token.Token

	// peekToken holds the next token which will come from the lexer.
	peekToken token.Token

	// e is the handle to our environment
	e *environment.Environment

	// Files we've included
	included map[string]bool
}

// New creates a new parser from the given input.
func New(input string) *Parser {
	p := &Parser{}
	p.l = lexer.New(input)
	p.e = environment.New()
	p.included = make(map[string]bool)

	p.nextToken()

	return p
}

// NewWithEnvironment creates a new parser, along with a defined
// environment.
func NewWithEnvironment(input string, env *environment.Environment) *Parser {
	p := New(input)
	p.e = env
	return p
}

// mapper is a helper to expand variables.
//
// ${foo} will be converted to the contents of the variable named foo
// which was created with `let foo = "bar"`, or failing that the contents
// of the environmental variable named `foo`.
func (p *Parser) mapper(val string) string {

	// Lookup a variable which exists?
	res, ok := p.e.Get(val)
	if ok {
		return res
	}

	// Lookup an environmental variable?
	return os.Getenv(val)
}

// expand processes a token returned from the parser, returning
// the appropriate value.
//
// The expansion really means two things:
//
// 1. If the string contains variables ${foo} replace them.
//
// 2. If the token is a backtick operation then run the command
//    and return the value.
//
func (p *Parser) expand(tok token.Token) (string, error) {

	// Get the argument, and expand variables
	value := tok.Literal
	value = os.Expand(value, p.mapper)

	// If this is a backtick we replace the value
	// with the result of running the command.
	if tok.Type == token.BACKTICK {

		tmp, err := p.runCommand(value)
		if err != nil {
			return "", fmt.Errorf("error running %s: %s", value, err.Error())
		}

		value = tmp
	}

	// Return the value we've found.
	return value, nil
}

// mark the given files as having already been included
func (p *Parser) includedAlready(seen map[string]bool) {
	for k := range seen {
		p.included[k] = true
	}
}

// Parse parses our input, returning an array of rules found,
// and any error which was encountered upon the way.
func (p *Parser) Parse() ([]rules.Rule, error) {

	// The rules we return
	var found []rules.Rule
	var err error

	// Parse forever
	for {

		// Get the next token.
		tok := p.nextToken()

		// Error-checking
		if tok.Type == token.ILLEGAL {
			return nil, fmt.Errorf("illegal token: %v", tok)
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

			// If so parse it.
			err = p.parseVariable()
			if err != nil {
				return found, err
			}

			continue
		}

		// Is this an include-file?
		if tok.Literal == "include" {

			// Get the thing we should include.
			t := p.nextToken()

			// We allow strings/backticks to be used
			if t.Type != token.STRING && t.Type != token.BACKTICK {
				return found, fmt.Errorf("only strings/backticks supported for include statements; got %v", t)
			}

			//
			// Expand variables in the argument, and
			// run the appropriate command if the token
			// is a backtick.
			//
			path, er := p.expand(t)
			if er != nil {
				return found, er
			}

			//
			// Inclusions might be conditional, so look to
			// see if the next token is "unless" or "if"
			//
			nxt := ""

			//
			// Conditionals we support.
			//
			var cond []*conditionals.ConditionCall

			// Error-checking.
			if p.peekTokenIs("if") || p.peekTokenIs("unless") {

				// skip the token - after saving it
				nxt = p.peekToken.Literal
				p.nextToken()

				//
				//
				// Get the name/arguments of the function call we
				// expect to come next.
				//
				fname, args, error := p.parseFunctionCall()

				if error != nil {
					return found, error
				}

				cond = append(cond, &conditionals.ConditionCall{Name: fname, Args: args})

			}

			//
			// Have we already included this file?
			//
			// If so then we'll ignore the second attempt.
			//
			if p.included[path] {
				continue
			}

			//
			// Mark the file as having been included
			//
			p.included[path] = true

			//
			// Read the file we're supposed to process.
			//
			data, er := ioutil.ReadFile(path)
			if er != nil {
				return found, er
			}

			//
			// Create a new parser instance, making sure
			// that it uses the same environment we're using.
			//
			tmp := NewWithEnvironment(string(data), p.e)

			//
			// Also make sure we propagate the files
			// we've already seen.
			//
			// This will ensure that recursive includes do
			// not cause us problems.
			tmp.includedAlready(p.included)

			//
			// Now parse the new input.
			//
			rules, er := tmp.Parse()
			if er != nil {
				return found, er
			}

			//
			// Append the results of what we received
			// to what we've already done in the main-file.
			//
			// Adding in the conditional if we needed to
			//
			for _, ent := range rules {

				if len(cond) > 0 {
					ent.Params["include_"+nxt] = cond[0]
				}
				found = append(found, ent)
			}

			continue
		}

		// Otherwise it must be a block statement.
		var r rules.Rule

		r, err = p.parseBlock(tok.Literal)
		if err != nil {
			return nil, err
		}

		found = append(found, r)
	}

	return found, err
}

// parseVariable parses a variable assignment, storing it in our map.
func (p *Parser) parseVariable() error {

	// name
	name := p.nextToken()

	// =
	t := p.nextToken()
	if t.Type != token.ASSIGN {
		return fmt.Errorf("expected '=', got %v", t)
	}

	// value
	val := p.nextToken()

	// Error-checking.
	if val.Type == token.ILLEGAL || val.Type == token.EOF {
		return fmt.Errorf("unterminated assignment")
	}

	// assignment only handles strings/command-ouptut
	if val.Type != token.STRING && val.Type != token.BACKTICK {
		return fmt.Errorf("unexpected value for variable assignment; expected string or backtick, got %v", val)
	}

	// Expand variables in the string, if present, and process
	// the command if it uses a backtick.
	value, err := p.expand(val)
	if err != nil {
		return err
	}

	// Set the value in the environment
	p.e.Set(name.Literal, value)

	return nil
}

// runCommand returns the output of the specified command
func (p *Parser) runCommand(command string) (string, error) {

	// Are we running under a fuzzer?  If so disable this
	if os.Getenv("FUZZ") == "FUZZ" {
		return command, nil
	}

	// Build up the thing to run, using a shell so that
	// we can handle pipes/redirection.
	toRun := []string{"/bin/bash", "-c", command}

	// Run the command
	cmd := exec.Command(toRun[0], toRun[1:]...)

	// Get the output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running command '%s' %s", command, err.Error())
	}

	// Strip trailing newline.
	ret := strings.TrimSuffix(string(output), "\n")
	return ret, nil
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
func (p *Parser) parseBlock(ty string) (rules.Rule, error) {

	var r rules.Rule
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

			//
			// Expand any variable in the
			// string, and if it is a backtick
			// run the command.
			//
			value, err := p.expand(t)
			if err != nil {
				return name, args, err
			}
			args = append(args, value)
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

	// string or backticks get expanded
	if t.Type == token.STRING || t.Type == token.BACKTICK {
		value, err := p.expand(t)
		return value, err
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
			a = append(a, os.Expand(t.Literal, p.mapper))
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
