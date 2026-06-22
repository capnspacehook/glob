package match

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Prefix matches any string with a given prefix; ex 'abc*'.
type Prefix struct {
	Prefix string
}

func NewPrefix(p string) Prefix {
	return Prefix{p}
}

func (self Prefix) Index(s string) (int, []int) {
	idx := strings.Index(s, self.Prefix)
	if idx == -1 {
		return -1, nil
	}

	length := len(self.Prefix)
	var sub string
	if len(s) > idx+length {
		sub = s[idx+length:]
	} else {
		sub = ""
	}

	segments := acquireSegments(len(sub) + 1)
	segments = append(segments, length)
	for i := range sub {
		// use the actual byte width consumed rather than utf8.RuneLen(r):
		// for invalid UTF-8 bytes range yields utf8.RuneError (RuneLen 3)
		// but only advances one byte.
		_, w := utf8.DecodeRuneInString(sub[i:])
		segments = append(segments, length+i+w)
	}

	return idx, segments
}

func (self Prefix) Len() int {
	return lenNo
}

func (self Prefix) Match(s string) bool {
	return strings.HasPrefix(s, self.Prefix)
}

func (self Prefix) String() string {
	return fmt.Sprintf("<prefix:%s>", self.Prefix)
}
