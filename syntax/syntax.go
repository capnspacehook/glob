package syntax

import (
	"github.com/capnspacehook/glob/syntax/ast"
	"github.com/capnspacehook/glob/syntax/lexer"
)

func Parse(s string) (*ast.Node, error) {
	return ast.Parse(lexer.NewLexer(s))
}

func IsSpecial(r rune) bool {
	return lexer.IsSpecial(r)
}
