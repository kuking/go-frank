package go_frank

import (
	"fmt"
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
		if res := s.SumInt64().First(); res.isEmpty() || res.Get() != exp {
			b.Fatal(res)
		}
	}
	b.StopTimer()
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
		if res := s.Sum().First(); !res.isPresent() || res.Get() != exp {
			b.Fatal(res, exp)
		}
	}
	b.ReportMetric(float64(size), "elems/op")
}

func BenchmarkMap(b *testing.B) {
	size := 1000000
	exp := int64(size*(size-1)/2) + int64(size)
	arr := givenInt64Array(size)
	s := ArrayStream(arr)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Reset()
		if res := s.Map(func(a int64) int64 { return a + 1 }).SumInt64().First(); res.isEmpty() || res.Get() != exp {
			b.Fatal(res, exp)
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(size), "elems/op")
}

func BenchmarkMapInt64(b *testing.B) {
	size := 1000000
	exp := int64(size*(size-1)/2) + int64(size)
	arr := givenInt64Array(size)
	s := ArrayStream(arr)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Reset()
		if res := s.MapInt64(func(a int64) int64 { return a + 1 }).SumInt64().First(); res.isEmpty() || res.Get() != exp {
			b.Fatal(res, exp)
		}
	}
	b.StopTimer()
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
		s.Reset()
		if res := s.Reduce(func(a, b int64) int64 { return a + b }).First(); res.isEmpty() || res.Get() != exp {
			b.Fatal(res)
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(size), "elems/op")
}

// X40 times faster
func BenchmarkReduceNA(b *testing.B) {
	size := 1000000
	exp := int64(size * (size - 1) / 2)
	arr := givenInt64Array(size)
	s := ArrayStream(arr)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var int64reducer Int64SumReducer
		s.Reset()
		if res := s.ReduceNA(&int64reducer).First(); res.isEmpty() || res.Get() != exp {
			b.Fatal(res)
		}
	}
	b.StopTimer()
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
	b.StopTimer()
	b.ReportMetric(float64(size), "elems/op")
}

// X24 times faster
func BenchmarkFilterNA(b *testing.B) {
	size := 1000000
	arr := givenInt64Array(size)
	s := ArrayStream(arr)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Reset()
		if res := s.FilterNA(func(i interface{}) bool { return i.(int64)%2 == 0 }).Count(); res != size/2 {
			b.Fatal(res)
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(size), "elems/op")
}

func BenchmarkGeneratorSumInt64(b *testing.B) {
	total := b.N
	exp := int64(total * (total - 1) / 2)
	s := givenInt64StreamGenerator(total)
	b.ReportAllocs()
	b.ResetTimer()
	if res := s.SumInt64().First(); res.isEmpty() && res.Get() != exp {
		b.Fatal(res)
	}
	b.StopTimer()
}

func BenchmarkGeneratorFilterNA(b *testing.B) {
	total := b.N
	s := givenInt64StreamGenerator(total)
	b.ReportAllocs()
	b.ResetTimer()
	if res := s.FilterNA(func(i interface{}) bool { return i.(int64)%2 == 0 }).Count(); res != (total+1)/2 {
		b.Fatal(fmt.Sprintf("res %v != %v exp ; total=%v", res, (total+1)/2, total))
	}
	b.StopTimer()
}

func BenchmarkGeneratorCounter(b *testing.B) {
	total := b.N
	s := givenInt64StreamGenerator(total)
	b.ReportAllocs()
	b.ResetTimer()
	if res := s.Count(); res != total {
		b.Fatal()
	}
	b.StopTimer()
}

// -------------------------------------------------------------------------------------------------------------------

func givenInt64StreamGenerator(total int) Stream {
	count := int64(0)
	return StreamGenerator(func() Optional {
		count++
		if count <= int64(total) {
			return OptionalOf(count)
		}
		return EmptyOptional()
	})
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
