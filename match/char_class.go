package match

import (
	"slices"
	"strings"
	"unicode/utf8"
)

type CharClass struct {
	Not    bool
	List   []rune
	Ranges []CharRange
}

type CharRange struct {
	Low  rune
	High rune
}

// TODO:
// - optimize by flattening ranges if they al have a size of 1
// ex. [a-b0-1] can be flattened to [ab01]
// - optimize by combining ranges if they overlap
// ex. [a-cb-d] can be simplified to [a-d]
func NewCharClass(not bool, list []rune, ranges []CharRange) CharClass {
	return CharClass{not, list, ranges}
}

func (c CharClass) Match(s string) bool {
	if s == "" {
		return false
	}

	r, w := utf8.DecodeRuneInString(s)
	if len(s) > w {
		return false
	}

	return c.matches(r) != c.Not
}

func (c CharClass) Index(s string) (int, []int) {
	for i, r := range s {
		if c.matches(r) != c.Not {
			return i, segmentsByRuneLength[utf8.RuneLen(r)]
		}
	}

	return -1, nil
}

func (c CharClass) matches(r rune) bool {
	if slices.Contains(c.List, r) {
		return true
	}
	for _, rg := range c.Ranges {
		if r >= rg.Low && r <= rg.High {
			return true
		}
	}

	return false
}

func (c CharClass) Len() int {
	return lenOne
}

func (c CharClass) String() string {
	var sb strings.Builder
	sb.WriteString("<char_class:")
	sb.WriteByte('[')
	if c.Not {
		sb.WriteByte('!')
	}
	sb.WriteString(string(c.List))

	for _, rg := range c.Ranges {
		sb.WriteRune(rg.Low)
		sb.WriteByte('-')
		sb.WriteRune(rg.High)
	}

	sb.WriteString("]>")
	return sb.String()
}
