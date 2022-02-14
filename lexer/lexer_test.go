package lexer

import (
	"os"
	"testing"

	"github.com/skx/marionette/token"
)

// TestEmpty tests a couple of different empty strings
func TestEmpty(t *testing.T) {
	empty := []string{
		";;;;;;;;;;;;;;;",
		"",
		"#!/usr/bin/blah",
		"#!/usr/bin/blah\n# Comment1\n# Comment2",
	}

	for _, line := range empty {
		lexer := New(line)
		result := lexer.NextToken()

		if result.Type != token.EOF {
			t.Fatalf("First token of empty input is %v", result)
		}
	}

}

// TestAssign tests we can assign something.
func TestAssign(t *testing.T) {
	input := `let foo = "steve";`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "let"},
		{token.IDENT, "foo"},
		{token.ASSIGN, "="},
		{token.STRING, "steve"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestEscape ensures that strings have escape-characters processed.
func TestStringEscape(t *testing.T) {
	input := `"Steve\n\r\\" "Kemp\n\t\n" "Inline \"quotes\"."`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STRING, "Steve\n\r\\"},
		{token.STRING, "Kemp\n\t\n"},
		{token.STRING, "Inline \"quotes\"."},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestComments ensures that single-line comments work.
func TestComments(t *testing.T) {
	input := `# This is a comment
"Steve"
# This is another comment`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STRING, "Steve"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestShebang skips the shebang
func TestShebang(t *testing.T) {
	input := `#!/usr/bin/env marionette
"Steve"
# This is another comment`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STRING, "Steve"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestUnterminatedString ensures that an unclosed-string is an error
func TestUnterminatedString(t *testing.T) {
	input := `#!/usr/bin/env marionette
"Steve`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.ILLEGAL, "unterminated string"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestBacktick string ensures that an backtick-string is OK.
func TestBacktick(t *testing.T) {
	input := "`ls`"

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.BACKTICK, "ls"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestUnterminatedBacktick string ensures that an unclosed-backtick is an error
func TestUnterminatedBacktick(t *testing.T) {
	input := "`Steve"

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.ILLEGAL, "unterminated backtick"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestContinue checks we continue newlines.
func TestContinue(t *testing.T) {
	input := `#!/usr/bin/env marionette
"This is a test \
which continues"
`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STRING, "This is a test which continues"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestSpecial ensures we can recognize special characters.
func TestSpecial(t *testing.T) {
	input := `[]{},=>=()`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.LSQUARE, "["},
		{token.RSQUARE, "]"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.COMMA, ","},
		{token.LASSIGN, "=>"},
		{token.ASSIGN, "="},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// Test15Assignment ensures that bug #15 is resolved
//    https://github.com/skx/marionette/issues/15
func Test15Assignment(t *testing.T) {
	input := `let foo="bar"`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "let"},
		{token.IDENT, "foo"},
		{token.ASSIGN, "="},
		{token.STRING, "bar"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}

	//
	// We've parsed the whole input.
	//
	// Reading more should just return \0.
	//
	i := 0
	for i < 10 {

		p := l.peekChar()
		if p != rune(0) {
			t.Errorf("after reading past input we didn't get null")
		}
		i++
	}
}

// TestInteger tests that we parse integers appropriately.
func TestInteger(t *testing.T) {
	os.Setenv("DECIMAL_NUMBERS", "true")

	type TestCase struct {
		input  string
		output string
	}

	tests := []TestCase{
		{input: "3", output: "3"},
		{input: "0xff", output: "255"},
		{input: "0b11111111", output: "255"},
	}

	for _, tst := range tests {

		lex := New(tst.input)
		tok := lex.NextToken()

		if tok.Literal != tst.output {
			t.Fatalf("error lexing %s - expected:%s got:%s", tst.input, tst.output, tok.Literal)
		}
	}
}
