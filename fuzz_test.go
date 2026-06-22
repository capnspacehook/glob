package glob

import (
	"strings"
	"testing"
)

func FuzzGlobCompileMatch(f *testing.F) {
	for _, tt := range tests {
		var delimiterSet bool
		var delimiter rune
		if len(tt.delimiters) > 0 {
			delimiter = tt.delimiters[0]
			delimiterSet = true
		}

		f.Add(tt.pattern, tt.match, delimiterSet, delimiter)
	}

	f.Fuzz(func(t *testing.T, pattern, match string, delimiterSet bool, delimiter rune) {
		// prevent the fuzzer from creating insanely long patterns,
		// match strings, or creating a pattern that wil take a very
		// long time to match and hanging
		if len(pattern) > 64 || len(match) > 64 {
			t.SkipNow()
		}
		if strings.Count(pattern, "*") > 7 {
			t.SkipNow()
		}

		var (
			g   Glob
			err error
		)
		if delimiterSet {
			g, err = Compile(pattern, delimiter)
		} else {
			g, err = Compile(pattern)
		}
		if err != nil {
			t.SkipNow()
		}

		matched1 := g.Match(match)
		matched2 := g.Match(match)

		if matched1 != matched2 {
			t.Fatal("consecutive match disagreement")
		}
	})
}
