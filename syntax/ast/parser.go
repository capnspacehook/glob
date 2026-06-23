package ast

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/capnspacehook/glob/syntax/lexer"
)

type Lexer interface {
	Next() lexer.Token
}

type parseFn func(*Node, Lexer) (parseFn, *Node, error)

func Parse(lexer Lexer) (*Node, error) {
	var parser parseFn

	root := NewNode(KindPattern, nil)

	var (
		tree *Node
		err  error
	)
	for parser, tree = parserMain, root; parser != nil; {
		parser, tree, err = parser(tree, lexer)
		if err != nil {
			return nil, err
		}
	}

	return root, nil
}

func parserMain(tree *Node, lex Lexer) (parseFn, *Node, error) {
	for {
		token := lex.Next()
		switch token.Type {
		case lexer.EOF:
			return nil, tree, nil

		case lexer.Error:
			return nil, tree, errors.New(token.Raw)

		case lexer.Text:
			Insert(tree, NewNode(KindText, Text{token.Raw}))
			return parserMain, tree, nil

		case lexer.Any:
			Insert(tree, NewNode(KindAny, nil))
			return parserMain, tree, nil

		case lexer.Super:
			Insert(tree, NewNode(KindSuper, nil))
			return parserMain, tree, nil

		case lexer.Single:
			Insert(tree, NewNode(KindSingle, nil))
			return parserMain, tree, nil

		case lexer.CharClassOpen:
			c := NewNode(KindCharClass, CharClass{})
			Insert(tree, c)

			return parserRange, c, nil

		case lexer.TermsOpen:
			a := NewNode(KindAnyOf, nil)
			Insert(tree, a)

			p := NewNode(KindPattern, nil)
			Insert(a, p)

			return parserMain, p, nil

		case lexer.Separator:
			p := NewNode(KindPattern, nil)
			Insert(tree.Parent, p)

			return parserMain, p, nil

		case lexer.TermsClose:
			return parserMain, tree.Parent.Parent, nil

		default:
			return nil, tree, fmt.Errorf("unexpected token: %s", token)
		}
	}
}

func parserRange(tree *Node, lex Lexer) (parseFn, *Node, error) {
	var (
		chars    string
		ranges   []Range
		curRange Range
	)

	for {
		token := lex.Next()
		switch token.Type {
		case lexer.EOF:
			return nil, tree, errors.New("unexpected end")

		case lexer.Error:
			return nil, tree, errors.New(token.Raw)

		case lexer.Not:
			c, ok := tree.Value.(CharClass)
			if !ok {
				return nil, tree, fmt.Errorf("unexpected type for character class node: %T", tree.Value)
			}
			c.Not = true
			tree.Value = c

		case lexer.RangeLow:
			r, w := utf8.DecodeRuneInString(token.Raw)
			if len(token.Raw) > w {
				return nil, tree, errors.New("unexpected length of low character")
			}
			curRange.Low = r

		case lexer.RangeHigh:
			r, w := utf8.DecodeRuneInString(token.Raw)
			if len(token.Raw) > w {
				return nil, tree, errors.New("unexpected length of high character")
			}

			if r < curRange.Low {
				return nil, tree, fmt.Errorf("high character %s must be greater than low %s", string(r), string(curRange.Low))
			}
			curRange.High = r
			ranges = append(ranges, curRange)

		case lexer.Text:
			chars += token.Raw

		case lexer.CharClassClose:
			if len(chars) > 0 {
				Insert(tree, NewNode(KindList, List{
					Chars: chars,
				}))
			}
			for _, r := range ranges {
				Insert(tree, NewNode(KindRange, r))
			}

			return parserMain, tree.Parent, nil
		}
	}
}
