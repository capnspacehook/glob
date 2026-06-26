package lexer

import (
	"errors"
	"fmt"
	"slices"
	"unicode/utf8"

	"github.com/capnspacehook/glob/util/runes"
)

const (
	charAny          = '*'
	charComma        = ','
	charSingle       = '?'
	charEscape       = '\\'
	charClassOpen    = '['
	charClassClose   = ']'
	charTermsOpen    = '{'
	charTermsClose   = '}'
	charListNot      = '!'
	charRangeBetween = '-'
)

var (
	specials = []rune{
		charAny,
		charSingle,
		charEscape,
		charClassOpen,
		charClassClose,
		charTermsOpen,
		charTermsClose,
	}

	inTextBreakers  = []rune{charSingle, charAny, charClassOpen, charTermsOpen}
	inTermsBreakers = append(inTextBreakers, charTermsClose, charComma)
)

func IsSpecial(r rune) bool {
	return slices.Contains(specials, r)
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

func (l *lexer) errorf(f string, args ...any) {
	l.err = fmt.Errorf(f, args...)
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

func (l *lexer) readItem() {
	r := l.read()
	switch {
	case r == eof:
		if l.termsLevel > 0 {
			l.error("expected brace close")
			return
		}

		l.tokens.push(Token{EOF, ""})
	case r == charTermsOpen:
		l.termsEnter()
		l.tokens.push(Token{TermsOpen, string(r)})
	case r == charComma && l.inTerms():
		l.tokens.push(Token{Separator, string(r)})
	case r == charTermsClose && l.inTerms():
		l.tokens.push(Token{TermsClose, string(r)})
		l.termsLeave()
	case r == charClassOpen:
		l.tokens.push(Token{CharClassOpen, string(r)})
		l.readCharClass()
	case r == charClassClose:
		l.error("unexpected list close")
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

func (l *lexer) readCharClass() {
	var (
		chars []rune
		first = true
	)

	for {
		r := l.read()
		if r == eof {
			l.error("unexpected end of input")
			return
		}

		if r == charClassClose {
			if len(chars) > 0 {
				l.tokens.push(Token{Text, string(chars)})
			}
			l.tokens.push(Token{CharClassClose, string(r)})
			return
		}

		if first {
			first = false
			if r == charListNot {
				l.tokens.push(Token{Not, string(r)})
				continue
			}
		}

		var escaped bool
		if r == charEscape {
			escaped = true
			r = l.read()
			if r == eof {
				l.error("unexpected end of input")
				return
			}
		}

		if !escaped && IsSpecial(r) {
			l.errorf("unexpected special character %c in character class", r)
			return
		}

		if sep, sepWidth := l.peek(); sep == charRangeBetween {
			peekHi, _ := l.peekAt(l.pos + sepWidth)
			if peekHi != eof && peekHi != charClassClose {
				l.seek(sepWidth)
				hi := l.read()

				var hiEscaped bool
				if hi == charEscape {
					hiEscaped = true
					hi = l.read()
				}
				if hi == eof {
					l.error("unexpected end of input")
					return
				}
				if !hiEscaped && IsSpecial(hi) {
					l.errorf("unexpected special character %c in character class", hi)
					return
				}

				if len(chars) > 0 {
					l.tokens.push(Token{Text, string(chars)})
					chars = chars[:0]
				}

				l.tokens.push(Token{RangeLow, string(r)})
				l.tokens.push(Token{RangeBetween, "-"})
				l.tokens.push(Token{RangeHigh, string(hi)})
				continue
			}
		}

		chars = append(chars, r)
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
