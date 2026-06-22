package match

import (
	"fmt"
	"strings"
)

// Suffix matches any string with a given suffix; ex '*abc'.
type Suffix struct {
	Suffix string
}

func NewSuffix(s string) Suffix {
	return Suffix{s}
}

func (self Suffix) Len() int {
	return lenNo
}

func (self Suffix) Match(s string) bool {
	return strings.HasSuffix(s, self.Suffix)
}

func (self Suffix) Index(s string) (int, []int) {
	idx := strings.Index(s, self.Suffix)
	if idx == -1 {
		return -1, nil
	}

	// '**' matches anything (separators included), so the match always starts
	// at 0 and every occurrence of the suffix is a valid match length; report
	// them all so callers (e.g. BTree) can try each one.
	var segments []int
	for occ := idx; occ != -1; {
		segments = append(segments, occ+len(self.Suffix))
		rel := strings.Index(s[occ+1:], self.Suffix)
		if rel == -1 {
			break
		}
		occ = occ + 1 + rel
	}

	return 0, segments
}

func (self Suffix) String() string {
	return fmt.Sprintf("<suffix:%s>", self.Suffix)
}
