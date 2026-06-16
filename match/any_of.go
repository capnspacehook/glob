package match

import "fmt"

// AnyOf matches if any of its sub-matchers match; ex '{a,b,c}'.
type AnyOf struct {
	Matchers Matchers
}

func NewAnyOf(m ...Matcher) AnyOf {
	return AnyOf{Matchers(m)}
}

func (self *AnyOf) Add(m Matcher) {
	self.Matchers = append(self.Matchers, m)
}

func (self AnyOf) Match(s string) bool {
	for _, m := range self.Matchers {
		if m.Match(s) {
			return true
		}
	}

	return false
}

func (self AnyOf) Index(s string) (int, []int) {
	index := -1

	segments := acquireSegments(len(s))
	for _, m := range self.Matchers {
		idx, seg := m.Index(s)
		if idx == -1 {
			continue
		}

		if index == -1 || idx < index {
			index = idx
			segments = append(segments[:0], seg...)
			continue
		}

		if idx > index {
			continue
		}

		// here idx == index
		segments = appendMerge(segments, seg)
	}

	if index == -1 {
		releaseSegments(segments)
		return -1, nil
	}

	return index, segments
}

func (self AnyOf) Len() int {
	var set bool
	var l int
	for _, m := range self.Matchers {
		ml := m.Len()
		if ml == -1 {
			return -1
		}

		if !set {
			l = ml
			set = true
			continue
		}
		if l != ml {
			return -1
		}
	}

	return l
}

func (self AnyOf) String() string {
	return fmt.Sprintf("<any_of:[%s]>", self.Matchers)
}
