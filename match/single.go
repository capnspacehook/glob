package match

import (
	"fmt"
	"unicode/utf8"

	"github.com/capnspacehook/glob/util/runes"
)

// Single matches any single non-separator character; '?'.
type Single struct {
	Separators []rune
}

func NewSingle(s []rune) Single {
	return Single{s}
}

func (self Single) Match(s string) bool {
	if s == "" {
		return false
	}

	r, w := utf8.DecodeRuneInString(s)
	if len(s) > w {
		return false
	}

	return runes.IndexRune(self.Separators, r) == -1
}

func (self Single) Len() int {
	return lenOne
}

func (self Single) Index(s string) (int, []int) {
	for i, r := range s {
		if runes.IndexRune(self.Separators, r) == -1 {
			// use the actual byte width consumed rather than
			// utf8.RuneLen(r): for invalid UTF-8 bytes range yields
			// utf8.RuneError (RuneLen 3) but only advances one byte.
			_, w := utf8.DecodeRuneInString(s[i:])
			return i, segmentsByRuneLength[w]
		}
	}

	return -1, nil
}

func (self Single) String() string {
	return fmt.Sprintf("<single:![%s]>", string(self.Separators))
}
