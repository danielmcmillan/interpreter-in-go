package ast

import (
	"testing"

	"danielmcm.com/interpreterbook/token"
)

func TestString(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&LetStatement{
				Token: token.Token{Type: token.LET, Literal: "let"},
				Name: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "myVar"},
					Value: "myVar",
				},
				Value: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "anotherVar"},
					Value: "anotherVar",
				},
			},
		},
	}

	str := program.String()
	if str != "let myVar = anotherVar;" {
		t.Errorf("program.String() is incorrect: %q", str)
	}
}
