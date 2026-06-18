package match

import (
	"fmt"
	"unicode/utf8"
)

// Row matches if all of its fixed-length sub-matchers match; used to
// optimize a sequence of fixed-length matchers; ex '[ab]?[h-y]'.
type Row struct {
	Matchers    Matchers
	RunesLength int
}

func NewRow(len int, m ...Matcher) Row {
	return Row{
		Matchers:    Matchers(m),
		RunesLength: len,
	}
}

func (r Row) matchAll(s string) (int, bool) {
	var idx int
	for _, m := range r.Matchers {
		length := m.Len()

		end := idx
		var runes int
		for runes < length && end < len(s) {
			_, width := utf8.DecodeRuneInString(s[end:])
			end += width
			runes++
		}

		if runes < length || !m.Match(s[idx:end]) {
			return 0, false
		}

		idx = end
	}

	return idx, true
}

func (r Row) lenOk(s string) bool {
	var i int
	for range s {
		i++
		if i > r.RunesLength {
			return false
		}
	}
	return r.RunesLength == i
}

func (r Row) Match(s string) bool {
	if !r.lenOk(s) {
		return false
	}
	_, ok := r.matchAll(s)
	return ok
}

func (r Row) Len() (l int) {
	return r.RunesLength
}

func (r Row) Index(s string) (int, []int) {
	for i := 0; i <= len(s); {
		if len(s[i:]) < r.RunesLength {
			break
		}
		if bytesLen, ok := r.matchAll(s[i:]); ok {
			return i, []int{bytesLen}
		}
		if i == len(s) {
			break
		}
		_, width := utf8.DecodeRuneInString(s[i:])
		i += width
	}

	return -1, nil
}

func (r Row) String() string {
	return fmt.Sprintf("<row_%d:[%s]>", r.RunesLength, r.Matchers)
}
