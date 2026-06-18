package match

import (
	"sync"
	"testing"
)

func benchPool(b *testing.B, i int) {
	b.Helper()

	pool := sync.Pool{New: func() any {
		return make([]int, 0, i)
	}}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s := pool.Get().([]int)[:0]
			pool.Put(s)
		}
	})
}

func benchMake(b *testing.B, i int) {
	b.Helper()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = make([]int, 0, i)
		}
	})
}

func BenchmarkSegmentsPool_1(b *testing.B) {
	benchPool(b, 1)
}

func BenchmarkSegmentsPool_2(b *testing.B) {
	benchPool(b, 2)
}

func BenchmarkSegmentsPool_4(b *testing.B) {
	benchPool(b, 4)
}

func BenchmarkSegmentsPool_8(b *testing.B) {
	benchPool(b, 8)
}

func BenchmarkSegmentsPool_16(b *testing.B) {
	benchPool(b, 16)
}

func BenchmarkSegmentsPool_32(b *testing.B) {
	benchPool(b, 32)
}

func BenchmarkSegmentsPool_64(b *testing.B) {
	benchPool(b, 64)
}

func BenchmarkSegmentsPool_128(b *testing.B) {
	benchPool(b, 128)
}

func BenchmarkSegmentsPool_256(b *testing.B) {
	benchPool(b, 256)
}

func BenchmarkSegmentsMake_1(b *testing.B) {
	benchMake(b, 1)
}

func BenchmarkSegmentsMake_2(b *testing.B) {
	benchMake(b, 2)
}

func BenchmarkSegmentsMake_4(b *testing.B) {
	benchMake(b, 4)
}

func BenchmarkSegmentsMake_8(b *testing.B) {
	benchMake(b, 8)
}

func BenchmarkSegmentsMake_16(b *testing.B) {
	benchMake(b, 16)
}

func BenchmarkSegmentsMake_32(b *testing.B) {
	benchMake(b, 32)
}

func BenchmarkSegmentsMake_64(b *testing.B) {
	benchMake(b, 64)
}

func BenchmarkSegmentsMake_128(b *testing.B) {
	benchMake(b, 128)
}

func BenchmarkSegmentsMake_256(b *testing.B) {
	benchMake(b, 256)
}
