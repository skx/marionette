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
	"unicode"

	"github.com/skx/marionette/token"
)

// Lexer is used as the lexer for our deployr "language".
type Lexer struct {
	position     int                  // current character position
	readPosition int                  // next character position
	ch           rune                 // current character
	characters   []rune               // rune slice of input string
	lookup       map[rune]token.Token // lookup map for simple tokens
}

// New a Lexer instance from string input.
func New(input string) *Lexer {
	l := &Lexer{characters: []rune(input),
		lookup: make(map[rune]token.Token)}
	l.readChar()

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

// NextToken to read next token, skipping the white space.
func (l *Lexer) NextToken() token.Token {

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
		// value we found
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
		tok.Literal = l.readIdentifier()
		tok.Type = token.IDENT
		return tok
	}

	// skip the character we've processed, and return the value
	l.readChar()
	return tok
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
func isIdentifier(ch rune) bool {
	return !isWhitespace(ch) && ch != rune(',') && ch != rune('(') && ch != rune(')') && ch != rune('=') && !isEmpty(ch)

}

// is the character white space?
func isWhitespace(ch rune) bool {
	return unicode.IsSpace(ch)
}

// is the given character empty?
func isEmpty(ch rune) bool {
	return rune(0) == ch
}
