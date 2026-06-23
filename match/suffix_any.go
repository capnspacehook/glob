package match

import (
	"fmt"
	"strings"

	sutil "github.com/capnspacehook/glob/util/strings"
)

// SuffixAny any matches a string with a given suffix that isn't followed
// by separators; ex '*abc'.
type SuffixAny struct {
	Suffix     string
	Separators []rune
}

func NewSuffixAny(s string, sep []rune) SuffixAny {
	return SuffixAny{s, sep}
}

func (self SuffixAny) Index(s string) (int, []int) {
	idx := strings.Index(s, self.Suffix)
	if idx == -1 {
		return -1, nil
	}

	// '*' cannot cross a separator, so the match starts right after the last
	// separator preceding the first suffix occurrence.
	i := sutil.LastIndexAnyRunes(s[:idx], self.Separators) + 1

	// The part matched by '*' (everything between i and the suffix) must not
	// contain a separator, so a suffix occurrence is only reachable from i if
	// it starts at or before the first separator at/after i. Note the suffix
	// literal itself may contain separators, ex '*.google.'.
	sepLimit := len(s)
	if rel := sutil.IndexAnyRunes(s[i:], self.Separators); rel != -1 {
		sepLimit = i + rel
	}

	// Report every reachable suffix occurrence so callers (e.g. BTree) can try
	// each possible match length.
	var segments []int
	for occ := idx; occ != -1 && occ <= sepLimit; {
		segments = append(segments, occ+len(self.Suffix)-i)
		rel := strings.Index(s[occ+1:], self.Suffix)
		if rel == -1 {
			break
		}
		occ = occ + 1 + rel
	}

	return i, segments
}

func (self SuffixAny) Len() int {
	return lenNo
}

func (self SuffixAny) Match(s string) bool {
	if !strings.HasSuffix(s, self.Suffix) {
		return false
	}
	return sutil.IndexAnyRunes(s[:len(s)-len(self.Suffix)], self.Separators) == -1
}

func (self SuffixAny) String() string {
	return fmt.Sprintf("<suffix_any:![%s]%s>", string(self.Separators), self.Suffix)
}
