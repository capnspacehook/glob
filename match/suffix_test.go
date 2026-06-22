package match

import (
	"reflect"
	"testing"
)

func TestSuffixIndex(t *testing.T) {
	for id, test := range []struct {
		prefix   string
		fixture  string
		index    int
		segments []int
	}{
		{
			"ab",
			"abc",
			0,
			[]int{2},
		},
		{
			"ab",
			"fffabfff",
			0,
			[]int{5},
		},
		{
			// '**a' matches at every occurrence of the suffix, so each one is
			// a valid match length (separators included, since '**' crosses
			// them).
			"a",
			"aba",
			0,
			[]int{1, 3},
		},
	} {
		p := NewSuffix(test.prefix)
		index, segments := p.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}

func BenchmarkIndexSuffix(b *testing.B) {
	m := NewSuffix("qwe")

	for b.Loop() {
		_, s := m.Index(bench_pattern)
		releaseSegments(s)
	}
}

func BenchmarkIndexSuffixParallel(b *testing.B) {
	m := NewSuffix("qwe")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, s := m.Index(bench_pattern)
			releaseSegments(s)
		}
	})
}
