package go_frank

import (
	"testing"
)

// after removing allocation issues, 175M elements/sec
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

func BenchmarkSumInt(b *testing.B) {
	size := 1000000
	exp := size * (size - 1) / 2
	arr := givenIntArray(size)
	s := ArrayStream(arr)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Reset()
		if res, ok := s.Sum().First(); !ok || res != exp {
			b.Fatal(res, exp)
		}
	}
	b.ReportMetric(float64(size), "elems/op")
}

func BenchmarkReduce(b *testing.B) {
	size := 1000000
	exp := int64(size * (size - 1) / 2)
	arr := givenInt64Array(size)
	s := ArrayStream(arr)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var int64reducer Int64SumReducer
		s.Reset()
		if res, ok := s.ReduceNA(&int64reducer).First(); !ok || res != exp {
			b.Fatal(res)
		}
	}
	b.ReportMetric(float64(size), "elems/op")
}

func BenchmarkFilter(b *testing.B) {
	size := 1000000
	arr := givenInt64Array(size)
	s := ArrayStream(arr)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Reset()
		if res := s.Filter(func(i int64) bool { return i%2 == 0 }).Count(); res != size/2 {
			b.Fatal(res)
		}
	}
	b.ReportMetric(float64(size), "elems/op")
}

func BenchmarkFilterNA(b *testing.B) {
	size := 1000000
	arr := givenInt64Array(size)
	s := ArrayStream(arr)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Reset()
		var oddFilter OddFilterInt64
		if res := s.FilterNA(oddFilter).Count(); res != size/2 {
			b.Fatal(res)
		}
	}
	b.ReportMetric(float64(size), "elems/op")
}

func givenIntArray(count int) []interface{} {
	elems := make([]interface{}, count)
	for i := 0; i < len(elems); i++ {
		elems[i] = i
	}
	return elems
}

func givenInt64Array(count int) []interface{} {
	elems := make([]interface{}, count)
	for i := 0; i < len(elems); i++ {
		elems[i] = int64(i)
	}
	return elems
}
