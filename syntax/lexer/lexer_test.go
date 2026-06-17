package lexer

import (
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"pgregory.net/rapid"
)

func TestLexerText(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		input := rapid.String().Filter(func(s string) bool {
			return !strings.ContainsAny(s, string(specials)) && utf8.ValidString(s)
		}).Draw(t, "input")

		l := NewLexer(input)
		tok := l.Next()
		if input == "" {
			if tok.Type != EOF {
				t.Fatalf("expected eof token, got %s", tok)
			}
			return
		}

		if tok.Type != Text {
			t.Errorf("expected text token, got %s", tok)
		}
		if tok.Raw != input {
			t.Errorf("expected %q, got %q", input, tok.Raw)
		}

		tok = l.Next()
		if tok.Type != EOF {
			t.Errorf("expected eof token, got %s", tok)
		}
	})
}

func assertToken(t *rapid.T, tok Token, expType TokenType, expRaw string) {
	t.Helper()

	if tok.Type != expType {
		t.Errorf("expected %s token, got %s", expType, tok.Type)
	}
	if tok.Raw != expRaw {
		t.Errorf("expected %s, got %s", expRaw, tok.Raw)
	}
}

var toQuote = []rune{'[', ']', '-', '{', '}', '\\'}

func quoteStr(s string) string {
	var quoted string
	for _, r := range s {
		if slices.Contains(toQuote, r) {
			quoted += `\` + string(r)
			continue
		}
		quoted += string(r)
	}

	return quoted
}

func TestLexerTerms(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		isRange := rapid.Bool().Draw(t, "isRange")
		input := rapid.String().Filter(func(s string) bool {
			return true
		}).Draw(t, "input")
		if isRange {
			input = "[" + quoteStr(input) + "]"
		} else {
			input = "{" + quoteStr(input) + "}"
		}
		t.Logf("input=%q", input)

		l := NewLexer(input)
		tok := l.Next()

		if isRange {
			assertToken(t, tok, RangeOpen, string(charRangeOpen))
		} else {
			assertToken(t, tok, TermsOpen, string(charTermsOpen))
		}

		var lastTok Token
		for {
			tok = l.Next()
			if tok.Type == EOF || tok.Type == Error {
				break
			}
			lastTok = tok
		}

		if isRange {
			assertToken(t, lastTok, RangeClose, string(charRangeClose))
		} else {
			assertToken(t, lastTok, TermsClose, string(charTermsClose))
		}
	})
}

func TestLexer(t *testing.T) {
	tests := []struct {
		pattern string
		items   []Token
	}{
		// {
		// 	pattern: "[*a-z]",
		// 	items: []Token{
		// 		{RangeOpen, "["},
		// 		{Text, "*"},
		// 		{RangeLo, "a"},
		// 		{RangeBetween, "-"},
		// 		{RangeHi, "z"},
		// 		{RangeClose, "]"},
		// 		{EOF, ""},
		// 	},
		// },
		{
			pattern: "",
			items: []Token{
				{EOF, ""},
			},
		},
		{
			pattern: "hello",
			items: []Token{
				{Text, "hello"},
				{EOF, ""},
			},
		},
		{
			pattern: "/{rate,[0-9]]}*",
			items: []Token{
				{Text, "/"},
				{TermsOpen, "{"},
				{Text, "rate"},
				{Separator, ","},
				{RangeOpen, "["},
				{RangeLo, "0"},
				{RangeBetween, "-"},
				{RangeHi, "9"},
				{RangeClose, "]"},
				{Text, "]"},
				{TermsClose, "}"},
				{Any, "*"},
				{EOF, ""},
			},
		},
		{
			pattern: "hello,world",
			items: []Token{
				{Text, "hello,world"},
				{EOF, ""},
			},
		},
		{
			pattern: "hello\\,world",
			items: []Token{
				{Text, "hello,world"},
				{EOF, ""},
			},
		},
		{
			pattern: "hello\\{world",
			items: []Token{
				{Text, "hello{world"},
				{EOF, ""},
			},
		},
		{
			pattern: "hello?",
			items: []Token{
				{Text, "hello"},
				{Single, "?"},
				{EOF, ""},
			},
		},
		{
			pattern: "hellof*",
			items: []Token{
				{Text, "hellof"},
				{Any, "*"},
				{EOF, ""},
			},
		},
		{
			pattern: "hello**",
			items: []Token{
				{Text, "hello"},
				{Super, "**"},
				{EOF, ""},
			},
		},
		{
			pattern: "[日-語]",
			items: []Token{
				{RangeOpen, "["},
				{RangeLo, "日"},
				{RangeBetween, "-"},
				{RangeHi, "語"},
				{RangeClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: "[!日-語]",
			items: []Token{
				{RangeOpen, "["},
				{Not, "!"},
				{RangeLo, "日"},
				{RangeBetween, "-"},
				{RangeHi, "語"},
				{RangeClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: "[日本語]",
			items: []Token{
				{RangeOpen, "["},
				{Text, "日本語"},
				{RangeClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: "[!日本語]",
			items: []Token{
				{RangeOpen, "["},
				{Not, "!"},
				{Text, "日本語"},
				{RangeClose, "]"},
				{EOF, ""},
			},
		},
		{
			// Escaped range bounds: the escape means "the following
			// character literally", so [\b-\a] is the (reversed) range b-a,
			// not the character list {b, -, a}.
			pattern: `[\b-\a]`,
			items: []Token{
				{RangeOpen, "["},
				{RangeLo, "b"},
				{RangeBetween, "-"},
				{RangeHi, "a"},
				{RangeClose, "]"},
				{EOF, ""},
			},
		},
		{
			// An escaped hi bound, e.g. a special character.
			pattern: `[a-\]]`,
			items: []Token{
				{RangeOpen, "["},
				{RangeLo, "a"},
				{RangeBetween, "-"},
				{RangeHi, "]"},
				{RangeClose, "]"},
				{EOF, ""},
			},
		},
		{
			// An escaped char not followed by '-' is still a character list.
			pattern: `[\*abc]`,
			items: []Token{
				{RangeOpen, "["},
				{Text, "*abc"},
				{RangeClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: "{a,b}",
			items: []Token{
				{TermsOpen, "{"},
				{Text, "a"},
				{Separator, ","},
				{Text, "b"},
				{TermsClose, "}"},
				{EOF, ""},
			},
		},
		{
			pattern: "/{z,ab}*",
			items: []Token{
				{Text, "/"},
				{TermsOpen, "{"},
				{Text, "z"},
				{Separator, ","},
				{Text, "ab"},
				{TermsClose, "}"},
				{Any, "*"},
				{EOF, ""},
			},
		},
		{
			pattern: "{[!日-語],*,?,{a,b,\\c}}",
			items: []Token{
				{TermsOpen, "{"},
				{RangeOpen, "["},
				{Not, "!"},
				{RangeLo, "日"},
				{RangeBetween, "-"},
				{RangeHi, "語"},
				{RangeClose, "]"},
				{Separator, ","},
				{Any, "*"},
				{Separator, ","},
				{Single, "?"},
				{Separator, ","},
				{TermsOpen, "{"},
				{Text, "a"},
				{Separator, ","},
				{Text, "b"},
				{Separator, ","},
				{Text, "c"},
				{TermsClose, "}"},
				{TermsClose, "}"},
				{EOF, ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			lexer := NewLexer(tt.pattern)
			for i, exp := range tt.items {
				act := lexer.Next()
				if act.Type != exp.Type {
					t.Errorf("%q: wrong %d-th item type: want: %q; got: %q\n\t(%s vs %s)", tt.pattern, i, exp.Type, act.Type, exp, act)
				}
				if act.Raw != exp.Raw {
					t.Errorf("%q: wrong %d-th item contents: want: %q; got: %q\n\t(%s vs %s)", tt.pattern, i, exp.Raw, act.Raw, exp, act)
				}
			}
		})
	}
}
