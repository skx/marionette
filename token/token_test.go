package token

import "testing"

func TestTokenString(t *testing.T) {
	// string
	s := &Token{Type: STRING, Literal: "Moi"}
	if s.String() != "Moi" {
		t.Fatalf("Unexpected string-version of token.STRING")
	}

	// backtick
	b := &Token{Type: BACKTICK, Literal: "/bin/ls"}
	if b.String() != "`/bin/ls`" {
		t.Fatalf("Unexpected string-version of token.BACKTICK")
	}

	// misc
	m := &Token{Type: LSQUARE, Literal: "["}
	if m.String() != "token{Type:[ Literal:[}" {
		t.Fatalf("Unexpected string-version of token.LSQUARE")
	}

}
