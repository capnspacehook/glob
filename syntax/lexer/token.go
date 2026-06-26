package lexer

import "fmt"

type TokenType int

const (
	EOF TokenType = iota
	Error
	Text
	Any            // *
	Super          // **
	Single         // ?
	Not            // !
	Separator      // ,
	CharClassOpen  // [
	CharClassClose // ]
	RangeLow
	RangeHigh
	RangeBetween // -
	TermsOpen    // {
	TermsClose   // }
)

func (tt TokenType) String() string {
	switch tt {
	case EOF:
		return "eof"
	case Error:
		return "error"
	case Text:
		return "text"
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
	case CharClassOpen:
		return "char_class_open"
	case CharClassClose:
		return "char_class_close"
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
