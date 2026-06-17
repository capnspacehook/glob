package lexer

import (
	"errors"
	"slices"
	"unicode/utf8"

	"github.com/gobwas/glob/util/runes"
)

const (
	charAny          = '*'
	charComma        = ','
	charSingle       = '?'
	charEscape       = '\\'
	charRangeOpen    = '['
	charRangeClose   = ']'
	charTermsOpen    = '{'
	charTermsClose   = '}'
	charRangeNot     = '!'
	charRangeBetween = '-'
)

var specials = []byte{
	charAny,
	charSingle,
	charEscape,
	charRangeOpen,
	charRangeClose,
	charTermsOpen,
	charTermsClose,
}

func Special(c byte) bool {
	return slices.Contains(specials, c)
}

type tokens []Token

func (t *tokens) shift() Token {
	ret := (*t)[0]
	*t = slices.Delete(*t, 0, 1)
	return ret
}

func (t *tokens) push(v Token) {
	*t = append(*t, v)
}

func (t *tokens) empty() bool {
	return len(*t) == 0
}

const eof rune = -1

type lexer struct {
	data string
	pos  int
	err  error

	tokens     tokens
	termsLevel int

	lastRune     rune
	lastRuneSize int
	hasRune      bool
}

func NewLexer(source string) *lexer {
	l := &lexer{
		data:   source,
		tokens: tokens(make([]Token, 0, 4)),
	}
	return l
}

func (l *lexer) Next() Token {
	if l.err != nil {
		return Token{Error, l.err.Error()}
	}
	if !l.tokens.empty() {
		return l.tokens.shift()
	}

	l.readItem()
	return l.Next()
}

func (l *lexer) peek() (rune, int) {
	return l.peekAt(l.pos)
}

func (l *lexer) peekAt(pos int) (rune, int) {
	if pos >= len(l.data) {
		return eof, 0
	}

	r, w := utf8.DecodeRuneInString(l.data[pos:])
	if r == utf8.RuneError && w <= 1 {
		l.error("could not read rune")
		r = eof
		w = 0
	}

	return r, w
}

func (l *lexer) read() rune {
	if l.hasRune {
		l.hasRune = false
		l.seek(l.lastRuneSize)
		return l.lastRune
	}

	r, s := l.peek()
	l.seek(s)

	l.lastRune = r
	l.lastRuneSize = s

	return r
}

func (l *lexer) seek(w int) {
	l.pos += w
}

func (l *lexer) unRead() {
	if l.hasRune {
		l.error("could not unread rune")
		return
	}
	l.seek(-l.lastRuneSize)
	l.hasRune = true
}

func (l *lexer) error(f string) {
	l.err = errors.New(f)
}

func (l *lexer) inTerms() bool {
	return l.termsLevel > 0
}

func (l *lexer) termsEnter() {
	l.termsLevel++
}

func (l *lexer) termsLeave() {
	l.termsLevel--
}

var (
	inTextBreakers  = []rune{charSingle, charAny, charRangeOpen, charTermsOpen}
	inTermsBreakers = append(inTextBreakers, charTermsClose, charComma)
)

func (l *lexer) readItem() {
	r := l.read()
	switch {
	case r == eof:
		l.tokens.push(Token{EOF, ""})
	case r == charTermsOpen:
		l.termsEnter()
		l.tokens.push(Token{TermsOpen, string(r)})
	case r == charComma && l.inTerms():
		l.tokens.push(Token{Separator, string(r)})
	case r == charTermsClose && l.inTerms():
		l.tokens.push(Token{TermsClose, string(r)})
		l.termsLeave()
	case r == charRangeOpen:
		l.tokens.push(Token{RangeOpen, string(r)})
		l.readRange()
	case r == charSingle:
		l.tokens.push(Token{Single, string(r)})
	case r == charAny:
		if l.read() == charAny {
			l.tokens.push(Token{Super, string(r) + string(r)})
		} else {
			l.unRead()
			l.tokens.push(Token{Any, string(r)})
		}
	default:
		l.unRead()

		var breakers []rune
		if l.inTerms() {
			breakers = inTermsBreakers
		} else {
			breakers = inTextBreakers
		}
		l.readText(breakers)
	}
}

func (l *lexer) readRange() {
	var (
		wantHi    bool
		wantClose bool
		seenNot   bool
	)

	for {
		r := l.read()
		if r == eof {
			l.error("unexpected end of input")
			return
		}

		if wantClose {
			if r != charRangeClose {
				l.error("expected close range character")
			} else {
				l.tokens.push(Token{RangeClose, string(r)})
			}
			return
		}

		if wantHi {
			// the hi bound may be escaped, ex [a-\]]
			if r == charEscape {
				r = l.read()
				if r == eof {
					l.error("unexpected end of input")
					return
				}
			}
			l.tokens.push(Token{RangeHi, string(r)})
			wantClose = true
			continue
		}

		if !seenNot && r == charRangeNot {
			l.tokens.push(Token{Not, string(r)})
			seenNot = true
			continue
		}

		if r == charEscape {
			// the lo bound is escaped, e.g. [\b-\a]. The escaped character
			// is the real lo bound, so look past it for the range separator
			esc, ew := l.peek()
			if esc == eof {
				l.error("unexpected end of input")
				return
			}
			if n, w := l.peekAt(l.pos + ew); n == charRangeBetween {
				l.seek(ew) // consume the escaped lo character
				l.seek(w)  // consume the range separator
				l.tokens.push(Token{RangeLo, string(esc)})
				l.tokens.push(Token{RangeBetween, string(n)})
				wantHi = true
				continue
			}
			// not a range; rewind to the escape and let fetchText handle it
			// as part of a character list
			l.unRead()
			l.readText([]rune{charRangeClose})
			wantClose = true
			continue
		}

		if n, w := l.peek(); n == charRangeBetween {
			l.seek(w)
			l.tokens.push(Token{RangeLo, string(r)})
			l.tokens.push(Token{RangeBetween, string(n)})
			wantHi = true
			continue
		}

		l.unRead() // unread first peek and fetch as text
		l.readText([]rune{charRangeClose})
		wantClose = true
	}
}

func (l *lexer) readText(breakers []rune) {
	var data []rune
	var escaped bool

	for {
		r := l.read()
		if r == eof {
			break
		}

		if !escaped {
			if r == charEscape {
				escaped = true
				continue
			}

			if runes.IndexRune(breakers, r) != -1 {
				l.unRead()
				break
			}
		}

		escaped = false
		data = append(data, r)
	}

	if len(data) > 0 {
		l.tokens.push(Token{Text, string(data)})
	}
}
