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
	Segments    []int
}

func NewRow(len int, m ...Matcher) Row {
	return Row{
		Matchers:    Matchers(m),
		RunesLength: len,
		Segments:    []int{len},
	}
}

func (self Row) matchAll(s string) bool {
	var idx int
	for _, m := range self.Matchers {
		length := m.Len()

		end := idx
		var runes int
		for runes < length && end < len(s) {
			_, width := utf8.DecodeRuneInString(s[end:])
			end += width
			runes++
		}

		if runes < length || !m.Match(s[idx:end]) {
			return false
		}

		idx = end
	}

	return true
}

func (self Row) lenOk(s string) bool {
	var i int
	for range s {
		i++
		if i > self.RunesLength {
			return false
		}
	}
	return self.RunesLength == i
}

func (self Row) Match(s string) bool {
	return self.lenOk(s) && self.matchAll(s)
}

func (self Row) Len() (l int) {
	return self.RunesLength
}

func (self Row) Index(s string) (int, []int) {
	for i := range s {
		if len(s[i:]) < self.RunesLength {
			break
		}
		if self.matchAll(s[i:]) {
			return i, self.Segments
		}
	}
	return -1, nil
}

func (self Row) String() string {
	return fmt.Sprintf("<row_%d:[%s]>", self.RunesLength, self.Matchers)
}
