package go_frank

import (
	"testing"
)

// after removing allocation issues, 175M elements/sec (
func BenchmarkSumInt64(b *testing.B) {
	size := 1000000
	exp := int64(size * (size - 1) / 2)
	arr := givenInt64Array(size)
	s := ArrayStream(arr)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Reset()
		if res, ok := s.SumInt64().First(); !ok || res != exp {
			b.Fatal(res)
		}
	}
	b.ReportMetric(float64(size), "elems/op")
}

// after removing allocation issues, 175M elements/sec (
func BenchmarkReduce(b *testing.B) {
	size := 1000000
	exp := int64(size * (size - 1) / 2)
	arr := givenInt64Array(size)
	s := ArrayStream(arr)
	b.ReportAllocs()
	b.ResetTimer()

	var int64reducer Int64SumReducer
	for i := 0; i < b.N; i++ {
		s.Reset()
		if res, ok := s.ReduceNA(&int64reducer).First(); !ok || res != exp {
			b.Fatal(res)
		}
	}
	b.ReportMetric(float64(size), "elems/op")
}

func givenInt64Array(count int) []interface{} {
	elems := make([]interface{}, count)
	for i := 0; i < len(elems); i++ {
		elems[i] = int64(i)
	}
	return elems
}
