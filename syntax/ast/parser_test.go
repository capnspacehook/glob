package ast

import (
	"reflect"
	"testing"

	"github.com/capnspacehook/glob/syntax/lexer"
)

type stubLexer struct {
	tokens []lexer.Token
	pos    int
}

func (s *stubLexer) Next() (ret lexer.Token) {
	if s.pos == len(s.tokens) {
		return lexer.Token{Type: lexer.EOF, Raw: ""}
	}
	ret = s.tokens[s.pos]
	s.pos++
	return
}

func TestParseString(t *testing.T) {
	tests := []struct {
		tokens []lexer.Token
		tree   *Node
	}{
		{
			// pattern: "abc",
			tokens: []lexer.Token{
				{Type: lexer.Text, Raw: "abc"},
				{Type: lexer.EOF, Raw: ""},
			},
			tree: NewNode(
				KindPattern, nil,
				NewNode(KindText, Text{Text: "abc"}),
			),
		},
		{
			// pattern: "a*c",
			tokens: []lexer.Token{
				{Type: lexer.Text, Raw: "a"},
				{Type: lexer.Any, Raw: "*"},
				{Type: lexer.Text, Raw: "c"},
				{Type: lexer.EOF, Raw: ""},
			},
			tree: NewNode(
				KindPattern, nil,
				NewNode(KindText, Text{Text: "a"}),
				NewNode(KindAny, nil),
				NewNode(KindText, Text{Text: "c"}),
			),
		},
		{
			// pattern: "a**c",
			tokens: []lexer.Token{
				{Type: lexer.Text, Raw: "a"},
				{Type: lexer.Super, Raw: "**"},
				{Type: lexer.Text, Raw: "c"},
				{Type: lexer.EOF, Raw: ""},
			},
			tree: NewNode(
				KindPattern, nil,
				NewNode(KindText, Text{Text: "a"}),
				NewNode(KindSuper, nil),
				NewNode(KindText, Text{Text: "c"}),
			),
		},
		{
			// pattern: "a?c",
			tokens: []lexer.Token{
				{Type: lexer.Text, Raw: "a"},
				{Type: lexer.Single, Raw: "?"},
				{Type: lexer.Text, Raw: "c"},
				{Type: lexer.EOF, Raw: ""},
			},
			tree: NewNode(
				KindPattern, nil,
				NewNode(KindText, Text{Text: "a"}),
				NewNode(KindSingle, nil),
				NewNode(KindText, Text{Text: "c"}),
			),
		},
		{
			// pattern: "[!a-z]",
			tokens: []lexer.Token{
				{Type: lexer.CharClassOpen, Raw: "["},
				{Type: lexer.Not, Raw: "!"},
				{Type: lexer.RangeLow, Raw: "a"},
				{Type: lexer.RangeHigh, Raw: "z"},
				{Type: lexer.CharClassClose, Raw: "]"},
				{Type: lexer.EOF, Raw: ""},
			},
			tree: NewNode(
				KindPattern, nil,
				NewNode(
					KindCharClass,
					CharClass{Not: true},
					NewNode(
						KindRange,
						Range{Low: 'a', High: 'z'},
					),
				),
			),
		},
		{
			// pattern: "[az]",
			tokens: []lexer.Token{
				{Type: lexer.CharClassOpen, Raw: "["},
				{Type: lexer.Text, Raw: "az"},
				{Type: lexer.CharClassClose, Raw: "]"},
				{Type: lexer.EOF, Raw: ""},
			},
			tree: NewNode(
				KindPattern, nil,
				NewNode(
					KindCharClass,
					CharClass{Not: false},
					NewNode(
						KindList,
						List{Chars: "az"},
					),
				),
			),
		},
		{
			// pattern: "{a,z}",
			tokens: []lexer.Token{
				{Type: lexer.TermsOpen, Raw: "{"},
				{Type: lexer.Text, Raw: "a"},
				{Type: lexer.Separator, Raw: ","},
				{Type: lexer.Text, Raw: "z"},
				{Type: lexer.TermsClose, Raw: "}"},
				{Type: lexer.EOF, Raw: ""},
			},
			tree: NewNode(
				KindPattern, nil,
				NewNode(
					KindAnyOf, nil,
					NewNode(
						KindPattern, nil,
						NewNode(KindText, Text{Text: "a"}),
					),
					NewNode(
						KindPattern, nil,
						NewNode(KindText, Text{Text: "z"}),
					),
				),
			),
		},
		{
			// pattern: "/{z,ab}*",
			tokens: []lexer.Token{
				{Type: lexer.Text, Raw: "/"},
				{Type: lexer.TermsOpen, Raw: "{"},
				{Type: lexer.Text, Raw: "z"},
				{Type: lexer.Separator, Raw: ","},
				{Type: lexer.Text, Raw: "ab"},
				{Type: lexer.TermsClose, Raw: "}"},
				{Type: lexer.Any, Raw: "*"},
				{Type: lexer.EOF, Raw: ""},
			},
			tree: NewNode(
				KindPattern, nil,
				NewNode(KindText, Text{Text: "/"}),
				NewNode(
					KindAnyOf, nil,
					NewNode(
						KindPattern, nil,
						NewNode(KindText, Text{Text: "z"}),
					),
					NewNode(
						KindPattern, nil,
						NewNode(KindText, Text{Text: "ab"}),
					),
				),
				NewNode(KindAny, nil),
			),
		},
		{
			// pattern: "{a,{x,y},?,[a-z],[!qwe]}",
			tokens: []lexer.Token{
				{Type: lexer.TermsOpen, Raw: "{"},
				{Type: lexer.Text, Raw: "a"},
				{Type: lexer.Separator, Raw: ","},
				{Type: lexer.TermsOpen, Raw: "{"},
				{Type: lexer.Text, Raw: "x"},
				{Type: lexer.Separator, Raw: ","},
				{Type: lexer.Text, Raw: "y"},
				{Type: lexer.TermsClose, Raw: "}"},
				{Type: lexer.Separator, Raw: ","},
				{Type: lexer.Single, Raw: "?"},
				{Type: lexer.Separator, Raw: ","},
				{Type: lexer.CharClassOpen, Raw: "["},
				{Type: lexer.RangeLow, Raw: "a"},
				{Type: lexer.RangeHigh, Raw: "z"},
				{Type: lexer.CharClassClose, Raw: "]"},
				{Type: lexer.Separator, Raw: ","},
				{Type: lexer.CharClassOpen, Raw: "["},
				{Type: lexer.Not, Raw: "!"},
				{Type: lexer.Text, Raw: "qwe"},
				{Type: lexer.CharClassClose, Raw: "]"},
				{Type: lexer.TermsClose, Raw: "}"},
				{Type: lexer.EOF, Raw: ""},
			},
			tree: NewNode(
				KindPattern, nil,
				NewNode(
					KindAnyOf, nil,
					NewNode(
						KindPattern, nil,
						NewNode(KindText, Text{Text: "a"}),
					),
					NewNode(
						KindPattern, nil,
						NewNode(
							KindAnyOf, nil,
							NewNode(
								KindPattern, nil,
								NewNode(KindText, Text{Text: "x"}),
							),
							NewNode(
								KindPattern, nil,
								NewNode(KindText, Text{Text: "y"}),
							),
						),
					),
					NewNode(
						KindPattern, nil,
						NewNode(KindSingle, nil),
					),
					NewNode(
						KindPattern, nil,
						NewNode(
							KindCharClass,
							CharClass{Not: false},
							NewNode(
								KindRange,
								Range{Low: 'a', High: 'z'},
							),
						),
					),
					NewNode(
						KindPattern, nil,
						NewNode(
							KindCharClass,
							CharClass{Not: true},
							NewNode(
								KindList,
								List{Chars: "qwe"},
							),
						),
					),
				),
			),
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			lexer := &stubLexer{tokens: tt.tokens}
			result, err := Parse(lexer)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			if !reflect.DeepEqual(tt.tree, result) {
				t.Errorf("Parse():\nwant:\t%s\ngot: \t%s\n", tt.tree, result)
			}
		})
	}
}
