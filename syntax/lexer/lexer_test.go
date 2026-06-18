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
		t.Fatalf("expected %s token, got %s", expType, tok.Type)
	}
	if tok.Raw != expRaw {
		t.Fatalf("expected %s, got %s", expRaw, tok.Raw)
	}
}

func quoteStr(s string, toQuote []rune) string {
	var quoted strings.Builder
	for _, r := range s {
		if slices.Contains(toQuote, r) {
			quoted.WriteByte('\\')
			quoted.WriteRune(r)
			continue
		}
		quoted.WriteString(string(r))
	}

	return quoted.String()
}

func TestLexerTerms(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		isRange := rapid.Bool().Draw(t, "isRange")
		input := rapid.String().Draw(t, "input")
		if isRange {
			input = "[" + quoteStr(input, specials) + "]"
		} else {
			input = "{" + quoteStr(input, specials) + "}"
		}
		t.Logf("input=%s", input)

		l := NewLexer(input)
		tok := l.Next()

		if isRange {
			assertToken(t, tok, ListOpen, string(charListOpen))
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
			assertToken(t, lastTok, ListClose, string(charListClose))
		} else {
			assertToken(t, lastTok, TermsClose, string(charTermsClose))
		}
	})
}

var (
	termRunes = []rune{
		charListOpen,
		charListClose,
		charTermsOpen,
		charTermsClose,
		charEscape,
	}
	listToQuote = append(specials, charListNot, charRangeBetween)
)

func writeList(t *rapid.T, sb *strings.Builder) (ranges int, negated bool) {
	length := rapid.IntRange(1, 16).Draw(t, "length")

	sb.WriteByte('[')
	if rapid.Bool().Draw(t, "negated") {
		sb.WriteByte('!')
		negated = true
	}

	for range length {
		// add a range 33% of the time
		choice := rapid.IntRange(0, 2).Draw(t, "choice")
		if choice == 0 {
			writeRune(t, sb, listToQuote)
		} else {
			ranges++
			writeRune(t, sb, listToQuote)
			sb.WriteByte('-')
			writeRune(t, sb, listToQuote)
		}
	}
	sb.WriteByte(']')

	return
}

func writeRune(t *rapid.T, sb *strings.Builder, toQuote []rune) {
	r := rapid.Rune().Draw(t, "rune")
	if slices.Contains(toQuote, r) {
		sb.WriteByte('\\')
		sb.WriteRune(r)
		return
	}
	sb.WriteRune(r)
}

func TestLexerLists(t *testing.T) {
	rapid.Check(t, testLexerLists)
}

func FuzzLexerLists(f *testing.F) {
	f.Fuzz(rapid.MakeFuzz(testLexerLists))
}

func testLexerLists(t *rapid.T) {
	rapid.Check(t, func(t *rapid.T) {
		var sb strings.Builder

		prefixLen := rapid.IntRange(0, 8).Draw(t, "prefixLen")
		for range prefixLen {
			writeRune(t, &sb, termRunes)
		}

		var numRanges int
		var numNots int
		numLists := rapid.IntRange(1, 3).Draw(t, "numLists")
		for range numLists {
			n, negated := writeList(t, &sb)
			numRanges += n
			if negated {
				numNots++
			}

			fillerLen := rapid.IntRange(0, 4).Draw(t, "fillerLen")
			for range fillerLen {
				writeRune(t, &sb, termRunes)
			}
		}

		suffixLen := rapid.IntRange(0, 8).Draw(t, "suffixLen")
		for range suffixLen {
			writeRune(t, &sb, termRunes)
		}

		input := sb.String()
		t.Logf("input=%s", input)

		var (
			listOpens     int
			listNots      int
			rangeLows     int
			rangeBetweens int
			rangeHighs    int
			listCloses    int
		)
		l := NewLexer(input)
		for {
			tok := l.Next()
			if tok.Type == EOF {
				break
			}

			switch tok.Type {
			case Error:
				t.Fatalf("unexpected error: %s", tok.Raw)
			case ListOpen:
				listOpens++
			case Not:
				listNots++
			case RangeLow:
				rangeLows++
			case RangeBetween:
				rangeBetweens++
			case RangeHigh:
				rangeHighs++
			case ListClose:
				listCloses++
			}

			t.Logf("token: %s", tok)
		}

		if listOpens != numLists {
			t.Errorf("expected %d list opens, got %d", numLists, listOpens)
		}
		if listNots != numNots {
			t.Errorf("expected %d list nots, got %d", numNots, listNots)
		}
		if rangeLows != numRanges {
			t.Errorf("expected %d range lows, got %d", numRanges, rangeLows)
		}
		if rangeBetweens != numRanges {
			t.Errorf("expected %d range betweens, got %d", numRanges, rangeBetweens)
		}
		if rangeHighs != numRanges {
			t.Errorf("expected %d range highs, got %d", numRanges, rangeHighs)
		}
		if listCloses != numLists {
			t.Errorf("expected %d list closes, got %d", numLists, listCloses)
		}
	})
}

