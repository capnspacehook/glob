package match

import (
	"fmt"
	"unicode/utf8"
)

// Min matches any string that has at least the given length; used for
// length lower-bounds in compiled patterns.
type Min struct {
	Limit int
}

func NewMin(l int) Min {
	return Min{l}
}

func (self Min) Match(s string) bool {
	var l int
	for range s {
		l += 1
		if l >= self.Limit {
			return true
		}
	}

	return false
}

func (self Min) Index(s string) (int, []int) {
	var count int

	c := len(s) - self.Limit + 1
	if c <= 0 {
		return -1, nil
	}

	segments := acquireSegments(c)
	for i := range s {
		count++
		if count >= self.Limit {
			// use the actual byte width consumed rather than
			// utf8.RuneLen(r): for invalid UTF-8 bytes range yields
			// utf8.RuneError (RuneLen 3) but only advances one byte.
			_, w := utf8.DecodeRuneInString(s[i:])
			segments = append(segments, i+w)
		}
	}

	if len(segments) == 0 {
		return -1, nil
	}

	return 0, segments
}

func (self Min) Len() int {
	return lenNo
}

func (self Min) String() string {
	return fmt.Sprintf("<min:%d>", self.Limit)
}
