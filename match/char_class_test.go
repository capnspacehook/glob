package match

import (
	"strings"
	"testing"

	"pgregory.net/rapid"
)

func TestCharClass(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		not := rapid.Bool().Draw(t, "not")
		list := rapid.StringN(1, 8, -1).Draw(t, "list")

		var ranges []CharRange
		numRanges := rapid.IntRange(0, 3).Draw(t, "numRanges")
		for range numRanges {
			low := rapid.Rune().Draw(t, "low")
			high := rapid.Rune().Filter(func(r rune) bool {
				return r > low
			}).Draw(t, "high")

			ranges = append(ranges, CharRange{
				Low:  low,
				High: high,
			})
		}

		input := string(rapid.Rune().Draw(t, "input"))

		cc := NewCharClass(not, []rune(list), ranges)
		matched := cc.Match(input)

		contains := strings.ContainsAny(input, list)
		var inRange bool
		for _, r := range input {
			for _, rg := range ranges {
				if r >= rg.Low && r <= rg.High {
					inRange = true
					break
				}
			}
		}

		inClass := contains || inRange
		shouldMatch := inClass != not
		if matched != shouldMatch {
			t.Logf("input=%s charClass=%s", input, cc)
			t.Fatalf("expected match=%t, got %t", shouldMatch, matched)
		}
	})
}
