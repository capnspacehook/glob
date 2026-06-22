package match

import (
	"reflect"
	"testing"
)

func TestPrefixSuffixIndex(t *testing.T) {
	for id, test := range []struct {
		prefix   string
		suffix   string
		fixture  string
		index    int
		segments []int
	}{
		{
			"a",
			"c",
			"abc",
			0,
			[]int{3},
		},
		{
			// the single leading "f" is not a valid match: the prefix and
			// suffix would overlap on the same character, so no segment of
			// length 1 is reported.
			"f",
			"f",
			"fffabfff",
			0,
			[]int{2, 3, 6, 7, 8},
		},
		{
			// "abc" cannot match prefix "ab" and suffix "bc": they would
			// overlap on 'b', so there is no match at all.
			"ab",
			"bc",
			"abc",
			-1,
			nil,
		},
		{
			// non-overlapping occurrence: "abbc" matches.
			"ab",
			"bc",
			"abbc",
			0,
			[]int{4},
		},
	} {
		p := NewPrefixSuffix(test.prefix, test.suffix)
		index, segments := p.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}

func TestPrefixSuffixMatch(t *testing.T) {
	for id, test := range []struct {
		prefix string
		suffix string
		s      string
		match  bool
	}{
		// prefix and suffix must not overlap: "a" cannot satisfy both a
		// one-char prefix and a one-char suffix simultaneously.
		{"a", "a", "a", false},
		{"a", "a", "aa", true},
		{"a", "a", "aba", true},
		{"ab", "yz", "abyz", true},
		{"ab", "yz", "abz", false},
		{"ab", "yz", "abyabyz", true},
		{"", "", "", true},
		{"a", "", "a", true},
		{"", "a", "a", true},
	} {
		p := NewPrefixSuffix(test.prefix, test.suffix)
		if got := p.Match(test.s); got != test.match {
			t.Errorf("#%d PrefixSuffix(%q,%q).Match(%q) = %v, want %v",
				id, test.prefix, test.suffix, test.s, got, test.match)
		}
	}
}

func BenchmarkIndexPrefixSuffix(b *testing.B) {
	m := NewPrefixSuffix("qew", "sqw")

	for b.Loop() {
		_, s := m.Index(bench_pattern)
		releaseSegments(s)
	}
}

func BenchmarkIndexPrefixSuffixParallel(b *testing.B) {
	m := NewPrefixSuffix("qew", "sqw")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, s := m.Index(bench_pattern)
			releaseSegments(s)
		}
	})
}
