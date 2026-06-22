package match

import (
	"fmt"
	"unicode/utf8"
)

// Max matches any string that has at most the given length; used for
// length upper-bounds in compiled patterns.
type Max struct {
	Limit int
}

func NewMax(l int) Max {
	return Max{l}
}

func (self Max) Match(s string) bool {
	var l int
	for range s {
		l += 1
		if l > self.Limit {
			return false
		}
	}

	return true
}

func (self Max) Index(s string) (int, []int) {
	segments := acquireSegments(self.Limit + 1)
	segments = append(segments, 0)
	var count int
	for i := range s {
		count++
		if count > self.Limit {
			break
		}
		// use the actual byte width consumed rather than utf8.RuneLen(r):
		// for invalid UTF-8 bytes range yields utf8.RuneError (RuneLen 3)
		// but only advances one byte.
		_, w := utf8.DecodeRuneInString(s[i:])
		segments = append(segments, i+w)
	}

	return 0, segments
}

func (self Max) Len() int {
	return lenNo
}

func (self Max) String() string {
	return fmt.Sprintf("<max:%d>", self.Limit)
}
