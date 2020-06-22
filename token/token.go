// Package token contains the token-types which our lexer produces,
// and which our parser understands.
package token

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
