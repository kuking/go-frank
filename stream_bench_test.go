package go_frank

import (
	"testing"
)

func BenchmarkSum(b *testing.B) {
	size := 1000000
	exp := size * (size - 1) / 2
	arr := givenArray(size)
	s := ArrayStream(arr)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Reset()
		if res, ok := s.Sum().First(); !ok || res != exp {
			b.Fatal(res)
		}
	}
	b.ReportMetric(float64(size), "elems/op")
}

func givenArray(count int) []interface{} {
	elems := make([]interface{}, count)
	for i := 0; i < len(elems); i++ {
		elems[i] = i
	}
	return elems
}
