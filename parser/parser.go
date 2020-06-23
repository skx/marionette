// Package parser contains a rudimentary parser for our input language.
package parser

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/google/uuid"
	"github.com/skx/marionette/lexer"
	"github.com/skx/marionette/rules"
	"github.com/skx/marionette/token"
)

// Condition holds a conditional expression.
//
// Currently we support two types "exists" and "equal", which can
// be used as the values for magical blocks "if" and "unless".
//
type Condition struct {
	// Name has the name of the conditional operation.
	Name string

	// Args contains the arguments
	Args []string
}

// String converts a Condition to a string.
func (c Condition) String() string {
	return fmt.Sprintf("%s(%s)", c.Name, strings.Join(c.Args, ","))
}

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

// mapper is a helper to expand variables.
func (p *Parser) mapper(val string) string {
	return p.vars[val]
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
		tok := p.l.NextToken()

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
		} else if tok.Literal == "include" {

			// Get the thing we should include.
			t := p.l.NextToken()

			// We allow strings/backticks to be used
			if t.Type != token.STRING && t.Type != token.BACKTICK {
				return found, fmt.Errorf("only strings/backticks supported for include statements; got %v", t)
			}

			// Get the argument, and expand variables
			path := t.Literal
			path = os.Expand(path, p.mapper)

			// If this is a backtick we replace the value
			// with the result of running the command.
			if t.Type == token.BACKTICK {
				out, er := p.runCommand(path)
				if er != nil {
					return found, fmt.Errorf("error running %s: %s", path, er.Error())
				}
				path = out
			}

			// Read the file.
			data, er := ioutil.ReadFile(path)
			if er != nil {
				return found, er
			}

			//
			// Parse it via a new instance of the parser
			//
			tmp := New(string(data))
			rules, er := tmp.Parse()
			if er != nil {
				return found, er
			}

			//
			// Append the results of what we received
			// to what we've already done in the main-file.
			//
			found = append(found, rules...)

		} else {

			// Otherwise it must be a block statement.
			var r rules.Rule

			r, err = p.parseBlock(tok.Literal)
			if err != nil {
				return nil, err
			}

			found = append(found, r)
		}
	}

	return found, err
}

// Variables returns any defined variables.
func (p *Parser) Variables() map[string]string {
	return p.vars
}

// parseVariable parses a variable assignment, storing it in our map.
func (p *Parser) parseVariable() error {

	// name
	name := p.l.NextToken()

	// =
	t := p.l.NextToken()
	if t.Type != token.ASSIGN {
		return fmt.Errorf("expected '=', got %v", t)
	}

	// value
	val := p.l.NextToken()

	if val.Type == token.ILLEGAL || val.Type == token.EOF {
		return fmt.Errorf("unterminated assignment")
	}

	// assignment only handles strings/command-ouptut
	if val.Type != token.STRING && val.Type != token.BACKTICK {
		return fmt.Errorf("unexpected value for variable assignment; expected string or backtick, got %v", val)
	}

	// replace backtick with the appropriate output
	if val.Type == token.BACKTICK {
		out, err := p.runCommand(val.Literal)
		if err != nil {
			return fmt.Errorf("error running %s: %s", val.Literal, err.Error())
		}
		val.Literal = out
	}

	p.vars[name.Literal] = val.Literal
	return nil
}

// runCommand returns the output of the specified command
func (p *Parser) runCommand(command string) (string, error) {

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
func (p *Parser) parseBlock(ty string) (rules.Rule, error) {

	var r rules.Rule
	r.Name = ""
	r.Params = make(map[string]interface{})
	r.Type = ty

	// We should find either "triggered" or "{".
	t := p.l.NextToken()
	if t.Literal == "triggered" {
		r.Triggered = true
		t = p.l.NextToken()
	}
	if t.Type != token.LBRACE {
		return r, fmt.Errorf("expected '{', got %v", t)
	}

	// Now loop until we find the end of the block, which is "}".
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

		//
		// Is this a conditional key?
		//
		if name == "if" || name == "unless" {

			//
			// Skip the assign
			//
			//  We expect:
			//
			//   "if" =>  FOO ( arg1, arg2 .. )
			//
			next := p.l.NextToken()
			if next.Literal != token.LASSIGN {
				return r, fmt.Errorf("expected => after conditional %s, got %v", name, next)
			}

			//
			// The type of operation "exists", "equal"
			//
			tType := p.l.NextToken()
			if tType.Type != token.IDENT {
				return r, fmt.Errorf("expected identifier name after conditional %s, got %v", tType, tType)
			}

			//
			// Skip the opening bracket
			//
			open := p.l.NextToken()
			if open.Type != token.LPAREN {
				return r, fmt.Errorf("expected ( after conditional name %s, got %v", open, open)
			}

			//
			// Collect the arguments, until we get a close-bracket
			//
			var args []string

			t := p.l.NextToken()
			for t.Literal != ")" && t.Type != token.EOF {

				//
				// Append the argument, unless it is a comma
				//
				if t.Type != token.COMMA {

					//
					// Expand any variable.
					//
					val := os.Expand(t.Literal, p.mapper)

					//
					// Backticks need to be expanded
					// of course.
					//
					if t.Type == token.BACKTICK {
						out, err := p.runCommand(val)
						if err != nil {
							return r, fmt.Errorf("error running %s: %s", val, err.Error())
						}

						val = out
					}

					args = append(args, val)
				}
				t = p.l.NextToken()
			}

			if t.Type == token.EOF {
				return r, fmt.Errorf("unexpected EOF in conditional")
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
			r.Params[name] = &Condition{Name: tType.Literal, Args: args}
		} else {

			// Now look for "=>"
			next := p.l.NextToken()
			if next.Literal != token.LASSIGN {
				return r, fmt.Errorf("expected => after name %s, got %v", name, next)
			}

			// Read the value
			value, err := p.readValue(name)
			if err != nil {
				return r, err
			}

			r.Params[name] = value
		}
	}

	// If we got a name then place it in our struct
	n, ok := r.Params["name"]
	if ok {
		str, ok := n.(string)
		if ok {
			r.Name = str
		}
	} else {

		// Generate a fakename if there is none present.
		id := uuid.New()
		r.Name = id.String()
	}

	return r, nil
}

// readValue returns the value associated with a name.
//
// The value is either a string, or an array of strings.
func (p *Parser) readValue(name string) (interface{}, error) {

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
		return os.Expand(t.Literal, p.mapper), nil
	}

	// backtick?
	if t.Type == token.BACKTICK {
		cmd := os.Expand(t.Literal, p.mapper)

		out, err := p.runCommand(cmd)
		if err != nil {
			return "", fmt.Errorf("error running %s: %s", cmd, err.Error())
		}
		return out, nil

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
			a = append(a, os.Expand(t.Literal, p.mapper))
		}

		if t.Type == token.RSQUARE {
			return a, nil
		}
	}
}
