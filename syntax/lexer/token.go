package lexer

import "fmt"

type TokenType int

const (
	EOF TokenType = iota
	Error
	Text
	Char
	Any
	Super
	Single
	Not
	Separator
	ListOpen
	ListClose
	RangeLow
	RangeBetween
	RangeHigh
	TermsOpen
	TermsClose
)

func (tt TokenType) String() string {
	switch tt {
	case EOF:
		return "eof"

	case Error:
		return "error"

	case Text:
		return "text"

	case Char:
		return "char"

	case Any:
		return "any"

	case Super:
		return "super"

	case Single:
		return "single"

	case Not:
		return "not"

	case Separator:
		return "separator"

	case ListOpen:
		return "list_open"

	case ListClose:
		return "list_close"

	case RangeLow:
		return "range_low"

	case RangeHigh:
		return "range_high"

	case RangeBetween:
		return "range_between"

	case TermsOpen:
		return "terms_open"

	case TermsClose:
		return "terms_close"

	default:
		return "undef"
	}
}

type Token struct {
	Type TokenType
	Raw  string
}

func (t Token) String() string {
	return fmt.Sprintf("%v<%q>", t.Type, t.Raw)
}
