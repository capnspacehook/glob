package match

import (
	"fmt"
	"slices"
	"strings"
)

// PrefixSuffix matches any string with a given prefix and suffix;
// ex 'abc*xyz'.
type PrefixSuffix struct {
	Prefix, Suffix string
}

func NewPrefixSuffix(p, s string) PrefixSuffix {
	return PrefixSuffix{p, s}
}

func (self PrefixSuffix) Index(s string) (int, []int) {
	prefixIdx := strings.Index(s, self.Prefix)
	if prefixIdx == -1 {
		return -1, nil
	}

	suffixLen := len(self.Suffix)
	if suffixLen <= 0 {
		return prefixIdx, []int{len(s) - prefixIdx}
	}

	if (len(s) - prefixIdx) <= 0 {
		return -1, nil
	}

	segments := acquireSegments(len(s) - prefixIdx)
	for sub := s[prefixIdx:]; ; {
		suffixIdx := strings.LastIndex(sub, self.Suffix)
		if suffixIdx == -1 {
			break
		}

		// the suffix must not overlap the prefix, ex 'a**a' must not match
		// "a". suffixIdx is relative to the prefix start, so a valid match
		// needs suffixIdx >= len(prefix). suffixIdx only decreases from here,
		// so no later iteration can satisfy this either.
		if suffixIdx < len(self.Prefix) {
			break
		}

		segments = append(segments, suffixIdx+suffixLen)
		sub = sub[:suffixIdx]
	}

	if len(segments) == 0 {
		releaseSegments(segments)
		return -1, nil
	}

	slices.Reverse(segments)

	return prefixIdx, segments
}

func (self PrefixSuffix) Len() int {
	return lenNo
}

func (self PrefixSuffix) Match(s string) bool {
	// the length check ensures the prefix and suffix do not overlap, so that
	// ex 'a**a' does not match "a" (which would satisfy both HasPrefix and
	// HasSuffix on the same single character).
	return len(s) >= len(self.Prefix)+len(self.Suffix) &&
		strings.HasPrefix(s, self.Prefix) && strings.HasSuffix(s, self.Suffix)
}

func (self PrefixSuffix) String() string {
	return fmt.Sprintf("<prefix_suffix:[%s,%s]>", self.Prefix, self.Suffix)
}
