// Package lexer contains a simple lexer for reading an input-string
// and converting it into a series of tokens.
//
// In terms of syntax we're not very complex, so our lexer only needs
// to care about simple tokens:
//
// - Comments
// - Strings
// - Some simple characters such as "(", ")", "[", "]", "=>", "=", etc.
// -
//
// We can catch some basic errors in the lexing stage, such as unterminated
// strings, but the parser is the better place to catch such things.
package lexer

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/skx/marionette/token"
)

// Lexer is used as the lexer for our deployr "language".
type Lexer struct {
	debug        bool                 // dump tokens as they're read?
	decimal      bool                 // convert numbers to decimal?
	position     int                  // current character position
	readPosition int                  // next character position
	ch           rune                 // current character
	characters   []rune               // rune slice of input string
	lookup       map[rune]token.Token // lookup map for simple tokens
}

// New a Lexer instance from string input.
func New(input string) *Lexer {
	l := &Lexer{
		characters: []rune(input),
		debug:      false,
		decimal:    false,
		lookup:     make(map[rune]token.Token),
	}
	l.readChar()

	if os.Getenv("DEBUG_LEXER") == "true" {
		l.debug = true
	}
	if os.Getenv("DECIMAL_NUMBERS") == "true" {
		l.decimal = true
	}

	//
	// Lookup map of simple token-types.
	//
	l.lookup['('] = token.Token{Literal: "(", Type: token.LPAREN}
	l.lookup[')'] = token.Token{Literal: ")", Type: token.RPAREN}
	l.lookup['['] = token.Token{Literal: "[", Type: token.LSQUARE}
	l.lookup[']'] = token.Token{Literal: "]", Type: token.RSQUARE}
	l.lookup['{'] = token.Token{Literal: "{", Type: token.LBRACE}
	l.lookup['}'] = token.Token{Literal: "}", Type: token.RBRACE}
	l.lookup[','] = token.Token{Literal: ",", Type: token.COMMA}
	l.lookup[rune(0)] = token.Token{Literal: "", Type: token.EOF}

	return l
}

// read one forward character
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.characters) {
		l.ch = rune(0)
	} else {
		l.ch = l.characters[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

// NextToken consumes and returns the next token from our input.
//
// It is a simple method which can optionally dump the tokens to the console
// if $DEBUG_LEXER is non-empty.
func (l *Lexer) NextToken() token.Token {

	tok := l.nextTokenReal()
	if l.debug {
		fmt.Printf("%v\n", tok)
	}

	return tok
}

// nextTokenReal does the real work of consuming and returning the next
// token from our input string.
func (l *Lexer) nextTokenReal() token.Token {
	var tok token.Token
	l.skipWhitespace()

	// skip single-line comments
	//
	// This also skips the shebang line at the start of a file - as
	// "#!/usr/bin/blah" is treated as a comment.
	if l.ch == rune('#') {
		l.skipComment()
		return (l.NextToken())
	}

	// Semi-colons are skipped, always.
	if l.ch == rune(';') {
		l.readChar()
		return (l.NextToken())
	}

	// Was this a simple token-type?
	val, ok := l.lookup[l.ch]
	if ok {
		// Yes, then skip the character itself, and return the
		// value we found.
		l.readChar()
		return val

	}

	// OK it wasn't a simple type
	switch l.ch {
	case rune('='):
		tok.Literal = "="
		tok.Type = token.ASSIGN
		if l.peekChar() == rune('>') {
			l.readChar()

			tok.Type = token.LASSIGN
			tok.Literal = "=>"
		}
	case rune('`'):
		str, err := l.readBacktick()

		if err == nil {
			tok.Type = token.BACKTICK
			tok.Literal = str
		} else {
			tok.Type = token.ILLEGAL
			tok.Literal = err.Error()
		}
	case rune('"'):
		str, err := l.readString()

		if err == nil {
			tok.Type = token.STRING
			tok.Literal = str
		} else {
			tok.Type = token.ILLEGAL
			tok.Literal = err.Error()
		}
	default:
		// is it a number?
		if isDigit(l.ch) {
			// Read it.
			tok = l.readDecimal()
			return tok
		}

		// is it an ident?
		tok.Literal = l.readIdentifier()
		tok.Type = token.IDENT

		// We don't have keywords, but we'll convert
		// the ident "true" or "false" into a boolean-type.
		if tok.Literal == "true" || tok.Literal == "false" {
			tok.Type = token.BOOLEAN
		}

		return tok
	}

	// skip the character we've processed, and return the value
	l.readChar()
	return tok
}

// readDecimal returns a token consisting of decimal numbers, base 10, 2, or
// 16.
func (l *Lexer) readDecimal() token.Token {

	str := ""

	// We usually just accept digits.
	accept := "0123456789"

	// But if we have `0x` as a prefix we accept hexadecimal instead.
	if l.ch == '0' && l.peekChar() == 'x' {
		accept = "0x123456789abcdefABCDEF"
	}

	// If we have `0b` as a prefix we accept binary digits only.
	if l.ch == '0' && l.peekChar() == 'b' {
		accept = "b01"
	}

	// While we have a valid character append it to our
	// result and keep reading/consuming characters.
	for strings.Contains(accept, string(l.ch)) {
		str += string(l.ch)
		l.readChar()
	}

	// Don't convert the number to decimal - just use the literal value.
	if !l.decimal {
		return token.Token{Type: token.NUMBER, Literal: str}
	}

	// OK convert to an integer, which we'll later turn to a string.
	//
	// We do this so we can convert 0xff -> "255", or "0b0011" to "3".
	val, err := strconv.ParseInt(str, 0, 64)
	if err != nil {
		tok := token.Token{Type: token.ILLEGAL, Literal: err.Error()}
		return tok
	}

	// Now return that number as a string.
	return token.Token{Type: token.NUMBER, Literal: fmt.Sprintf("%d", val)}
}

// read Identifier
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isIdentifier(l.ch) {
		l.readChar()
	}
	return string(l.characters[position:l.position])
}

// skip white space
func (l *Lexer) skipWhitespace() {
	for isWhitespace(l.ch) {
		l.readChar()
	}
}

// skip comment (until the end of the line).
func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != rune(0) {
		l.readChar()
	}
	l.skipWhitespace()
}