func TestLexer(t *testing.T) {
	tests := []struct {
		pattern string
		items   []Token
	}{
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
				{ListOpen, "["},
				{RangeLow, "0"},
				{RangeBetween, "-"},
				{RangeHigh, "9"},
				{ListClose, "]"},
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
				{ListOpen, "["},
				{RangeLow, "日"},
				{RangeBetween, "-"},
				{RangeHigh, "語"},
				{ListClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: "[!日-語]",
			items: []Token{
				{ListOpen, "["},
				{Not, "!"},
				{RangeLow, "日"},
				{RangeBetween, "-"},
				{RangeHigh, "語"},
				{ListClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: "[日本語]",
			items: []Token{
				{ListOpen, "["},
				{Text, "日本語"},
				{ListClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: "[!日本語]",
			items: []Token{
				{ListOpen, "["},
				{Not, "!"},
				{Text, "日本語"},
				{ListClose, "]"},
				{EOF, ""},
			},
		},
		{
			// Escaped range bounds: the escape means "the following
			// character literally", so [\b-\a] is the (reversed) range b-a,
			// not the character list {b, -, a}.
			pattern: `[\b-\a]`,
			items: []Token{
				{ListOpen, "["},
				{RangeLow, "b"},
				{RangeBetween, "-"},
				{RangeHigh, "a"},
				{ListClose, "]"},
				{EOF, ""},
			},
		},
		{
			// An escaped hi bound, e.g. a special character.
			pattern: `[a-\]]`,
			items: []Token{
				{ListOpen, "["},
				{RangeLow, "a"},
				{RangeBetween, "-"},
				{RangeHigh, "]"},
				{ListClose, "]"},
				{EOF, ""},
			},
		},
		{
			// An escaped char not followed by '-' is still a character list.
			pattern: `[\*abc]`,
			items: []Token{
				{ListOpen, "["},
				{Text, "*abc"},
				{ListClose, "]"},
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
				{ListOpen, "["},
				{Not, "!"},
				{RangeLow, "日"},
				{RangeBetween, "-"},
				{RangeHigh, "語"},
				{ListClose, "]"},
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
		{
			pattern: `[\--\-]`,
			items: []Token{
				{ListOpen, "["},
				{RangeLow, "-"},
				{RangeBetween, "-"},
				{RangeHigh, "-"},
				{ListClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: `[\*a-z]`,
			items: []Token{
				{ListOpen, "["},
				{Text, "*"},
				{RangeLow, "a"},
				{RangeBetween, "-"},
				{RangeHigh, "z"},
				{ListClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: `[a-z\*]`,
			items: []Token{
				{ListOpen, "["},
				{RangeLow, "a"},
				{RangeBetween, "-"},
				{RangeHigh, "z"},
				{Text, "*"},
				{ListClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: `[\?a-z\*]`,
			items: []Token{
				{ListOpen, "["},
				{Text, "?"},
				{RangeLow, "a"},
				{RangeBetween, "-"},
				{RangeHigh, "z"},
				{Text, "*"},
				{ListClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: `[a-zA-Z]`,
			items: []Token{
				{ListOpen, "["},
				{RangeLow, "a"},
				{RangeBetween, "-"},
				{RangeHigh, "z"},
				{RangeLow, "A"},
				{RangeBetween, "-"},
				{RangeHigh, "Z"},
				{ListClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: `ab[\?a-z\*]cd`,
			items: []Token{
				{Text, "ab"},
				{ListOpen, "["},
				{Text, "?"},
				{RangeLow, "a"},
				{RangeBetween, "-"},
				{RangeHigh, "z"},
				{Text, "*"},
				{ListClose, "]"},
				{Text, "cd"},
				{EOF, ""},
			},
		},
		{
			pattern: `[!a]`,
			items: []Token{
				{ListOpen, "["},
				{Not, "!"},
				{Text, "a"},
				{ListClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: `[\!a]`,
			items: []Token{
				{ListOpen, "["},
				{Text, "!a"},
				{ListClose, "]"},
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
