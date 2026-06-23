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

func NewCharClass(not bool, list []rune, ranges []CharRange) CharClass {
	slices.Sort(list)
	deduped := slices.Compact(list)

	// optimize ranges by combining overlapping ranges
	expandedRanges := ranges
	for _, r := range ranges {
		for i := range expandedRanges {
			if r.High == expandedRanges[i].Low {
				expandedRanges[i].Low = r.Low
			} else if r.Low == expandedRanges[i].High {
				expandedRanges[i].High = r.High
			} else if r.Low < expandedRanges[i].Low && r.High == expandedRanges[i].High {
				expandedRanges[i].Low = r.Low
			} else if r.Low == expandedRanges[i].Low && r.High > expandedRanges[i].High {
				expandedRanges[i].High = r.High
			}
		}
	}

	uniqueRanges := make([]CharRange, 0, len(expandedRanges))
	for i := range expandedRanges {
		for k, r := range expandedRanges {
			if i == k {
				continue
			}

			if containsRange(expandedRanges[i:], r) && !containsRange(uniqueRanges, r) {
				uniqueRanges = append(uniqueRanges, r)
				break
			}
		}
	}

	// only keep characters not in ranges
	uniqueChars := make([]rune, 0, len(deduped))
	for _, c := range deduped {
		contained := false
		for _, r := range expandedRanges {
			if c >= r.Low && c <= r.High {
				contained = true
				break
			}
		}

		if !contained {
			uniqueChars = append(uniqueChars, c)
		}
	}
	uniqueChars = slices.Clip(uniqueChars)

	return CharClass{not, uniqueChars, expandedRanges}
}

func containsRange(ranges []CharRange, r CharRange) bool {
	return slices.ContainsFunc(ranges, func(cr CharRange) bool {
		return r.Low >= cr.Low && r.High <= cr.High
	})
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
			// use the actual byte width consumed rather than
			// utf8.RuneLen(r): for invalid UTF-8 bytes range yields
			// utf8.RuneError (RuneLen 3) but only advances one byte.
			_, w := utf8.DecodeRuneInString(s[i:])
			return i, segmentsByRuneLength[w]
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