// read string
func (l *Lexer) readString() (string, error) {
	out := ""

	for {
		l.readChar()
		if l.ch == '"' {
			break
		}
		if l.ch == rune(0) {
			return "", errors.New("unterminated string")
		}

		//
		// Handle \n, \r, \t, \", etc.
		//
		if l.ch == '\\' {

			// Line ending with "\" + newline
			if l.peekChar() == '\n' {
				// consume the newline.
				l.readChar()
				continue
			}

			l.readChar()

			if l.ch == rune('n') {
				l.ch = '\n'
			}
			if l.ch == rune('r') {
				l.ch = '\r'
			}
			if l.ch == rune('t') {
				l.ch = '\t'
			}
			if l.ch == rune('"') {
				l.ch = '"'
			}
			if l.ch == rune('\\') {
				l.ch = '\\'
			}
		}
		out = out + string(l.ch)

	}

	return out, nil
}

// read a backtick-enquoted string
func (l *Lexer) readBacktick() (string, error) {
	out := ""

	for {
		l.readChar()
		if l.ch == '`' {
			break
		}
		if l.ch == rune(0) {
			return "", errors.New("unterminated backtick")
		}
		out = out + string(l.ch)
	}

	return out, nil
}

// peek ahead at the next character
func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.characters) {
		return rune(0)
	}
	return l.characters[l.readPosition]
}

// determinate whether the given character is legal within an identifier or not.
//
// This is very permissive.
func isIdentifier(ch rune) bool {
	return !isWhitespace(ch) &&
		ch != rune(',') &&
		ch != rune('(') &&
		ch != rune(')') &&
		ch != rune('{') &&
		ch != rune('}') &&
		ch != rune('=') &&
		ch != rune(';') &&
		!isEmpty(ch)
}

// Is the character white space?
func isWhitespace(ch rune) bool {
	return unicode.IsSpace(ch)
}

// Is the given character empty?
func isEmpty(ch rune) bool {
	return rune(0) == ch
}

// Is the given character a digit?
func isDigit(ch rune) bool {
	return rune('0') <= ch && ch <= rune('9')
}
