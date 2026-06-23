package match

import (
	"strings"
	"testing"

	"pgregory.net/rapid"
)

func TestCharClass(t *testing.T) {
	rapid.Check(t, testCharClass)
}

func FuzzCharClass(f *testing.F) {
	f.Fuzz(rapid.MakeFuzz(testCharClass))
}

func testCharClass(t *rapid.T) {
	not := rapid.Bool().Draw(t, "not")
	list := rapid.StringN(1, 16, -1).Draw(t, "list")

	var ranges []CharRange
	numRanges := rapid.IntRange(0, 8).Draw(t, "numRanges")
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
	if len(cc.List) > len(list) {
		t.Errorf("optimized list grew; from %d to %d", len(list), len(cc.List))
	}
	if len(cc.Ranges) > len(ranges) {
		t.Errorf("optimized ranges grew; from %d to %d", len(ranges), len(cc.Ranges))
	}

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
}
