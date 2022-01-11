// Package token contains the token-types which our lexer produces,
// and which our parser understands.
package token

import "fmt"

// Type is a string
type Type string

// Token struct represent the lexer token
type Token struct {
	Type    Type
	Literal string
}

// pre-defined TokenTypes
const (
	ASSIGN   = "="
	BACKTICK = "`"
	COMMA    = ","
	EOF      = "EOF"
	IDENT    = "IDENT"
	ILLEGAL  = "ILLEGAL"
	LASSIGN  = "=>"
	LBRACE   = "{"
	LSQUARE  = "["
	RBRACE   = "}"
	RSQUARE  = "]"
	LPAREN   = "("
	RPAREN   = ")"
	STRING   = "STRING"
)

// String turns the token into a readable string
func (t Token) String() string {

	// string?
	if t.Type == STRING {
		return t.Literal
	}

	// backtick?
	if t.Type == BACKTICK {
		return "`" + t.Literal + "`"
	}

	// everything else is less pretty
	return fmt.Sprintf("token{Type:%s Literal:%s}", t.Type, t.Literal)
}
